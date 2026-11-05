package main

import (
	"log"

	"otp-service/internal/bootstrap"
)

func main() {
	if err := bootstrap.RunAPI(); err != nil {
		log.Fatalf("API stopped: %v", err)
	}
}
