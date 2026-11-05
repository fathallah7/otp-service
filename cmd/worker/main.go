package main

import (
	"log"

	"otp-service/internal/bootstrap"
	"otp-service/internal/queue"

	"github.com/hibiken/asynq"
)

func main() {
	// Initialize shared dependencies
	deps, err := bootstrap.NewCommon()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer deps.Close()

	// Initialize Task Handler
	taskHandler := queue.NewTaskHandler(deps.OTPService)

	// Register task handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(
		queue.TypeOTPDelivery,
		taskHandler.HandleOTPDeliveryTask,
	)

	// Start Worker
	worker := queue.NewWorker(deps.Config.RedisAddr)

	log.Println("OTP Worker starting...")

	if err := worker.Run(mux); err != nil {
		log.Fatal(err)
	}
}
