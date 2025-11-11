package logger

import (
	"bytes"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "json format to stdout",
			config: Config{
				Level:      "info",
				Format:     "json",
				OutputPath: "stdout",
			},
		},
		{
			name: "console format to stderr",
			config: Config{
				Level:      "debug",
				Format:     "console",
				OutputPath: "stderr",
			},
		},
		{
			name: "default config",
			config: Config{
				Level:  "",
				Format: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.config)
			assert.NotNil(t, logger)
		})
	}
}

func TestAppLogger_LogLevels(t *testing.T) {
	config := Config{
		Level:      "debug",
		Format:     "json",
		OutputPath: "stdout",
	}
	logger := NewLogger(config)

	// Verify that calling log methods doesn't panic
	// We don't check event details as that's zerolog's responsibility
	assert.NotPanics(t, func() {
		logger.Debug()
		logger.Info()
		logger.Warn()
		logger.Error()
		logger.With()
	})
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		level    string
		expected zerolog.Level
	}{
		{"debug", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"fatal", zerolog.FatalLevel},
		{"invalid", zerolog.InfoLevel}, // defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			result := parseLevel(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppLogger_LogAuthEvent(t *testing.T) {
	var buf bytes.Buffer
	zlog := zerolog.New(&buf)
	logger := &AppLogger{logger: zlog}

	logger.LogAuthEvent("login", "user123", "test@example.com", true, "success")

	assert.Contains(t, buf.String(), "login")
	assert.Contains(t, buf.String(), "user123")
	assert.Contains(t, buf.String(), "test@example.com")
}

func TestAppLogger_LogAPIRequest(t *testing.T) {
	var buf bytes.Buffer
	zlog := zerolog.New(&buf)
	logger := &AppLogger{logger: zlog}

	logger.LogAPIRequest("GET", "/api/portfolios", 200, 100*time.Millisecond, "user123")

	assert.Contains(t, buf.String(), "GET")
	assert.Contains(t, buf.String(), "/api/portfolios")
	assert.Contains(t, buf.String(), "200")
}

func TestAppLogger_LogError(t *testing.T) {
	var buf bytes.Buffer
	zlog := zerolog.New(&buf)
	logger := &AppLogger{logger: zlog}

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	logger.LogError(assert.AnError, "test error", fields)

	assert.Contains(t, buf.String(), "test error")
}

func TestAppLogger_LogDatabaseOperation(t *testing.T) {
	t.Run("successful operation", func(t *testing.T) {
		var buf bytes.Buffer
		zlog := zerolog.New(&buf).Level(zerolog.DebugLevel)
		logger := &AppLogger{logger: zlog}

		// Just verify the method doesn't panic
		logger.LogDatabaseOperation("SELECT", "users", 50*time.Millisecond, nil)
		assert.True(t, true) // If we got here, test passed
	})

	t.Run("failed operation", func(t *testing.T) {
		var buf bytes.Buffer
		zlog := zerolog.New(&buf).Level(zerolog.DebugLevel)
		logger := &AppLogger{logger: zlog}

		// Just verify the method doesn't panic
		logger.LogDatabaseOperation("INSERT", "users", 50*time.Millisecond, assert.AnError)
		assert.True(t, true) // If we got here, test passed
	})
}

func TestInitGlobalLogger(t *testing.T) {
	config := Config{
		Level:      "debug",
		Format:     "json",
		OutputPath: "stdout",
	}

	InitGlobalLogger(config)

	logger := GetLogger()
	assert.NotNil(t, logger)
}

func TestGetLogger_Fallback(t *testing.T) {
	// Reset global logger
	globalLogger = nil

	logger := GetLogger()
	assert.NotNil(t, logger)
}
