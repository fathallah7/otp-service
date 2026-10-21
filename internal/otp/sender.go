package otp

import "context"

type Sender interface {
	SendOTP(ctx context.Context, recipient Recipient, code string) error
}
