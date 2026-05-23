package config

import "os"

type Config struct {
	ServerPort   string
	DatabaseURL  string
	RedisURL     string
	EMQXHost     string
	EMQXClientID string
	JWTSecret    string
}

func Load() (*Config, error) {
	return &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://ion:ion123@localhost:5432/ion_respirator?sslmode=disable"),
		RedisURL:     getEnv("REDIS_URL", "localhost:6379"),
		EMQXHost:     getEnv("EMQX_HOST", "tcp://localhost:1883"),
		EMQXClientID: getEnv("EMQX_CLIENT_ID", "backend-service"),
		JWTSecret:    getEnv("JWT_SECRET", "change-me-in-production"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
