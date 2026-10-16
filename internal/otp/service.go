package otp

import (
	"context"
	"errors"
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

	_, err := ResolveRecipient(input)
	if err != nil {
		return err
	}

	// TODO: Generate, store and deliver OTP
	return ErrNotImplemented
}
