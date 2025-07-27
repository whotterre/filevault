package commands

import (
	"errors"
	"filevault/services"
	"fmt"
)

type MkdirCommand struct {
	fileService *services.FileService
}

func NewMkdirCommand(fileService *services.FileService) ICommand {
	return &MkdirCommand{
		fileService: fileService,
	}
}

func (c *MkdirCommand) Execute(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no directory name provided")
	}
	dirName := args[0]
	if len(args) == 1 {
		fmt.Printf("Creating directory: %s\n", dirName)
		if err := c.fileService.CreateEmptyFolder(dirName, ""); err != nil {
			return err
		}
	} else if len(args) == 2 {
		parentId := args[1]
		fmt.Printf("Creating subdirectory %s in directory with id of %s\n", dirName, parentId)
		if err := c.fileService.CreateEmptyFolder(dirName, parentId); err != nil {
			return err
		}
	} else {
		return errors.New("Too many arguments provided to mkdir")
	}

	return nil
}

func (c *MkdirCommand) Name() string {
	return "mkdir"
}

func (c *MkdirCommand) HelpContent() string {
	return "Create a new directory in the vault. Usage: mkdir <directory_name>"
}
