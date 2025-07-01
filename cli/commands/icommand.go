package commands

// ICommand defines the interface for a command in the CLI application.
type ICommand interface {
	// Execute runs the command with the provided arguments.
	Execute(args []string) error
	Name() string
	HelpContent() string
}
