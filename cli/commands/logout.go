package commands

import (
	"core/services"
	"errors"
)

var (
	ErrInvalidArgs = errors.New("Invalid number of arguments provided")
)

type LogoutCommand struct {
	authService *services.AuthService
}

func NewLogoutCommand(authService *services.AuthService) ICommand {
	return &LogoutCommand{
		authService: authService,
	}
}

func (c *LogoutCommand) Execute(args []string) error {
	if len(args) > 0 {

		return ErrInvalidArgs
	}

	err := c.authService.Logout()
	if err != nil {
		return err
	}
	return nil
}

func (c *LogoutCommand) Name() string {
	return "logout"
}

func (c *LogoutCommand) HelpContent() string {
	return `Logs out a user`
}
