package main

import (
	"context"
	"encoding/hex"
	"log"
	"time"

	"otp-service/internal/config"
	"otp-service/internal/delivery"
	"otp-service/internal/otp"
	"otp-service/internal/queue"
	"otp-service/internal/storage"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
	// Load configuration
	_ = godotenv.Load()
	cfg := config.Load()

	// Connect to Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	redisClient, err := storage.NewRedis(ctx, cfg.RedisAddr)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	// Decode HMAC secret
	secret, err := hex.DecodeString(cfg.OTPHMACKeyHex)
	if err != nil || len(secret) != 32 {
		log.Fatal("Invalid OTP HMAC key")
	}

	producer := queue.NewProducer(cfg.RedisAddr)
	defer producer.Close()

	// Initialize delivery
	emailSender := delivery.NewEmailSender(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPFrom,
	)

	whatsappSender := delivery.NewWhatsAppSender(
		cfg.WhatsAppAPIVersion,
		cfg.WhatsAppPhoneNumberID,
		cfg.WhatsAppAccessToken,
		cfg.WhatsAppTemplateName,
		cfg.WhatsAppTemplateLanguage,
	)

	sender := delivery.NewDispatcher(emailSender, whatsappSender)

	// Initialize OTP Service and Task Handler
	otpService := otp.NewService(redisClient, secret, sender, producer)
	taskHandler := queue.NewTaskHandler(otpService)

	// Register task handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(
		queue.TypeOTPDelivery,
		taskHandler.HandleOTPDeliveryTask,
	)

	// Start Worker
	worker := queue.NewWorker(cfg.RedisAddr)

	log.Println("OTP Worker starting...")

	if err := worker.Run(mux); err != nil {
		log.Fatal(err)
	}
}
