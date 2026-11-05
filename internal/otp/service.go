package otp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// Interface implemented by queue.Producer
type OTPEnqueuer interface {
	EnqueueOTP(ctx context.Context, requestID, recipient string) error
}

type Service struct {
	redisClient *redis.Client
	hmacSecret  []byte
	sender      Sender
	enqueuer    OTPEnqueuer
}

func NewService(
	redisClient *redis.Client,
	secret []byte,
	sender Sender,
	enqueuer OTPEnqueuer,
) *Service {
	return &Service{
		redisClient: redisClient,
		hmacSecret:  secret,
		sender:      sender,
		enqueuer:    enqueuer,
	}
}

// Called by the HTTP API
func (s *Service) RequestOTP(ctx context.Context, input string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// 1. Validate and normalize recipient
	recipient, err := ResolveRecipient(input)
	if err != nil {
		return err
	}

	// 2. Generate a unique request ID
	var id [16]byte

	if _, err := rand.Read(id[:]); err != nil {
		return fmt.Errorf("generate request ID: %w", err)
	}

	requestID := hex.EncodeToString(id[:])

	// 3. Enqueue the task
	if s.enqueuer == nil {
		return fmt.Errorf("OTP enqueuer is not configured")
	}

	result, err := s.checkOTPRequestLimits(
		ctx,
		recipient,
		requestID,
	)

	if err != nil {
		return fmt.Errorf("OTP request rate limit: %w", err)
	}

	switch result {
	case 1:
		return ErrResendCooldown
	case 2:
		return ErrRequestRateLimit
	case 0:
	default:
		return fmt.Errorf("unexpected rate limit result: %d", result)
	}

	if err := s.enqueuer.EnqueueOTP(
		ctx,
		requestID,
		recipient.Value,
	); err != nil {
		return fmt.Errorf("enqueue OTP: %w", err)
	}

	return nil
}

// Called by the Asynq Worker
func (s *Service) ProcessQueuedOTP(
	ctx context.Context,
	requestID, input string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if requestID == "" {
		return fmt.Errorf("missing request ID: %w", ErrInvalidRequest)
	}

	// 1. Resolve recipient
	recipient, err := ResolveRecipient(input)
	if err != nil {
		return err
	}

	// 2. Generate OTP
	code, err := GenerateOTP()
	if err != nil {
		return fmt.Errorf("generate OTP: %w", err)
	}

	// 3. Hash OTP
	digest := hashOTP(s.hmacSecret, recipient.Value, code)

	// 4. Generate Redis key
	key := otpRedisKey(s.hmacSecret, recipient)

	// 5. Store OTP digest
	if err := s.saveOTP(ctx, key, digest); err != nil {
		return fmt.Errorf("store OTP: %w", err)
	}

	// 6. Deliver OTP
	if err := s.sender.SendOTP(ctx, recipient, code); err != nil {
		return fmt.Errorf("%w: %w", ErrDeliveryFailed, err)
	}

	log.Printf(
		"OTP delivered: request=%s channel=%s",
		requestID,
		recipient.Channel,
	)

	return nil
}

func (s *Service) VerifyOTP(
	ctx context.Context,
	input, code string,
) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	// 1. Validate recipient
	recipient, err := ResolveRecipient(input)
	if err != nil {
		return false, err
	}

	// 2. Validate OTP format
	if len(code) != 6 {
		return false, ErrInvalidRequest
	}

	for _, digit := range code {
		if digit < '0' || digit > '9' {
			return false, ErrInvalidRequest
		}
	}

	// 3. Generate Redis key
	key := otpRedisKey(s.hmacSecret, recipient)

	// 4. Hash submitted OTP
	digest := hashOTP(s.hmacSecret, recipient.Value, code)

	// 5. Verify and consume OTP
	valid, err := s.consumeOTP(ctx, key, digest)
	if err != nil {
		return false, fmt.Errorf("verify OTP: %w", err)
	}

	return valid, nil
}
