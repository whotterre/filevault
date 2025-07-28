package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

const TypeThumbnailGeneration = "thumbnail:generate"

type TaskProcessor interface {
	ProcessThumbnailGenerationTask(ctx context.Context, task *asynq.Task) error
	Start() error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	db     *sql.DB
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, db *sql.DB) TaskProcessor {
	server := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 10,
		// Add timeout for task processing
		ShutdownTimeout: 30 * time.Second,
	})

	return &RedisTaskProcessor{
		server: server,
		db:     db,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeThumbnailGeneration, processor.ProcessThumbnailGenerationTask)

	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) ProcessThumbnailGenerationTask(ctx context.Context, task *asynq.Task) error {
	var payload PayloadThumbnailGeneration
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	log.Printf("Processing thumbnail generation for file: %s (ID: %s)", payload.ImagePath, payload.FileId)

	// Create a channel to handle the result
	done := make(chan error, 1)

	// Run thumbnail generation in a goroutine to respect context cancellation
	go func() {
		err := GenerateThumbnail(payload.ImagePath, payload.FileId)
		done <- err
	}()

	// Wait for either completion or context cancellation
	select {
	case err := <-done:
		if err != nil {
			log.Printf("Failed to generate thumbnail for file %s: %v", payload.ImagePath, err)
			return fmt.Errorf("failed to generate thumbnail: %w", err)
		}
		log.Printf("Successfully generated thumbnail for file: %s", payload.ImagePath)
		return nil
	case <-ctx.Done():
		log.Printf("Thumbnail generation cancelled for file: %s", payload.ImagePath)
		return fmt.Errorf("thumbnail generation cancelled: %w", ctx.Err())
	}
}
