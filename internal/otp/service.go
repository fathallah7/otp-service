package otp

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var ErrNotImplemented = errors.New("OTP delivery not implemented")

type Service struct {
	redisClient *redis.Client
	hmacSecret  []byte
	sender      Sender
}

func NewService(redisClient *redis.Client, secret []byte, sender Sender) *Service {
	return &Service{
		redisClient: redisClient,
		hmacSecret:  secret,
		sender:      sender,
	}
}

func (s *Service) RequestOTP(ctx context.Context, input string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// 1. Resolve the recipient (e.g., email or phone number)
	recipient, err := ResolveRecipient(input)
	if err != nil {
		return err
	}

	if recipient.Channel != ChannelEmail {
		return ErrNotImplemented
	}

	// 2. Generate the OTP
	code, err := GenerateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// 3. Generate protected OTP digest
	digest := hashOTP(s.hmacSecret, recipient.Value, code)

	// 4. Generate Redis key
	key := otpRedisKey(s.hmacSecret, recipient)

	// 5. Store OTP digest in Redis
	if err := s.saveOTP(ctx, key, digest); err != nil {
		return fmt.Errorf("store OTP: %w", err)
	}

	if err := s.sender.SendOTP(ctx, recipient, code); err != nil {
		return fmt.Errorf("deliver OTP: %w", err)
	}

	return nil
}
