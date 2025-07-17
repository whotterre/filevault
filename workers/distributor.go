package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeThumbnailGeneration(
		imagePath, fileID string,
		ctx context.Context,
		opts ...asynq.Option) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}

type PayloadThumbnailGeneration struct {
	ImagePath string `json:"image_path"`
	FileId    string `json:"file_id"`
}

// DistributeThumbnailGeneration creates a new task for thumbnail generation.
func (distributor *RedisTaskDistributor) DistributeThumbnailGeneration(
	imagePath, fileID string,
	ctx context.Context,
	opts ...asynq.Option) error {

	payload := PayloadThumbnailGeneration{
		ImagePath: imagePath,
		FileId:    fileID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON %w", err)
	}

	task := asynq.NewTask("thumbnail:generate", jsonPayload, opts...)
	taskInfo, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("Failed to enqueue task %s because %w", taskInfo.ID, err)
	}

	return nil

}
