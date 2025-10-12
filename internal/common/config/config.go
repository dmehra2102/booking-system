package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServiceName string
	ServicePort string
	Environment string
	LogLevel    string

	// Database
	PostgresURL string
	RedisURL    string

	// Kafka
	KafkaBrokers []string

	// Observability
	JaegerEndpoint string
	MetricsPort    string

	// Security
	JWTSecret string
	JWTExpiry time.Duration

	// SMTP
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		ServiceName: getEnvOrDefault("SERVICE_NAME", "booking-service"),
		ServicePort: getEnvOrDefault("SERVICE_PORT", "8080"),
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
		LogLevel:    getEnvOrDefault("LOG_LEVEL", "info"),

		PostgresURL: getEnvOrDefault("POSTGRES_URL", "postgres://booking_user:booking_pass@localhost:5432/booking_db?sslmode=disable"),
		RedisURL:    getEnvOrDefault("REDIS_URL", "redis://localhost:6379"),

		KafkaBrokers: strings.Split(getEnvOrDefault("KAFKA_BROKERS", "localhost:29092"), ","),

		JaegerEndpoint: getEnvOrDefault("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
		MetricsPort:    getEnvOrDefault("METRICS_PORT", "2112"),

		JWTSecret: getEnvOrDefault("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		JWTExpiry: parseDurationOrDefault(getEnvOrDefault("JWT_EXPIRY", "24h")),

		SMTPHost:     getEnvOrDefault("SMTP_HOST", "localhost"),
		SMTPPort:     parseIntOrDefault(getEnvOrDefault("SMTP_PORT", "1025")),
		SMTPUsername: getEnvOrDefault("SMTP_USERNAME", ""),
		SMTPPassword: getEnvOrDefault("SMTP_PASSWORD", ""),
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseIntOrDefault(value string) int {
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return 0
}

func parseDurationOrDefault(value string) time.Duration {
	if d, err := time.ParseDuration(value); err == nil {
		return d
	}
	return 24 * time.Hour
}
