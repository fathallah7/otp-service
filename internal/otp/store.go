package otp

import (
	"context"
	"time"
)

const otpTTL = 5 * time.Minute

func (s *Service) saveOTP(
	ctx context.Context,
	key string,
	digest string,
) error {
	return s.redisClient.Set(
		ctx,
		key,
		digest,
		otpTTL,
	).Err()
}
