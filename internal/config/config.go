package config

import "os"

type Config struct {
	Port          string
	Enviroment    string
	RedisAddr     string
	OTPHMACKeyHex string

	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	otpHMACKey := os.Getenv("OTP_HMAC_KEY")

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")

	return Config{
		Port:          port,
		Enviroment:    env,
		RedisAddr:     redisAddr,
		OTPHMACKeyHex: otpHMACKey,

		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
		SMTPUsername: smtpUsername,
		SMTPPassword: smtpPassword,
		SMTPFrom:     smtpFrom,
	}
}
