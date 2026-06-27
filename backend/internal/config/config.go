package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment         string
	Port                string
	DatabaseURL         string
	JWTSecret           string
	JWTPrivateKeyPath   string
	JWTPublicKeyPath    string
	LogLevel            string
	FrontendURL         string
	MigrationsPath      string
	SMTPHost            string
	SMTPPort            string
	SMTPUser            string
	SMTPPassword        string
	SMTPFromAddress     string
	AlertEmailRecipients []string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Environment:          getEnv("ENV", "development"),
		Port:                 getEnv("PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/parkir?sslmode=disable"),
		JWTSecret:            getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTPrivateKeyPath:    getEnv("JWT_PRIVATE_KEY_PATH", "dev-keys/jwt-private.pem"),
		JWTPublicKeyPath:     getEnv("JWT_PUBLIC_KEY_PATH", "dev-keys/jwt-public.pem"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		FrontendURL:          getEnv("FRONTEND_URL", "http://localhost:3000"),
		MigrationsPath:       getEnv("MIGRATIONS_PATH", "migrations"),
		SMTPHost:             getEnv("SMTP_HOST", ""),
		SMTPPort:             getEnv("SMTP_PORT", "587"),
		SMTPUser:             getEnv("SMTP_USER", ""),
		SMTPPassword:         getEnv("SMTP_PASSWORD", ""),
		SMTPFromAddress:      getEnv("SMTP_FROM_ADDRESS", ""),
		AlertEmailRecipients: splitEnv(getEnv("ALERT_EMAIL_RECIPIENTS", "")),
	}

	if cfg.Environment == "production" {
		if cfg.JWTPrivateKeyPath == "" || cfg.JWTPublicKeyPath == "" {
			return nil, fmt.Errorf("JWT_PRIVATE_KEY_PATH and JWT_PUBLIC_KEY_PATH must be set in production")
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func splitEnv(val string) []string {
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
