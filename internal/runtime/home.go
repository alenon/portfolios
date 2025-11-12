package runtime

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DefaultHomeDirName is the default directory name for the runtime home
	DefaultHomeDirName = ".portfolios"

	// ConfigFileName is the name of the runtime configuration file
	ConfigFileName = "config.yaml"

	// LogsDirName is the name of the logs directory
	LogsDirName = "logs"

	// ServerLogFileName is the name of the server log file
	ServerLogFileName = "server.log"

	// RequestLogFileName is the name of the request log file
	RequestLogFileName = "requests.log"
)

// HomeDir represents the runtime home directory structure
type HomeDir struct {
	Root       string // Root path of the home directory
	ConfigPath string // Full path to the config file
	LogsDir    string // Full path to the logs directory
	ServerLog  string // Full path to the server log file
	RequestLog string // Full path to the request log file
}

// InitHomeDir initializes the runtime home directory structure
// If homePath is empty, it defaults to ~/.portfolios
func InitHomeDir(homePath string) (*HomeDir, error) {
	// Determine home directory path
	if homePath == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		homePath = filepath.Join(userHome, DefaultHomeDirName)
	}

	// Create home directory structure
	home := &HomeDir{
		Root:       homePath,
		ConfigPath: filepath.Join(homePath, ConfigFileName),
		LogsDir:    filepath.Join(homePath, LogsDirName),
		ServerLog:  filepath.Join(homePath, LogsDirName, ServerLogFileName),
		RequestLog: filepath.Join(homePath, LogsDirName, RequestLogFileName),
	}

	// Create directories with appropriate permissions
	if err := home.createDirectories(); err != nil {
		return nil, err
	}

	return home, nil
}

// createDirectories creates all necessary directories for the home directory
func (h *HomeDir) createDirectories() error {
	// Create root directory (0755 - rwxr-xr-x)
	if err := os.MkdirAll(h.Root, 0755); err != nil {
		return fmt.Errorf("failed to create home directory %s: %w", h.Root, err)
	}

	// Create logs directory (0755 - rwxr-xr-x)
	if err := os.MkdirAll(h.LogsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory %s: %w", h.LogsDir, err)
	}

	return nil
}

// Exists checks if the home directory exists
func (h *HomeDir) Exists() bool {
	info, err := os.Stat(h.Root)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ConfigExists checks if the configuration file exists
func (h *HomeDir) ConfigExists() bool {
	info, err := os.Stat(h.ConfigPath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// EnsureLogFiles creates log files if they don't exist
func (h *HomeDir) EnsureLogFiles() error {
	// Create server log file if it doesn't exist (0600 - rw-------)
	if err := h.ensureFile(h.ServerLog); err != nil {
		return fmt.Errorf("failed to create server log file: %w", err)
	}

	// Create request log file if it doesn't exist (0600 - rw-------)
	if err := h.ensureFile(h.RequestLog); err != nil {
		return fmt.Errorf("failed to create request log file: %w", err)
	}

	return nil
}

// ensureFile creates a file if it doesn't exist
func (h *HomeDir) ensureFile(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); err == nil {
		return nil // File already exists
	} else if !os.IsNotExist(err) {
		return err // Some other error occurred
	}

	// Create the file with restricted permissions (0600 - rw-------)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	return file.Close()
}

// GetDefaultHomePath returns the default home directory path (~/.portfolios)
func GetDefaultHomePath() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(userHome, DefaultHomeDirName), nil
}
