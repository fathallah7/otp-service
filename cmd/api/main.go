package main

import (
	"context"
	"log"
	"net/http"
	"otp-service/internal/config"
	"otp-service/internal/otp"
	"otp-service/internal/server"
	"otp-service/internal/storage"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode based on the environment
	if cfg.Enviroment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	redisClient, err := storage.NewRedis(ctx, cfg.RedisAddr)
	cancel()

	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	defer redisClient.Close()

	log.Println("Connected to Redis successfully")

	// Create Service
	otpService := otp.NewService(redisClient)

	// Create Handler and inject Service
	otpHandler := otp.NewHandler(otpService)

	// Create http router
	router, err := server.NewRouter(otpHandler)
	if err != nil {
		log.Fatalf("Failed to create router: %v", err)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start the server
	log.Printf("Starting server on port %s in %s mode", cfg.Port, cfg.Enviroment)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
