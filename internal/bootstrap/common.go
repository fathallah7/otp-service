package bootstrap

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"otp-service/internal/config"
	"otp-service/internal/delivery"
	"otp-service/internal/otp"
	"otp-service/internal/queue"
	"otp-service/internal/storage"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

// Common holds the dependencies shared by the API and worker entry points.
type Common struct {
	Config      config.Config
	RedisClient *redis.Client
	Producer    *queue.Producer
	OTPService  *otp.Service
}

// NewCommon loads configuration and initializes the dependencies shared by the
// API and worker processes.
func NewCommon() (*Common, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := config.Load()

	secret, err := decodeHMACSecret(cfg.OTPHMACKeyHex)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	redisClient, err := storage.NewRedis(ctx, cfg.RedisAddr)
	cancel()

	if err != nil {
		return nil, fmt.Errorf("connect to Redis: %w", err)
	}

	producer := queue.NewProducer(cfg.RedisAddr)

	sender := delivery.NewDispatcher(
		delivery.NewEmailSender(
			cfg.SMTPHost,
			cfg.SMTPPort,
			cfg.SMTPUsername,
			cfg.SMTPPassword,
			cfg.SMTPFrom,
		),
		delivery.NewWhatsAppSender(
			cfg.WhatsAppAPIVersion,
			cfg.WhatsAppPhoneNumberID,
			cfg.WhatsAppAccessToken,
			cfg.WhatsAppTemplateName,
			cfg.WhatsAppTemplateLanguage,
		),
	)

	log.Println("Connected to Redis successfully")

	return &Common{
		Config:      cfg,
		RedisClient: redisClient,
		Producer:    producer,
		OTPService:  otp.NewService(redisClient, secret, sender, producer),
	}, nil
}

// Close releases the resources owned by Common.
func (c *Common) Close() error {
	var errs []error

	if c.Producer != nil {
		errs = append(errs, c.Producer.Close())
	}

	if c.RedisClient != nil {
		errs = append(errs, c.RedisClient.Close())
	}

	return errors.Join(errs...)
}

func decodeHMACSecret(value string) ([]byte, error) {
	secret, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("invalid OTP_HMAC_KEY format: %w", err)
	}

	if len(secret) != 32 {
		return nil, errors.New("OTP_HMAC_KEY must contain exactly 32 bytes")
	}

	return secret, nil
}
