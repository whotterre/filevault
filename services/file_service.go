package services

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
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
)

type FileService struct {
}

type FileMetadata struct {
	FileId     string    `json:"file_id"` // UUID
	FileName   string    `json:"file_name"`
	Size       int64     `json:"size"`        // In bytes
	Path       string    `json:"path"`        // ./uploads/notes.txt"
	UploadedAt time.Time `json:"uploaded_at"` // Iykyk

}

func NewFileService() *FileService {
	return &FileService{}
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

	// Enough shalaye, let's upload the file!
	uploadedFile, err := os.Open(pathname)
	if err != nil {
		return err
	}
	defer uploadedFile.Close() // Close the file to preserve system resources
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

	// Store the metadata of the file in metadata.json
	fileMetadata := FileMetadata{
		FileId:     uuid.New().String(),
		FileName:   osStat.Name(),
		Size:       osStat.Size(),
		Path:       destinationPath,
		UploadedAt: time.Now(),
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
