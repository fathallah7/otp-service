package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type OTPProcessor interface {
	ProcessQueuedOTP(ctx context.Context, requestID, recipient string) error
}

type TaskHandler struct {
	processor OTPProcessor
}

func NewTaskHandler(processor OTPProcessor) *TaskHandler {
	return &TaskHandler{
		processor: processor,
	}
}

func (h *TaskHandler) HandleOTPDeliveryTask(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload OTPDeliveryPayload

	// 1. Decode task payload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode OTP task: %w: %w", err, asynq.SkipRetry)
	}

	// 2. Validate required fields
	if payload.RequestID == "" || payload.Recipient == "" {
		return fmt.Errorf("missing OTP task fields: %w", asynq.SkipRetry)
	}

	// 3. Execute business logic
	return h.processor.ProcessQueuedOTP(
		ctx,
		payload.RequestID,
		payload.Recipient,
	)
}
