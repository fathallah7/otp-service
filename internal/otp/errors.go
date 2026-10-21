package otp

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid request body")
	ErrDeliveryFailed = errors.New("OTP delivery failed")
)
