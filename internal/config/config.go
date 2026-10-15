package config

import "os"

type Config struct {
	Port       string
	Enviroment string
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

	return Config{
		Port:       port,
		Enviroment: env,
	}
}
