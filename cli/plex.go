package main

import (
	"cli/commands"
	"core/services"
	"fmt"
	"strings"
)

// CommandRouter acts as the Invoker in the Command Pattern.
// It holds and dispatches various commands.
type CommandRouter struct {
	commands map[string]commands.ICommand
}

func NewCommandRouter(fs *services.FileService, as *services.AuthService) *CommandRouter {
	router := &CommandRouter{
		commands: make(map[string]commands.ICommand),
	}

	/* File based commands */
	uploadCmd := commands.NewUploadCommand(fs)
	listCmd := commands.NewListCommand(fs)
	deleteCmd := commands.NewDeleteCommand(fs)
	mkdirCmd := commands.NewMkdirCommand(fs)
	lsCmd := commands.NewLsCommand(fs)
	/* Auth commands */
	registerCmd := commands.NewRegisterCommand(as)
	loginCmd := commands.NewLoginCommand(as)
	logoutCmd := commands.NewLogoutCommand(as)

	router.RegisterCommand(uploadCmd)
	router.RegisterCommand(listCmd)
	router.RegisterCommand(deleteCmd)
	router.RegisterCommand(registerCmd)
	router.RegisterCommand(loginCmd)
	router.RegisterCommand(logoutCmd)
	router.RegisterCommand(mkdirCmd)
	router.RegisterCommand(lsCmd)
	router.RegisterCommand(&HelpCommand{router: router})

	return router
}

func (r *CommandRouter) RegisterCommand(cmd commands.ICommand) {
	r.commands[cmd.Name()] = cmd
}

func (r *CommandRouter) ExecuteCommand(commandName string, args []string) error {
	cmd, exists := r.commands[commandName]
	if !exists {
		return fmt.Errorf("unknown command '%s'. Type 'help' for a list of commands.", commandName)
	}

	return cmd.Execute(args)
}

// PrintHelp displays general help content for all registered commands.
func (r *CommandRouter) PrintHelp() {
	fmt.Println("\n--- FileVault CLI Commands ---")
	for _, cmd := range r.commands {
		fmt.Printf("  %s\n", cmd.HelpContent())
	}
	fmt.Println("\nType 'exit' or 'quit' to close the CLI.")
	fmt.Println("-----------------------------")
}

type HelpCommand struct {
	router *CommandRouter
}

// Name returns the command's name.
func (c *HelpCommand) Name() string {
	return "help"
}

// Execute runs the help command.
func (c *HelpCommand) Execute(args []string) error {
	if len(args) > 1 {
		// If user types "help <command_name>"
		targetCmdName := strings.ToLower(args[1])
		if cmd, found := c.router.commands[targetCmdName]; found {
			fmt.Printf("Usage for '%s':\n  %s\n", targetCmdName, cmd.HelpContent())
		} else {
			return fmt.Errorf("help for unknown command '%s'", targetCmdName)
		}
	} else {
		c.router.PrintHelp()
	}
	return nil
}

// HelpContent returns a description for the help command.
func (c *HelpCommand) HelpContent() string {
	return "help [command] - Displays general help or specific command usage."
}
