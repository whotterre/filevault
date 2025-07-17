package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

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
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
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

	// Generate thumbnail
	err := GenerateThumbnail(payload.ImagePath, payload.FileId)
	if err != nil {
		return fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	log.Printf("Successfully generated thumbnail for file: %s", payload.ImagePath)
	return nil
}
