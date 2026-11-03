package queue

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

type Producer struct {
	client *asynq.Client
}

func NewProducer(redisAddr string) *Producer {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: redisAddr,
	})

	return &Producer{
		client: client,
	}
}

func (p *Producer) EnqueueOTP(
	ctx context.Context,
	requestID, recipient string,
) error {
	// 1. Create the task
	task, err := NewOTPDeliveryTask(requestID, recipient)
	if err != nil {
		return err
	}

	// 2. Add task to Redis queue
	_, err = p.client.EnqueueContext(
		ctx,
		task,
		asynq.Queue("otp"),
		asynq.MaxRetry(3),
	)
	if err != nil {
		return fmt.Errorf("enqueue OTP task: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.client.Close()
}
