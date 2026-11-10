package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	AppPort    string
}

func LoadConfig() (*Config, error) {
	// Missing .env is fine. Real environment variables are left untouched.
	_ = godotenv.Load()
	_ = godotenv.Load("backend/.env")

	cfg := &Config{
		DBHost:     getenv("DB_HOST", "localhost"),
		DBPort:     getenv("DB_PORT", "5432"),
		DBUser:     getenv("DB_USER", "postgres"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     getenv("DB_NAME", "postgres"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		AppPort:    getenv("APP_PORT", "8080"),
	}

	if strings.TrimSpace(cfg.DBPassword) == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if strings.TrimSpace(cfg.DBName) == "" {
		return nil, fmt.Errorf("DB_NAME is required")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
