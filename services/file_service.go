package services

import (
	"context"
	"encoding/json"
	"errors"
	"filevault/repositories"
	"filevault/utils"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Constants
const DEFAULT_FOLDER_NAME = "default"
// Errors
var (
	ErrMissingPathname     = errors.New("Missing pathname")
	ErrInvalidFileFormat   = errors.New("You either passed an invalid file or a directory")
	ErrCreatingUploadDir   = errors.New("Failed to create upload directory")
	ErrInvalidFile         = errors.New("Failed to create file")
	ErrDatabaseAccessFail  = errors.New("Failed to open database file")
	ErrReadingFileContent  = errors.New("Failed to read file content")
	ErrDatabaseWriteFail   = errors.New("Failed to write file content")
	ErrMetadataJSONMarshal = errors.New("Failed to convert metadata to JSON")
	ErrJSONUnmarshal       = errors.New("Failed to unmarshal data")
	ErrFileUpload          = errors.New("Failed to upload file to filesystem")
	ErrFileNotExistent     = errors.New("File doesn't exist")
)

type FileService struct {
	conn *redis.Client
	repo repositories.FileRepository
}

type FileMetadata struct {
	FileId     string    `json:"file_id"` // UUID
	FileName   string    `json:"file_name"`
	Size       int64     `json:"size"`        // In bytes
	Path       string    `json:"path"`        // ./uploads/notes.txt"
	UploadedAt time.Time `json:"uploaded_at"` // Iykyk
	FileType   string 	 `json:"file_type"`
	ParentId   string    `json:"parent_id"`  

}

func NewFileService(conn *redis.Client, repo repositories.FileRepository) *FileService {
	return &FileService{
		conn: conn,
		repo: repo,
	}
}
func (s *FileService) determineFileType(filePath string) (string, error) {
	// This classifies the file into one of three classes
	// Generic file, folder, or image

	// Check if it's an image
	matched, err:= regexp.MatchString(`(?i)\.(jpg|tiff|gif|bmp|png)$`, filePath)
	if matched {
		return "image", nil
	}
	if err != nil {
		return "", fmt.Errorf("couldn't apply regex on input because %w", err)
	}
	// Check if it's a directory
	fileInfo, err := os.Stat(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("file does not exist: %w", err)
	}
	if fileInfo.IsDir() {
		return "folder", nil  
	}
	
	return "file", nil 		// Generic file case
}

// Gets the user ID from Redis using the session token
func (s *FileService) getUserID(sessionToken string, conn *redis.Client) (string, error) {
	ctx := context.Background()
	email, err := conn.Get(ctx, sessionToken).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("Session token does not exist")
		}
		return "", fmt.Errorf("Error getting user ID: %v", err)
	}

	// Query the database to get the user ID
	var id string
	err = s.repo.GetUserByEmail(ctx, email, &id)
	if err != nil {
		return "", fmt.Errorf("Error querying user ID: %v", err)
	}
	if id == "" {
		return "", fmt.Errorf("User ID not found for email: %s", email)
	}

	return id, nil
}

func (s *FileService) checkIsAuthenticated(sessionToken, userID string, conn *redis.Client) (bool, error) {
	ctx := context.Background()
	// Check if the user has a session token
	email, err := conn.Get(ctx, sessionToken).Result()
	if err != nil {
		return false, errors.New("User isn't authenticated as they don't have a session token")
	}

	// Get user record from the SQLite3 db
	var id string
	err = s.repo.GetUserByEmail(ctx, email, &id)
	if err != nil {
		return false, errors.New("Something went wrong in trying to get user by email for auth")
	}

	// User exists if a record is found
	if id == "" {
		return false, errors.New("User ID not found for email: " + email)
	}
	fmt.Println("User is authenticated with ID:", id)
	return true, nil
}

