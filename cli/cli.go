package cli


import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// Display welcome message
func Welcome() {
	println("Welcome to the FileVault CLI!")
	println("Use 'help' to see available commands.")
	println("Type 'exit' to close the CLI.")
}

// Display help message
func Help() {
	println("Available commands:")
	println("  help    - Show this help message")
	println("  exit    - Exit the application")
	println("  read    - Displays metadata for a specific file")
	println("  vault   - Manage vaults")
	println("  ls [parent_id] - Lists contents of a folder")
	println("  mkdir <folder_name> [parent_id] - Creates an empty folder")
	println("  upload <filepath> - Manage files in vaults")
	println("  publish <fileId>  - Makes a file or folder public")
	println("  unpublish <fileId>  - Makes a file or folder private")
}

// Display exit message
func Exit() {
	println("'\033[92m'Exiting FileVault CLI. Goodbye!")
	os.Exit(0)
}

func Prompt() {
	fmt.Print("filevault> ")
}
func Error(err error) {
	fmt.Printf("\033[91mError: %v\033[0m\n", err)
}

func Clear() {
	fmt.Print("\033[2J\033[H")
}

func IsValidCommand(command string) bool {
	var validCommands = []string{
		"help",
		"exit",
		"upload",
		"list",
		"mkdir",
		"ls",
		"publish",
		"unpublish",
		"read",
		"delete",
	}
	return slices.Contains(validCommands, command)
}

func SplitCommand(input string) []string {
	return strings.Fields(input)
}