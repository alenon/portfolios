package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	JWT        JWTConfig        `yaml:"jwt"`
	SMTP       SMTPConfig       `yaml:"smtp"`
	Security   SecurityConfig   `yaml:"security"`
	MarketData MarketDataConfig `yaml:"market_data"`
	Runtime    RuntimeConfig    `yaml:"runtime"`
	Logging    LoggingConfig    `yaml:"logging"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port        string   `yaml:"port"`
	Environment string   `yaml:"environment"`
	CORSOrigins []string `yaml:"cors_origins"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	URL string `yaml:"url"`
}

// JWTConfig holds JWT token configuration
type JWTConfig struct {
	Secret                    string        `yaml:"secret"`
	AccessTokenDuration       time.Duration `yaml:"access_token_duration"`
	RefreshTokenDuration      time.Duration `yaml:"refresh_token_duration"`
	RememberMeAccessDuration  time.Duration `yaml:"remember_me_access_duration"`
	RememberMeRefreshDuration time.Duration `yaml:"remember_me_refresh_duration"`
}

// SMTPConfig holds email service configuration
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	RateLimitRequests int           `yaml:"rate_limit_requests"`
	RateLimitDuration time.Duration `yaml:"rate_limit_duration"`
}

// MarketDataConfig holds market data provider configuration
type MarketDataConfig struct {
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key"`
}

// RuntimeConfig holds runtime directory configuration
type RuntimeConfig struct {
	HomeDir string `yaml:"home_dir"` // Path to runtime home directory
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level          string `yaml:"level"`          // debug, info, warn, error
	Format         string `yaml:"format"`         // json, console
	ServerLogPath  string `yaml:"server_log"`     // Path to server log file
	RequestLogPath string `yaml:"request_log"`    // Path to request log file
	EnableConsole  bool   `yaml:"enable_console"` // Enable console output
	EnableFile     bool   `yaml:"enable_file"`    // Enable file output
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	return LoadWithYAML("")
}

// LoadWithYAML reads configuration from a YAML file and environment variables
// Environment variables take precedence over YAML file values
func LoadWithYAML(yamlPath string) (*Config, error) {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	// Initialize with defaults
	config := &Config{
		Server: ServerConfig{
			Port:        "8080",
			Environment: "development",
			CORSOrigins: []string{"http://localhost:5173"},
		},
		JWT: JWTConfig{
			AccessTokenDuration:       30 * time.Minute,
			RefreshTokenDuration:      7 * 24 * time.Hour,
			RememberMeAccessDuration:  24 * time.Hour,
			RememberMeRefreshDuration: 30 * 24 * time.Hour,
		},
		SMTP: SMTPConfig{
			Port: 587,
		},
		Security: SecurityConfig{
			RateLimitRequests: 5,
			RateLimitDuration: 1 * time.Minute,
		},
		MarketData: MarketDataConfig{
			Provider: "alphavantage",
		},
		Logging: LoggingConfig{
			Level:         "info",
			Format:        "json",
			EnableConsole: true,
			EnableFile:    false,
		},
	}

	// Load from YAML file if provided
	if yamlPath != "" {
		if err := loadFromYAML(yamlPath, config); err != nil {
			return nil, fmt.Errorf("failed to load YAML config: %w", err)
		}
	}

	// Override with environment variables (env vars take precedence)
	applyEnvironmentOverrides(config)

	// Validate required fields
	if config.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if config.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return config, nil
}

// loadFromYAML loads configuration from a YAML file
func loadFromYAML(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, skip loading
			return nil
		}
		return err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	return nil
}

// applyEnvironmentOverrides applies environment variable overrides to the config
func applyEnvironmentOverrides(config *Config) {
	// Server config
	if val := getEnv("SERVER_PORT", ""); val != "" {
		config.Server.Port = val
	}
	if val := getEnv("ENVIRONMENT", ""); val != "" {
		config.Server.Environment = val
	}
	if val := getEnvAsSlice("CORS_ALLOWED_ORIGINS", nil); val != nil {
		config.Server.CORSOrigins = val
	}

	// Database config
	if val := getEnv("DATABASE_URL", ""); val != "" {
		config.Database.URL = val
	}

	// JWT config
	if val := getEnv("JWT_SECRET", ""); val != "" {
		config.JWT.Secret = val
	}
	if val := getEnvAsDuration("JWT_ACCESS_TOKEN_DURATION", 0); val != 0 {
		config.JWT.AccessTokenDuration = val
	}
	if val := getEnvAsDuration("JWT_REFRESH_TOKEN_DURATION", 0); val != 0 {
		config.JWT.RefreshTokenDuration = val
	}
	if val := getEnvAsDuration("JWT_REMEMBER_ME_ACCESS_DURATION", 0); val != 0 {
		config.JWT.RememberMeAccessDuration = val
	}
	if val := getEnvAsDuration("JWT_REMEMBER_ME_REFRESH_DURATION", 0); val != 0 {
		config.JWT.RememberMeRefreshDuration = val
	}

	// SMTP config
	if val := getEnv("SMTP_HOST", ""); val != "" {
		config.SMTP.Host = val
	}
	if val := getEnvAsInt("SMTP_PORT", 0); val != 0 {
		config.SMTP.Port = val
	}
	if val := getEnv("SMTP_USERNAME", ""); val != "" {
		config.SMTP.Username = val
	}
	if val := getEnv("SMTP_PASSWORD", ""); val != "" {
		config.SMTP.Password = val
	}
	if val := getEnv("SMTP_FROM", ""); val != "" {
		config.SMTP.From = val
	}

	// Security config
	if val := getEnvAsInt("RATE_LIMIT_REQUESTS", 0); val != 0 {
		config.Security.RateLimitRequests = val
	}
	if val := getEnvAsDuration("RATE_LIMIT_DURATION", 0); val != 0 {
		config.Security.RateLimitDuration = val
	}

	// Market data config
	if val := getEnv("MARKET_DATA_PROVIDER", ""); val != "" {
		config.MarketData.Provider = val
	}
	if val := getEnv("MARKET_DATA_API_KEY", ""); val != "" {
		config.MarketData.APIKey = val
	}

	// Runtime config
	if val := getEnv("RUNTIME_HOME_DIR", ""); val != "" {
		config.Runtime.HomeDir = val
	}

	// Logging config
	if val := getEnv("LOG_LEVEL", ""); val != "" {
		config.Logging.Level = val
	}
	if val := getEnv("LOG_FORMAT", ""); val != "" {
		config.Logging.Format = val
	}
	if val := getEnv("LOG_SERVER_PATH", ""); val != "" {
		config.Logging.ServerLogPath = val
	}
	if val := getEnv("LOG_REQUEST_PATH", ""); val != "" {
		config.Logging.RequestLogPath = val
	}
	if val := getEnvAsBool("LOG_ENABLE_CONSOLE", false); val {
		config.Logging.EnableConsole = val
	}
	if val := getEnvAsBool("LOG_ENABLE_FILE", false); val {
		config.Logging.EnableFile = val
	}
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

// getEnvAsBool retrieves an environment variable as a boolean or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	// Parse boolean values
	switch valueStr {
	case "true", "True", "TRUE", "1", "yes", "Yes", "YES":
		return true
	case "false", "False", "FALSE", "0", "no", "No", "NO":
		return false
	default:
		return defaultValue
	}
}
