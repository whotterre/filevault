package services

import (
	"context"
	"encoding/json"
	"errors"
	"filevault/repositories"
	"filevault/utils"
	worker "filevault/workers"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Constants
const (
	DEFAULT_FOLDER_NAME = "default"
	UPLOAD_FOLDER_PATH  = "./storage/uploads"
)

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
	ErrNoFolderNamePassed  = errors.New("Folder name wasn't passed while trying to create new folder")
	ErrMissingFileId       = errors.New("File id wasn't passed")
	ErrNotAuthenticated    = errors.New("User not authenticated")
)

type FileService struct {
	conn            *redis.Client
	authRepo        repositories.UserRepository
	fileRepo        repositories.FileRepository
	taskDistributor worker.TaskDistributor
}

type FileMetadata struct {
	FileId     string    `json:"file_id"` // UUID
	FileName   string    `json:"file_name"`
	Size       int64     `json:"size"`        // In bytes
	Path       string    `json:"path"`        // ./uploads/notes.txt"
	UploadedAt time.Time `json:"uploaded_at"` // Iykyk
	FileType   string    `json:"file_type"`
	ParentId   string    `json:"parent_id"`
}

func NewFileService(conn *redis.Client,
	fileRepo repositories.FileRepository,
	taskDistributor worker.TaskDistributor,
) *FileService {
	return &FileService{
		conn:            conn,
		fileRepo:        fileRepo,
		taskDistributor: taskDistributor,
	}
}

// This classifies the file into one of three classes
// Generic file, folder, or image
func (s *FileService) determineFileType(filePath string) (string, error) {
	// Check if it's an image
	matched, err := regexp.MatchString(`(?i)\.(jpg|tiff|gif|bmp|png)$`, filePath)
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

	return "file", nil // Generic file case
}

// Gets the user ID from Redis using the session token
func (s *FileService) getUserID() (string, error) {
	ctx := context.Background()
	currentUser, err := utils.GetCurrentUser()
	if err != nil {
		return "", fmt.Errorf("Failed to get current user email %v", err)
	}

	// Query the database to get the user ID
	id, err := s.fileRepo.GetUserByEmail(ctx, currentUser)
	if err != nil {
		return "", fmt.Errorf("Error querying user ID: %v", err)
	}
	if id == "" {
		return "", fmt.Errorf("User ID not found for email: %s", currentUser)
	}

	return id, nil
}

