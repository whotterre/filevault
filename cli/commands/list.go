package commands

import (
	"errors"
	"filevault/services"
	"fmt"
)

type ListCommand struct {
	fileService *services.FileService
}

func NewListCommand(fileService *services.FileService) *ListCommand {
	return &ListCommand{
		fileService: fileService,
	}
}

func (c *ListCommand) Execute(args []string) error {
	if len(args) > 0 {
		return errors.New("list command doesn't require any arguments")
	}

	fmt.Println("Listing file metadata")
	err := c.fileService.ListUploaded()
	if err != nil {
		return err
	}

	return nil

}

func (c *ListCommand) Name() string {
	return "list"
}

func (c *ListCommand) HelpContent() string {
	return "Lists all uploaded files with basic metadata."
}
