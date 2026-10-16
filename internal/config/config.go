package config

import "os"

type Config struct {
	Port       string
	Enviroment string
	RedisAddr string
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

	return Config{
		Port:       port,
		Enviroment: env,
		RedisAddr:  redisAddr,
	}
}
