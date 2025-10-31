package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	SMTP     SMTPConfig
	Security SecurityConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port        string
	Environment string
	CORSOrigins []string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	URL string
}

// JWTConfig holds JWT token configuration
type JWTConfig struct {
	Secret                   string
	AccessTokenDuration      time.Duration
	RefreshTokenDuration     time.Duration
	RememberMeAccessDuration time.Duration
	RememberMeRefreshDuration time.Duration
}

// SMTPConfig holds email service configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	RateLimitRequests int
	RateLimitDuration time.Duration
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:        getEnv("SERVER_PORT", "8080"),
			Environment: getEnv("ENVIRONMENT", "development"),
			CORSOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", ""),
		},
		JWT: JWTConfig{
			Secret:                   getEnv("JWT_SECRET", ""),
			AccessTokenDuration:      getEnvAsDuration("JWT_ACCESS_TOKEN_DURATION", 30*time.Minute),
			RefreshTokenDuration:     getEnvAsDuration("JWT_REFRESH_TOKEN_DURATION", 7*24*time.Hour),
			RememberMeAccessDuration: getEnvAsDuration("JWT_REMEMBER_ME_ACCESS_DURATION", 24*time.Hour),
			RememberMeRefreshDuration: getEnvAsDuration("JWT_REMEMBER_ME_REFRESH_DURATION", 30*24*time.Hour),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getEnvAsInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
		},
		Security: SecurityConfig{
			RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 5),
			RateLimitDuration: getEnvAsDuration("RATE_LIMIT_DURATION", 1*time.Minute),
		},
	}

	// Validate required fields
	if config.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if config.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return config, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsDuration retrieves an environment variable as a duration or returns a default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsSlice retrieves an environment variable as a slice or returns a default value
func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	// Simple comma-separated parsing
	var result []string
	current := ""
	for _, char := range valueStr {
		if char == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
