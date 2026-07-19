package config

import (
	"os"
	"time"
)

type Config struct {
	DatabaseURL    string
	APIPort        string
	JWTSecret      string
	JWTAccessTTL   time.Duration
	JWTRefreshTTL  time.Duration
	OpenAIToken    string
	AIModelText    string
	GoogleClientID string
}

func Load() Config {
	return Config{
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		APIPort:        getEnv("API_PORT", "8080"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTAccessTTL:   parseDuration(getEnv("JWT_ACCESS_TTL", "15m")),
		JWTRefreshTTL:  parseDuration(getEnv("JWT_REFRESH_TTL", "720h")), // 30 days
		OpenAIToken:    getEnv("OPENAI_TOKEN", ""),
		AIModelText:    getEnv("AI_MODEL_TEXT", "gpt-4o"),
		GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}
