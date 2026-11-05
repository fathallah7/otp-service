package otp

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const otpTTL = 5 * time.Minute

// Save OTP and reset failed attempts atomically.
var saveOTPScript = redis.NewScript(`
	redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
	redis.call("DEL", KEYS[2])
	return 1
`)

func (s *Service) saveOTP(
	ctx context.Context,
	key, digest string,
) error {
	return saveOTPScript.Run(
		ctx,
		s.redisClient,
		[]string{key, key + ":attempts"},
		digest,
		otpTTL.Milliseconds(),
	).Err()
}

var verifyOTPScript = redis.NewScript(`
    local stored = redis.call("GET", KEYS[1])

	-- OTP doesn't exist or has expired
	if not stored then
		return 0
	end

	local attempts = tonumber(
		redis.call("GET", KEYS[2]) or "0"
	)

	-- Maximum attempts already reached
	if attempts >= 5 then
		return 0
	end

	-- Correct OTP: delete digest and attempt counter
	if stored == ARGV[1] then
		redis.call("DEL", KEYS[1], KEYS[2])
		return 1
	end

	-- Incorrect OTP: increment failed attempts
	attempts = redis.call("INCR", KEYS[2])

	-- Make counter expire with the OTP
	local ttl = redis.call("PTTL", KEYS[1])
	if ttl > 0 then
		redis.call("PEXPIRE", KEYS[2], ttl)
	end

	-- Invalidate OTP after 5 wrong attempts
	if attempts >= 5 then
		redis.call("DEL", KEYS[1])
	end

	return 0
`)

func (s *Service) consumeOTP(
	ctx context.Context,
	key,
	digest string,
) (bool, error) {
	result, err := verifyOTPScript.Run(
		ctx,
		s.redisClient,
		[]string{
			key,
			key + ":attempts",
		},
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
