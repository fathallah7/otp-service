package main

import (
	"context"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"otp-service/internal/config"
	"otp-service/internal/delivery"
	"otp-service/internal/otp"
	"otp-service/internal/server"
	"otp-service/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

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

	// Decode HMAC secret
	secret, err := hex.DecodeString(cfg.OTPHMACKeyHex)

	if err != nil {
		log.Fatalf("Invalid OTP_HMAC_KEY format: %v", err)
	}

	if len(secret) != 32 {
		log.Fatal("OTP_HMAC_KEY must contain exactly 32 bytes")
	}

	emailSender := delivery.NewEmailSender(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPFrom,
	)

	// Create Service
	otpService := otp.NewService(redisClient, secret, emailSender)

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