// UploadFiles uploads files to the server.
// It checks if the "uploads" directory exists in the storage subdirectory.
// Parameters:
//   - pathname: The path of the file to be uploaded.
func (s *FileService) UploadFile(pathname string) error {
	// Check if the "uploads" directory exists in the storage subdirectory
	// If it doesn't exist, create it.
	// Then extract the file metadata, generate UUID for file and then upload the file
	// Returning it's UUID

	// Ensure user is logged in 
	if !utils.ValidateUser(s.conn) {
		return errors.New("user is not logged in")
	}

	if pathname == "" {
		return ErrMissingPathname
	}

	// Check if the file exists
	fileExists := true
	osStat, err := os.Stat(pathname)
	if err != nil {
		fileExists = false
	}
	if !fileExists || osStat.IsDir() {
		return ErrInvalidFileFormat
	}

	// Check that the uploads folder exists
	uploadPath := "./storage/uploads"
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		err = os.MkdirAll(uploadPath, os.ModePerm) // So modeperm is like the 777 in chmod 777
		if err != nil {
			return err
		}
	}
	// Check the file type 
	fileType, err := s.determineFileType(pathname)
	if err != nil {
		return err
	}
	// Enough shalaye, let's upload the file!
	uploadedFile, err := os.Open(pathname)
	if err != nil {
		return err
	}
	defer uploadedFile.Close()
	destinationPath := uploadPath + "/" + osStat.Name()
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}

	// Copy the content of the old file to the new file
	_, err = io.Copy(destinationFile, uploadedFile)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// Get user ID from the key value pair [sessionToken -> userId]
	sessionToken, err := utils.GetSessionTokenFromFile()
	if err != nil {
		return fmt.Errorf("failed to get session token: %w", err)
	}
	// Get user ID by session token
	userId, err := s.getUserID(sessionToken, s.conn)
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	valid, err := s.checkIsAuthenticated(sessionToken, userId, s.conn)
	if err != nil {
		return fmt.Errorf("Failed to validate user because %w", err)
	}
	if !valid {
		return fmt.Errorf("User isn't authenticated")
	}

	// Store the metadata of the file in metadata.json
	fileMetadata := FileMetadata{
		FileId:     uuid.New().String(),
		FileName:   osStat.Name(),
		Size:       osStat.Size(),
		Path:       destinationPath,
		UploadedAt: time.Now(),
		FileType:   fileType,
	}
	// Add database record of metadata
	err = s.repo.CreateFile(fileMetadata.FileId, fileMetadata.FileName,
		 userId, fileMetadata.Path, fileMetadata.FileType, fileMetadata.Size,
		 fileMetadata.UploadedAt)
	if err != nil {
		return fmt.Errorf("failed to execute database statement: %w", err)
	}

	// Open metadata.json
	databaseFile, err := os.Open("./storage/metadata.json")
	if err != nil {
		return err
	}
	defer databaseFile.Close()

	// Read the file content
	fileContent, err := os.ReadFile("./storage/metadata.json")
	if err != nil {
		return err
	}
	var metadataList []FileMetadata
	// Handle empty file or initialize empty array
	if len(fileContent) == 0 || string(fileContent) == "{}" || string(fileContent) == "" {
		metadataList = []FileMetadata{}
	} else {
		err = json.Unmarshal(fileContent, &metadataList)
		if err != nil {
			var singleMetadata FileMetadata
			err = json.Unmarshal(fileContent, &singleMetadata)
			if err != nil {
				return err
			}
			metadataList = []FileMetadata{singleMetadata}
		}
	}
	// Update the metadata list with the new file metadata
	metadataList = append(metadataList, fileMetadata)
	updatedMetadata, err := json.Marshal(metadataList)
	if err != nil {
		return err
	}

	// Write to the database
	err = os.WriteFile("./storage/metadata.json", updatedMetadata, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (s *FileService) ListUploaded() error {
	// Check if the metadata.json file

	// Ensure user is logged in 
	if !utils.ValidateUser(s.conn) {
		return errors.New("user is not logged in")
	}

	const pathName = "./storage/metadata.json"
	fileExists := true
	osStat, err := os.Stat(pathName)
	if err != nil {
		fileExists = false
	}
	if !fileExists || osStat.IsDir() {
		return ErrInvalidFileFormat
	}

	// Read the file content
	fileContent, err := os.ReadFile("./storage/metadata.json")
	if err != nil {
		return err
	}
	// File metadata
	fileMetadata := []FileMetadata{}

	// Unmarshal/Parse the JSON
	err = json.Unmarshal(fileContent, &fileMetadata)
	if err != nil {
		return err
	}
	// Print table header
	fmt.Println("ID                                   | Name                   | Size      | Uploaded At")
	fmt.Println("-------------------------------------+------------------------+-----------+--------------------")
	if len(fileMetadata) == 0 {
		fmt.Println("No entries found in metadata.json")
		return nil
	}
	for _, entry := range fileMetadata {
		fmt.Printf("%-36s | %-22s | %-9s | %s\n",
			entry.FileId,
			entry.FileName,
			utils.GetSizeField(entry.Size),
			entry.UploadedAt,
		)
	}
	return nil
}

func (s *FileService) DeleteFile(fileId string) error {
	// Ensure user is logged in 
	if !utils.ValidateUser(s.conn) {
		return errors.New("user is not logged in")
	}

	// Ensure the fileId was passed
	if fileId == "" {
		fmt.Print("Error fileID wasn't passed")
		return errors.New("file ID is missing")
	}

	// Delete records of it from the database
	err := s.repo.DeleteFile(fileId)
	if err != nil {
		fmt.Print("Failed to delete file")
		return errors.New("file ID is missing ")
	}
	// Check if the metadata.json file exists
	metadataPath := "./storage/metadata.json"
	fileExists := true
	osStat, err := os.Stat(metadataPath)
	if err != nil {
		fileExists = false
	}
	if !fileExists || osStat.IsDir() {
		return ErrInvalidFileFormat
	}

	// Read the file content
	fileContent, err := os.ReadFile(metadataPath)
	if err != nil {
		return err
	}
	// File metadata
	fileMetadata := []FileMetadata{}

	// Unmarshal/Parse the JSON
	err = json.Unmarshal(fileContent, &fileMetadata)
	if err != nil {
		return err
	}
	// Read the metadata file and check if there exists an entry with the fileId
	for _, x := range fileMetadata {
		if x.FileId == fileId {
			// If the file exists, delete it from the filesystem
			if _, err := os.Stat(x.Path); err == nil {
				err = os.Remove(x.Path)
				if err != nil {
					return ErrFileUpload
				}
			} else {
				return ErrFileNotExistent
			}

			// Remove the entry from the metadata list
			for i, entry := range fileMetadata {
				if entry.FileId == fileId {
					fileMetadata = append(fileMetadata[:i], fileMetadata[i+1:]...)
					break
				}
			}

			// Write the updated metadata back to the file
			updatedMetadata, err := json.Marshal(fileMetadata)
			if err != nil {
				return ErrMetadataJSONMarshal
			}
			err = os.WriteFile("./storage/metadata.json", updatedMetadata, 0644)
			if err != nil {
				return ErrDatabaseWriteFail
			}
			fmt.Printf("File with ID %s has been deleted successfully\n", fileId)
			return nil
		}
	}

	return nil
}
