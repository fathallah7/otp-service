package otp

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid request body")
	ErrDeliveryFailed = errors.New("OTP delivery failed")
	ErrResendCooldown = errors.New("please wait before requesting another OTP")
	ErrRequestRateLimit = errors.New("OTP request limit exceeded")
)
