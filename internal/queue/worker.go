package queue

import (
	"context"
	"log"

	"github.com/hibiken/asynq"
)

func NewWorker(redisAddr string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"otp": 1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(
				func(ctx context.Context, task *asynq.Task, err error) {
					log.Printf(
						"Task failed: type=%s error=%v",
						task.Type(),
						err,
					)
				},
			),
		},
	)
}
