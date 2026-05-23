package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerPort   string
	DatabaseURL  string
	RedisURL     string
	EMQXHost     string
	EMQXClientID string
	JWTSecret    string
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		RedisURL:     getEnv("REDIS_URL", "redis://localhost:6379"),
		EMQXHost:     getEnv("EMQX_HOST", "tcp://localhost:1883"),
		EMQXClientID: getEnv("EMQX_CLIENT_ID", "backend-service"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
