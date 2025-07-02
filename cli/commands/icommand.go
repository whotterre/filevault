package commands
// To use the Command Pattern https://refactoring.guru/design-patterns/command/go/example#example-0.
// ICommand defines the interface for a command in the CLI application.
type ICommand interface {
	// Execute runs the command with the provided arguments.
	Execute(args []string) error
	Name() string
	HelpContent() string
}
