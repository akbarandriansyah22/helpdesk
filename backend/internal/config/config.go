package config

import (
	"os"

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
	// Load .env file (tidak error kalau file tidak ada)
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:     getenv("DB_HOST", "localhost"),
		DBPort:     getenv("DB_PORT", "5432"),
		DBUser:     getenv("DB_USER", "postgres"),
		DBPassword: getenv("DB_PASSWORD", "1234"),
		DBName:     getenv("DB_NAME", "postgres"),
		JWTSecret:  getenv("JWT_SECRET", "supersecret"),
		AppPort:    getenv("APP_PORT", "8080"),
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
