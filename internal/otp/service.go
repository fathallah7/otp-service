package otp

import (
	"context"
	"errors"
	"fmt"
)

var ErrNotImplemented = errors.New("OTP delivery not implemented")

type Service struct{}

func NewService() *Service {
	return &Service{}
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

	// 2. Generate the OTP
	code, err := GenerateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// TODO: store and deliver OTP
	_ = recipient
	_ = code

	return ErrNotImplemented
}
