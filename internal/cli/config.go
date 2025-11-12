package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the CLI configuration
type Config struct {
	APIBaseURL   string `mapstructure:"api_base_url"`
	AccessToken  string `mapstructure:"access_token"`
	RefreshToken string `mapstructure:"refresh_token"`
	OutputFormat string `mapstructure:"output_format"` // table, json, csv
}

// LoadConfig loads configuration from file
func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".portfolios")
	configFile := filepath.Join(configDir, "config.yaml")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set defaults
	viper.SetDefault("api_base_url", "http://localhost:8080")
	viper.SetDefault("output_format", "table")

	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	// Read config file if it exists, create if it doesn't
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found, create it with defaults
		if err := viper.SafeWriteConfig(); err != nil {
			// Try regular WriteConfig if SafeWriteConfig fails
			if writeErr := viper.WriteConfig(); writeErr != nil {
				return nil, fmt.Errorf("failed to create config file: %w", writeErr)
			}
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config) error {
	viper.Set("api_base_url", config.APIBaseURL)
	viper.Set("access_token", config.AccessToken)
	viper.Set("refresh_token", config.RefreshToken)
	viper.Set("output_format", config.OutputFormat)

	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// ClearTokens clears authentication tokens from config
func ClearTokens() error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	config.AccessToken = ""
	config.RefreshToken = ""

	return SaveConfig(config)
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".portfolios", "config.yaml"), nil
}
