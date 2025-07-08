package commands

import (
	"errors"
	"filevault/services"
)

var (
	ErrInvalidCommandArgs = errors.New("Invalid command arguments. Please provide email and password.")
)

type RegisterCommand struct {
	service *services.AuthService
}

func NewRegisterCommand(service *services.AuthService) ICommand {
	return &RegisterCommand{
		service: service,
	}
}

func (c *RegisterCommand) Execute(args []string) error {
	if len(args) < 2 {
		return ErrInvalidCommandArgs
	}
	email := args[0]
	password := args[1]

	err := c.service.Register(email, password)
	if err != nil {
		return err
	}

	return nil
}

func (c *RegisterCommand) Name() string {
	return "register"
}

func (c *RegisterCommand) HelpContent() string {
	return "register <email> <password> - Register a new user with email and password"
}

