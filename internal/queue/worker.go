package queue

import "github.com/hibiken/asynq"

func NewWorker(redisAddr string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{
			Addr: redisAddr,
		},
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"otp": 1,
			},
		},
	)
}
