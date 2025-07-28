package commands

import (
	"filevault/services"
	"fmt"
)

type UploadCommand struct {
	fileService *services.FileService
}

// NewUploadCommand creates a new instance of UploadCommand.
func NewUploadCommand(fileService *services.FileService) ICommand {
	return &UploadCommand{
		fileService: fileService,
	}
}

// Execute runs the upload command with the provided arguments.
func (c *UploadCommand) Execute(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: upload <file_path> [folder_name]")
	}
	filePath := args[0]
	var folderName string
	if len(args) > 1 {
		folderName = args[1]
	}

	fmt.Printf("Uploading file: %s", filePath)
	if folderName != "" {
		fmt.Printf(" to folder: %s", folderName)
	}
	fmt.Println("...")

	var err error
	if folderName == "" {
		// Upload to root directory
		err = c.fileService.UploadFile(filePath, "")
	} else {
		// Upload to specific folder
		err = c.fileService.UploadFileToFolder(filePath, folderName)
	}

	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	fmt.Printf("Uploaded file: %s\n", filePath)
	return nil
}

func (c *UploadCommand) Name() string {
	return "upload"
}

func (c *UploadCommand) HelpContent() string {
	return "upload <file_path> [folder_name] - Upload a file to a folder"
}
