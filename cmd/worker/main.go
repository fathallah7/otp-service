package main

import (
	"log"

	"otp-service/internal/bootstrap"
)

func main() {
	if err := bootstrap.RunWorker(); err != nil {
		log.Fatalf("Worker stopped: %v", err)
	}
}
