package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port        int
	DatabaseURL string
	JWTSecret   string
}

func Load() *Config {
	cfg := &Config{
		Port:        getEnvInt("PORT", 8085),
		DatabaseURL: getEnvString("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/authdb?sslmode=disable"),
		JWTSecret:   getEnvString("JWT_SECRET", "mysecret"),
	}
	if cfg.DatabaseURL == "" || cfg.JWTSecret == "" {
		log.Fatal("DATABASE_URL and JWT_SECRET are required")
	}
	return cfg
}
func getEnvString(key, fullback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fullback
	}
	return value
}
func getEnvInt(key string, fullback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fullback
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return fullback
	}
	return i
}
