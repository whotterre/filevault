package commands

import (
	"core/services"
	"errors"
	"fmt"
)

type DeleteCommand struct {
	fileService *services.FileService
}

func NewDeleteCommand(fileService *services.FileService) ICommand {
	return &DeleteCommand{
		fileService: fileService,
	}
}

func (c *DeleteCommand) Execute(args []string) error {
	fmt.Print(args)
	if len(args) < 1 {
		return errors.New("no file id provided")
	}
	fileId := args[0]
	if err := c.fileService.DeleteFile(fileId); err != nil {
		return err
	}
	return nil
}

func (c *DeleteCommand) Name() string {
	return "delete"
}

func (c *DeleteCommand) HelpContent() string {
	return `
			Deletes a file from the vault. Usage: delete <filepath>.
			Eg vault delete /file.txt
		`
}
