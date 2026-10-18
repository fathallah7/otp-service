package otp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func otpRedisKey(secret []byte, recipient Recipient) string {
	mac := hmac.New(sha256.New, secret)

	mac.Write([]byte("otp:key:v1"))
	mac.Write([]byte{0})
	mac.Write([]byte(recipient.Channel))
	mac.Write([]byte{0})
	mac.Write([]byte(recipient.Value))

	return "otp:request:" + hex.EncodeToString(mac.Sum(nil))
}

func hashOTP(secret []byte, recipient, code string) string {
	mac := hmac.New(sha256.New, secret)

	mac.Write([]byte(recipient))
	mac.Write([]byte{0})
	mac.Write([]byte(code))

	return hex.EncodeToString(mac.Sum(nil))
}
