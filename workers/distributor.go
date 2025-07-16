package worker

import (
	// "context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	// DistributeThumbnailGen(ctx context.Context, taskID string, filePath string) error
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





