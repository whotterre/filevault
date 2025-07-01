package main

import (
	"bufio"
	"filevault/cli"
	"filevault/services"
	"fmt"
	"os"
	"strings"
)

func init() {
	cli.Welcome()
	cli.Help()

}

func main() {
	// Get user input
	scanner := bufio.NewScanner(os.Stdin)
	fileService := services.NewFileService()
	cm := cli.NewCommandRouter(fileService)
	for {
		// Prompt for input
		cli.Prompt()
		// Read user input
		if scanner.Scan() {
			input := scanner.Text()

			// Parse the input into command and arguments
			parts := strings.Fields(input)
			if len(parts) == 0 {
				continue
			}

			commandName := parts[0]
			args := parts[1:]

			err := cm.ExecuteCommand(commandName, args)
			if err != nil {
				cli.Error(err)
				continue
			}
			fmt.Printf("Command '%s' executed successfully\n", commandName)
		}

		if err := scanner.Err(); err != nil {
			cli.Error(err)
			break
		}

	}
}
