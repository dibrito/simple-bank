package worker

import (
	"context"

	db "github.com/dibrito/simple-bank/db/sqlc"
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynq/internal/context"
)

type TaskProcessor interface {
	// we need to register the task within the server
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.SQLStore
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.SQLStore) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{},
	)
	return &RedisTaskProcessor{
		server: server,
		store:  store,
	}
}

// we need to register the tasks to eachs handlers
func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	//  this handle func analogies are crazyyieee!!
	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)
	return processor.server.Start(mux)
}
