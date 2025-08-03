package main

import (
	"api/routes"
	"core/db"
	"core/repositories"
	"core/services"
	worker "core/workers"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
)

func main() {
	// Initialize dependencies
	// Initialize repos
	dbConn, err := db.GetSQLiteDBConn()
	if err != nil {
		log.Printf("Error connecting to database: %v\n", err)
		return
	}
	// Redis backing store
	redisClient, err := db.GetRedisClient()
	if err != nil {
		log.Printf("Error connecting to Redis: %v\n", err)
		return
	}
	authRepo := repositories.NewUserRepository(dbConn)
	fileRepo := repositories.NewFileRepository(dbConn)
	sessionRepo := repositories.NewSessionRepository(redisClient)
	redisOpt := asynq.RedisClientOpt{Addr: "172.17.0.3:6379"}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, dbConn)
	go func() {
		fmt.Println("Starting task processor...")
		if err := taskProcessor.Start(); err != nil {
			fmt.Printf("Failed to start task processor: %v\n", err)
		}
	}()
	// Initialize services
	fileService := services.NewFileService(redisClient, fileRepo, taskDistributor)
	authService := services.NewAuthService(redisClient, authRepo, sessionRepo)
	server := routes.NewAPIServer(authRepo, fileRepo, sessionRepo,
		fileService, authService, redisClient, taskDistributor)
	server.Start()
}
