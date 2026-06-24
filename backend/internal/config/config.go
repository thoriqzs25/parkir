package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment    string
	Port           string
	DatabaseURL    string
	JWTSecret      string
	LogLevel       string
	FrontendURL    string
	MigrationsPath string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Environment:    getEnv("ENV", "development"),
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/parkir?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		FrontendURL:    getEnv("FRONTEND_URL", "http://localhost:3000"),
		MigrationsPath: getEnv("MIGRATIONS_PATH", "migrations"),
	}

	if cfg.JWTSecret == "dev-secret-change-in-production" && cfg.Environment == "production" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
