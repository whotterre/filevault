package main

import (
	"bufio"
	"filevault/cli"
	"filevault/db"
	"filevault/repositories"
	"filevault/services"
	worker "filevault/workers"
	"fmt"
	"os"
	"strings"

	"github.com/hibiken/asynq"
)

func init() {
	cli.Welcome()
	cli.Help()
}

func main() {
	// Get user input
	scanner := bufio.NewScanner(os.Stdin)
	// Connect to backing stores
	// SQLite3 backing store
	dbConn, err := db.GetSQLiteDBConn()
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return 
	}
	// Redis backing store
	redisClient, err := db.GetRedisClient()
	if err != nil {
		fmt.Printf("Error connecting to Redis: %v\n", err)
		return
	}
	// Initialize async task distributor 
	redisOpt := asynq.RedisClientOpt{Addr: "localhost:6379"}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)
	// Initialize repos
	authRepo := repositories.NewUserRepository(dbConn)
	fileRepo := repositories.NewFileRepository(dbConn)
	sessionRepo := repositories.NewSessionRepository(redisClient)
	// Initialize services
	fileService := services.NewFileService(redisClient, fileRepo, taskDistributor)
	authService := services.NewAuthService(redisClient, authRepo, sessionRepo)
	cm := cli.NewCommandRouter(fileService, authService)

	for {
		// Prompt for input
		cli.Prompt()
		// Read user input
		if scanner.Scan() {
			input := scanner.Text()

			// Tokenize the input into command and arguments
			parts := strings.Fields(input)
			if len(parts) == 0 {
				continue
			}

			// Handle special commands that don't need "vault" prefix
			if parts[0] == "exit" || parts[0] == "quit" {
				fmt.Println("Goodbye!")
				break
			}

			if parts[0] == "help" {
				err := cm.ExecuteCommand("help", parts)
				if err != nil {
					cli.Error(err)
				}
				continue
			}

			// Handle vault commands
			if parts[0] == "vault" {
				if len(parts) < 2 {
					cli.Error(fmt.Errorf("vault command requires a subcommand. Usage: vault <command> [args...]"))
					continue
				}

				commandName := parts[1]
				args := parts[2:]

				err := cm.ExecuteCommand(commandName, args)
				if err != nil {
					cli.Error(err)
					continue
				}
				fmt.Printf("Command '%s' executed successfully\n", commandName)
			} else {
				cli.Error(fmt.Errorf("unknown command '%s'. Commands must start with 'vault' (e.g., 'vault upload file.txt')", parts[0]))
			}
		}

		if err := scanner.Err(); err != nil {
			cli.Error(err)
			break
		}
	}
}