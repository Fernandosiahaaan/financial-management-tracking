package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application.
type Config struct {
	Port        int
	DatabaseURL string
	Env         string
	JWTSecret   string
	JWTExpiry   time.Duration
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	port := getEnvInt("PORT", 8080)
	dbURL := getEnv("DATABASE_URL", "postgres://fintrack:fintrack@localhost:5432/fintrack?sslmode=disable")
	env := getEnv("APP_ENV", "development")
	jwtSecret := getEnv("JWT_SECRET", "fintrack-dev-secret-key-change-in-production-min32bytes!")
	expiryHours := getEnvInt("JWT_EXPIRY_HOURS", 24)

	return Config{
		Port:        port,
		DatabaseURL: dbURL,
		Env:         env,
		JWTSecret:   jwtSecret,
		JWTExpiry:   time.Duration(expiryHours) * time.Hour,
	}
}

// DSN returns the database connection string.
func (c Config) DSN() string {
	return c.DatabaseURL
}

// Addr returns the server listen address.
func (c Config) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
