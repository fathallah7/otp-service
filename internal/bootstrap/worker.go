package bootstrap

import (
	"fmt"
	"log"

	"otp-service/internal/queue"

	"github.com/hibiken/asynq"
)

// RunWorker initializes the delivery worker and processes OTP tasks until the
// worker stops.
func RunWorker() error {
	common, err := NewCommon()
	if err != nil {
		return err
	}
	defer common.Close()

	// Initialize Task Handler
	taskHandler := queue.NewTaskHandler(common.OTPService)

	// Register task handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(
		queue.TypeOTPDelivery,
		taskHandler.HandleOTPDeliveryTask,
	)

	// Start Worker
	worker := queue.NewWorker(common.Config.RedisAddr)

	log.Println("OTP Worker starting...")

	if err := worker.Run(mux); err != nil {
		return fmt.Errorf("run worker: %w", err)
	}

	return nil
}
