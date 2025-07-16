package worker

import (
	"database/sql"

	"github.com/hibiken/asynq"
)
type TaskProcessor interface {

}

type RedisTaskProcessor struct {
	server *asynq.Server
	db *sql.DB
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, db *sql.DB) TaskProcessor {
	server := asynq.NewServer(redisOpt, asynq.Config{})
	return &RedisTaskProcessor{
		server: server,
		db: db,
	}
}