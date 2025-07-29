package commands

import (
	"core/services"
	"errors"
	"fmt"
)

type LsCommand struct {
	fileService *services.FileService
}

func NewLsCommand(fileService *services.FileService) ICommand {
	return &LsCommand{
		fileService: fileService,
	}
}

func (c *LsCommand) Execute(args []string) error {
	// Lists folder contents
	// If args == 0, lists contents of storage/uploads
	if len(args) == 0 {
		fmt.Println("Listing folder contents")
		err := c.fileService.ListFilesInFolder("")
		if err != nil {
			return err
		}
	}
	//
	if len(args) == 1 {
		folderId := args[0]
		fmt.Println("Listing folder contents")
		err := c.fileService.ListFilesInFolder(folderId)
		if err != nil {
			return err
		}
	}
	if len(args) > 1 {
		return errors.New("Too many arguments passed")
	}

	return nil

}

func (c *LsCommand) Name() string {
	return "ls"
}

func (c *LsCommand) HelpContent() string {
	return "Lists files in a folder with given id"
}
