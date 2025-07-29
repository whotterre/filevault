package commands

import "core/services"

type LoginCommand struct {
	services *services.AuthService
}

func NewLoginCommand(service *services.AuthService) ICommand {
	return &LoginCommand{
		services: service,
	}
}

func (c *LoginCommand) Execute(args []string) error {
	if len(args) < 2 {
		return ErrInvalidCommandArgs
	}
	email := args[0]
	password := args[1]

	err := c.services.Login(email, password)
	if err != nil {
		return err
	}

	return nil
}

func (c *LoginCommand) Name() string {
	return "login"
}

func (c *LoginCommand) HelpContent() string {
	return "login <email> <password> - Log in with your email and password"
}
