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

var otpRequestLimitScript = redis.NewScript(`
    if redis.call("EXISTS", KEYS[1]) == 1 then
        return 1
    end

    local count = tonumber(redis.call("GET", KEYS[2]) or "0")

    if count >= 5 then
        return 2
    end

    count = redis.call("INCR", KEYS[2])

    if count == 1 then
        redis.call("EXPIRE", KEYS[2], 3600)
    end

    redis.call("SET", KEYS[1], ARGV[1], "EX", 60)

    return 0
`)

func (s *Service) checkOTPRequestLimits(
	ctx context.Context,
	recipient Recipient,
	requestID string,
) (int, error) {
	key := otpRedisKey(s.hmacSecret, recipient)

	return otpRequestLimitScript.Run(
		ctx,
		s.redisClient,
		[]string{
			key + ":cooldown",
			key + ":hourly",
		},
		requestID,
	).Int()
}
