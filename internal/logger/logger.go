package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger interface for structured logging
type Logger interface {
	Debug() *zerolog.Event
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
	Fatal() *zerolog.Event
	With() zerolog.Context
}

// Config holds logger configuration
type Config struct {
	Level      string // debug, info, warn, error
	Format     string // json, console
	OutputPath string // stdout, stderr, or file path
}

// AppLogger wraps zerolog.Logger
type AppLogger struct {
	logger zerolog.Logger
}

// NewLogger creates a new structured logger
func NewLogger(config Config) *AppLogger {
	// Set log level
	level := parseLevel(config.Level)
	zerolog.SetGlobalLevel(level)

	// Configure output
	var output io.Writer
	switch config.OutputPath {
	case "stdout", "":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// File output with secure permissions (read/write for owner only)
		file, err := os.OpenFile(config.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			log.Fatal().Err(err).Str("path", config.OutputPath).Msg("Failed to open log file")
		}
		output = file
	}

	// Configure format
	var logger zerolog.Logger
	if config.Format == "console" {
		// Pretty console output for development
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
		}
		logger = zerolog.New(output).With().Timestamp().Caller().Logger()
	} else {
		// JSON output for production
		logger = zerolog.New(output).With().Timestamp().Caller().Logger()
	}

	return &AppLogger{logger: logger}
}

// Debug returns a debug level event
func (l *AppLogger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

// Info returns an info level event
func (l *AppLogger) Info() *zerolog.Event {
	return l.logger.Info()
}

// Warn returns a warn level event
func (l *AppLogger) Warn() *zerolog.Event {
	return l.logger.Warn()
}

// Error returns an error level event
func (l *AppLogger) Error() *zerolog.Event {
	return l.logger.Error()
}

// Fatal returns a fatal level event
func (l *AppLogger) Fatal() *zerolog.Event {
	return l.logger.Fatal()
}

// With returns a new context for adding fields
func (l *AppLogger) With() zerolog.Context {
	return l.logger.With()
}

// parseLevel converts string level to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// LogAuthEvent logs authentication-related events
func (l *AppLogger) LogAuthEvent(event string, userID string, email string, success bool, reason string) {
	l.Info().
		Str("event", event).
		Str("user_id", userID).
		Str("email", email).
		Bool("success", success).
		Str("reason", reason).
		Msg("Authentication event")
}

// LogAPIRequest logs API request details
func (l *AppLogger) LogAPIRequest(method string, path string, statusCode int, duration time.Duration, userID string) {
	l.Info().
		Str("method", method).
		Str("path", path).
		Int("status_code", statusCode).
		Dur("duration_ms", duration).
		Str("user_id", userID).
		Msg("API request")
}

// LogError logs an error with context
func (l *AppLogger) LogError(err error, message string, fields map[string]interface{}) {
	event := l.Error().Err(err).Str("message", message)
	for key, value := range fields {
		event = event.Interface(key, value)
	}
	event.Msg("Error occurred")
}

// LogDatabaseOperation logs database operations
func (l *AppLogger) LogDatabaseOperation(operation string, table string, duration time.Duration, err error) {
	if err != nil {
		l.Error().
			Err(err).
			Str("operation", operation).
			Str("table", table).
			Dur("duration_ms", duration).
			Msg("Database operation failed")
	} else {
		l.Debug().
			Str("operation", operation).
			Str("table", table).
			Dur("duration_ms", duration).
			Msg("Database operation completed")
	}
}

// Global logger instance (initialized in main)
var globalLogger *AppLogger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(config Config) {
	globalLogger = NewLogger(config)
}

// GetLogger returns the global logger instance
func GetLogger() *AppLogger {
	if globalLogger == nil {
		// Fallback to default logger if not initialized
		globalLogger = NewLogger(Config{
			Level:  "info",
			Format: "json",
		})
	}
	return globalLogger
}