func (s *FileService) checkIsAuthenticated(sessionToken string, conn *redis.Client) (bool, error) {
	ctx := context.Background()
	// Check if the user has a session token
	email, err := conn.Get(ctx, sessionToken).Result()
	if err != nil {
		return false, errors.New("User isn't authenticated as they don't have a session token")
	}

	// Get user record from the SQLite3 db
	id, err := s.fileRepo.GetUserByEmail(ctx, email)
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
func (s *FileService) UploadFile(pathname, parentID string) error {
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
	if _, err := os.Stat(UPLOAD_FOLDER_PATH); os.IsNotExist(err) {
		err = os.MkdirAll(UPLOAD_FOLDER_PATH, os.ModePerm) // So modeperm is like the 777 in chmod 777
		if err != nil {
			return err
		}
	}
	// Check the file type
	fileType, err := s.determineFileType(pathname)
	if err != nil {
		return err
	}


	// Copy the file FIRST
	uploadedFile, err := os.Open(pathname)
	if err != nil {
		return err
	}
	defer uploadedFile.Close()

	destinationPath := UPLOAD_FOLDER_PATH + "/" + osStat.Name() // Fix path separator
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// Copy the content of the old file to the new file
	_, err = io.Copy(destinationFile, uploadedFile)
	if err != nil {
		return err
	}

	// Get user authentication details
	sessionToken, err := utils.GetSessionTokenFromFile()
	if err != nil {
		return fmt.Errorf("failed to get session token: %w", err)
	}

	userId, err := s.getUserID()
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	valid, err := s.checkIsAuthenticated(sessionToken, s.conn)
	if err != nil {
		return fmt.Errorf("Failed to validate user because %w", err)
	}
	if !valid {
		return fmt.Errorf("User isn't authenticated")
	}

	// Store the metadata of the file
	fileMetadata := FileMetadata{
		FileId:     uuid.New().String(),
		FileName:   osStat.Name(),
		Size:       osStat.Size(),
		Path:       destinationPath,
		UploadedAt: time.Now(),
		FileType:   fileType,
	}

	// Add database record of metadata
	err = s.fileRepo.CreateFile(fileMetadata.FileId, fileMetadata.FileName,
		userId, fileMetadata.Path, fileMetadata.FileType, fileMetadata.Size, "",
		fileMetadata.UploadedAt)
	if err != nil {
		return fmt.Errorf("failed to execute database statement: %w", err)
	}

	// NOW queue thumbnail generation AFTER the file is copied and saved
	if fileType == "image" {
		// This is NOT async: worker.GenerateThumbnail runs synchronously.
		// If you want async, use s.taskDistributor.DistributeThumbnailGeneration (uncomment below).
		err := worker.GenerateThumbnail(destinationPath, fileMetadata.FileId)
		if err != nil {
			// Don't fail the upload if thumbnail queueing fails
			fmt.Printf("Warning: Failed to queue thumbnail generation: %v\n", err)
		} else {
			fmt.Println("✓ Thumbnail generation queued")
		}
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
		metadataList = make([]FileMetadata, 0, 1) // Pre-allocate with capacity for at least one element
	} else {
		var rawMetadata any
		err = json.Unmarshal(fileContent, &rawMetadata)
		if err != nil {
			return err
		}

		switch data := rawMetadata.(type) {
		case []any:
			metadataList = make([]FileMetadata, 0, len(data))
			err = json.Unmarshal(fileContent, &metadataList)
			if err != nil {
				return err
			}
		case map[string]any:
			var singleMetadata FileMetadata
			err = json.Unmarshal(fileContent, &singleMetadata)
			if err != nil {
				return err
			}
			metadataList = make([]FileMetadata, 0, 1) // Pre-allocate with capacity for one element
			metadataList = append(metadataList, singleMetadata)
		default:
			return fmt.Errorf("unexpected metadata format")
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

func (s *FileService) createFolderWithFiles(foldername, sourcePath string) error {
	// Validate and sanitize foldername
	validFolderName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validFolderName.MatchString(foldername) {
		return fmt.Errorf("invalid folder name: %s", foldername)
	}

	// Create the folder with the sanitized name
	err := os.Mkdir(foldername, os.ModePerm)
	if err != nil {
		return err
	}

	// Recursively read all contents of the directory
	files, err := os.ReadDir(sourcePath)
	if err != nil {
		return err
	}

	// Copy files to the created folder
	for _, file := range files {
		sourceFilePath := sourcePath + "/" + file.Name()
		destinationFilePath := foldername + "/" + file.Name()

		// Copy each file from the source directory to the destination directory
		sourceFile, err := os.Open(sourceFilePath)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		destinationFile, err := os.Create(destinationFilePath)
		if err != nil {
			return err
		}
		defer destinationFile.Close()

		_, err = io.Copy(destinationFile, sourceFile)
		if err != nil {
			return err
		}
	}

	// Generate metadata for the folder
	folderMetadata := FileMetadata{
		FileId:     uuid.New().String(),
		FileName:   foldername,
		Size:       0, // Folders don't have a size in this context
		Path:       foldername,
		UploadedAt: time.Now(),
		FileType:   "folder",
	}

	// Read the existing metadata from metadata.json
	metadataPath := "./storage/metadata.json"
	fileContent, err := os.ReadFile(metadataPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var metadataList []FileMetadata
	if len(fileContent) > 0 {
		err = json.Unmarshal(fileContent, &metadataList)
		if err != nil {
			return err
		}
	}

	// Append the new folder metadata
	metadataList = append(metadataList, folderMetadata)

	// Write the updated metadata back to metadata.json
	updatedMetadata, err := json.Marshal(metadataList)
	if err != nil {
		return err
	}

	err = os.WriteFile(metadataPath, updatedMetadata, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Folder '%s' created successfully and metadata stored.\n", foldername)
	return nil
}

func (s *FileService) ListUploaded() error {
	// Check if the metadata.json file exists
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
	err := s.fileRepo.DeleteFile(fileId)
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

// Creates an empty folder
func (s *FileService) CreateEmptyFolder(folderName, parentName string) error {
	if folderName == "" {
		return ErrNoFolderNamePassed
	}

	// Ensure the storage/uploads directory exists
	if _, err := os.Stat(UPLOAD_FOLDER_PATH); os.IsNotExist(err) {
		err = os.MkdirAll(UPLOAD_FOLDER_PATH, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create uploads directory: %w", err)
		}
	}

	// Create the subdirectory with the given name inside uploads
	folderPath := UPLOAD_FOLDER_PATH + "/" + folderName
	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	// Get user ID by session token
	userId, err := s.getUserID()
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	// Make folder database entry
	folderID := uuid.New().String()
	err = s.fileRepo.CreateFile(folderID, folderName, userId, folderPath, "folder", 0, parentName, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create database entry for new folder: %w", err)
	}

	// If parentId is provided, create a nested folder inside the parent folder
	if parentName != "" {
		ctx := context.Background()
		parentInfo, err := s.fileRepo.GetFolderByName(ctx, parentName)
		if err != nil {
			return fmt.Errorf("failed to get parent folder info: %w", err)
		}
		if parentInfo.Path == "" {
			return fmt.Errorf("parent folder with name %s does not exist", parentName)
		}

		nestedFolderPath := parentInfo.Path + "/" + folderName
		if err := os.MkdirAll(nestedFolderPath, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create nested folder: %w", err)
		}

		nestedFolderID := uuid.New().String()
		err = s.fileRepo.CreateFile(nestedFolderID, folderName, userId, nestedFolderPath, "folder", 0, parentInfo.FileId, time.Now())
		if err != nil {
			return fmt.Errorf("failed to create database entry for nested folder: %w", err)
		}
	}

	return nil
}

//

func (s *FileService) PublishFile(fileId string) error {
	ctx := context.Background()
	if fileId == "" {
		return ErrMissingFileId
	}

	// Ensure user is authenticated
	sessionToken, err := utils.GetSessionTokenFromFile()
	if err != nil {
		return fmt.Errorf("failed to get session token: %w", err)
	}
	isAuth, err := s.checkIsAuthenticated(sessionToken, s.conn)
	if err != nil {
		return fmt.Errorf("Something went wrong while checking whether user was authenticated: %w", err)
	}

	if !isAuth {
		return ErrNotAuthenticated
	}

	// We need to check that the user id is the same as the one for the file in the db
	userId, err := s.getUserID()
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}
	fileOwnerId, err := s.fileRepo.GetFileOwnerId(ctx, fileId)
	if err != nil {
		return fmt.Errorf("failed to get file owner ID: %w", err)
	}
	if userId != fileOwnerId {
		return errors.New("user is not authorized to make file public")
	}
	err = s.fileRepo.PublishFile(ctx, fileId)
	if err != nil {
		return fmt.Errorf("failed to make file public: %w", err)
	}

	fmt.Printf("File with ID %s has successfully been made public\n", fileId)
	return nil
}

func (s *FileService) UnPublishFile(fileId string) error {
	ctx := context.Background()
	if fileId == "" {
		return ErrMissingFileId
	}

	// Ensure user is authenticated
	sessionToken, err := utils.GetSessionTokenFromFile()
	if err != nil {
		return fmt.Errorf("failed to get session token: %w", err)
	}
	isAuth, err := s.checkIsAuthenticated(sessionToken, s.conn)
	if err != nil {
		return fmt.Errorf("Something went wrong while checking whether user was authenticated: %w", err)
	}

	if !isAuth {
		return ErrNotAuthenticated
	}

	// We need to check that the user id is the same as the one for the file in the db
	userId, err := s.getUserID()
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}
	fileOwnerId, err := s.fileRepo.GetFileOwnerId(ctx, fileId)
	if err != nil {
		return fmt.Errorf("failed to get file owner ID: %w", err)
	}
	if userId != fileOwnerId {
		return errors.New("user is not authorized to make file private")
	}
	err = s.fileRepo.PublishFile(ctx, fileId)
	if err != nil {
		return fmt.Errorf("failed to make file private: %w", err)
	}

	fmt.Printf("File with ID %s has successfully been made private\n", fileId)
	return nil
}

func (s *FileService) UploadFileToFolder(filePath string, folderName string) error {
	// Ensure user is logged in
	if !utils.ValidateUser(s.conn) {
		return errors.New("user is not logged in")
	}

	if filePath == "" {
		return ErrMissingPathname
	}

	// Check if the file exists and is not a directory
	osStat, err := os.Stat(filePath)
	if err != nil || osStat.IsDir() {
		return ErrInvalidFileFormat
	}

	// Ensure the folder exists inside UPLOAD_FOLDER_PATH
	folderPath := UPLOAD_FOLDER_PATH + "/" + folderName
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err = os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			return ErrCreatingUploadDir
		}
	}

	// Copy the file to the folder
	uploadedFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer uploadedFile.Close()

	destinationPath := folderPath + "/" + osStat.Name()
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, uploadedFile)
	if err != nil {
		return err
	}
	userId, err := s.getUserID()
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	// Get file type
	fileType, err := s.determineFileType(filePath)
	if err != nil {
		return err
	}

	// Store metadata in DB
	fileMetadata := FileMetadata{
		FileId:     uuid.New().String(),
		FileName:   osStat.Name(),
		Size:       osStat.Size(),
		Path:       destinationPath,
		UploadedAt: time.Now(),
		FileType:   fileType,
		ParentId:   folderName,
	}
	err = s.fileRepo.CreateFile(fileMetadata.FileId, fileMetadata.FileName,
		userId, fileMetadata.Path, fileMetadata.FileType, fileMetadata.Size, folderName,
		fileMetadata.UploadedAt)
	if err != nil {
		return fmt.Errorf("failed to execute database statement: %w", err)
	}

	// Update metadata.json
	metadataPath := "./storage/metadata.json"
	fileContent, err := os.ReadFile(metadataPath)
	var metadataList []FileMetadata
	if err == nil && len(fileContent) > 0 {
		_ = json.Unmarshal(fileContent, &metadataList)
	}
	metadataList = append(metadataList, fileMetadata)
	updatedMetadata, err := json.Marshal(metadataList)
	if err != nil {
		return err
	}
	err = os.WriteFile(metadataPath, updatedMetadata, 0644)
	if err != nil {
		return err
	}

	return nil
}