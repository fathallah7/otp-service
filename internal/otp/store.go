package otp

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
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

var verifyOTPScript = redis.NewScript(`
    local stored = redis.call("GET", KEYS[1])

    if not stored then
        return 0
    end

    if stored ~= ARGV[1] then
        return 0
    end

    redis.call("DEL", KEYS[1])
    return 1
`)

func (s *Service) consumeOTP(
	ctx context.Context,
	key,
	digest string,
) (bool, error) {
	result, err := verifyOTPScript.Run(
		ctx,
		s.redisClient,
		[]string{key},
		digest,
	).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}
