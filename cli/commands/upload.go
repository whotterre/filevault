package commands

import (
	"filevault/services"
	"fmt"
)

type UploadCommand struct {
	fileService *services.FileService
}

// NewUploadCommand creates a new instance of UploadCommand.
func NewUploadCommand(fileService *services.FileService) *UploadCommand {
	return &UploadCommand{
		fileService: fileService,
	}
}

// Execute runs the upload command with the provided arguments.
func (c *UploadCommand) Execute(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no file path provided")
	}
	filePath := args[0]
	fmt.Printf("Uploading file: %s\n...", filePath)
	err := c.fileService.UploadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	fmt.Printf("Uploaded file: %s\n...", filePath)
	return nil
}

func (c *UploadCommand) Name() string {
	return "upload"
}

func (c *UploadCommand) HelpContent() string {
    return "Upload a file to the vault. Usage: upload <filepath>"
}