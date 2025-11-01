package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Success(t *testing.T) {
	// Set required environment variables
	_ = os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	_ = os.Setenv("JWT_SECRET", "test-secret-key-for-testing-purposes-only")
	_ = os.Setenv("SERVER_PORT", "9090")
	_ = os.Setenv("ENVIRONMENT", "test")
	_ = os.Setenv("JWT_ACCESS_TOKEN_DURATION", "15m")
	_ = os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")
	_ = os.Setenv("SMTP_HOST", "smtp.test.com")
	_ = os.Setenv("SMTP_PORT", "587")
	_ = os.Setenv("RATE_LIMIT_REQUESTS", "10")
	_ = os.Setenv("RATE_LIMIT_DURATION", "2m")

	defer func() {
		_ = os.Unsetenv("DATABASE_URL")
		_ = os.Unsetenv("JWT_SECRET")
		_ = os.Unsetenv("SERVER_PORT")
		_ = os.Unsetenv("ENVIRONMENT")
		_ = os.Unsetenv("JWT_ACCESS_TOKEN_DURATION")
		_ = os.Unsetenv("CORS_ALLOWED_ORIGINS")
		_ = os.Unsetenv("SMTP_HOST")
		_ = os.Unsetenv("SMTP_PORT")
		_ = os.Unsetenv("RATE_LIMIT_REQUESTS")
		_ = os.Unsetenv("RATE_LIMIT_DURATION")
	}()

	config, err := Load()

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "9090", config.Server.Port)
	assert.Equal(t, "test", config.Server.Environment)
	assert.Equal(t, "postgres://test:test@localhost:5432/test", config.Database.URL)
	assert.Equal(t, "test-secret-key-for-testing-purposes-only", config.JWT.Secret)
	assert.Equal(t, 15*time.Minute, config.JWT.AccessTokenDuration)
	assert.Equal(t, []string{"http://localhost:3000", "http://localhost:5173"}, config.Server.CORSOrigins)
	assert.Equal(t, "smtp.test.com", config.SMTP.Host)
	assert.Equal(t, 587, config.SMTP.Port)
	assert.Equal(t, 10, config.Security.RateLimitRequests)
	assert.Equal(t, 2*time.Minute, config.Security.RateLimitDuration)
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	_ = os.Unsetenv("DATABASE_URL")
	_ = os.Setenv("JWT_SECRET", "test-secret")
	defer func() { _ = os.Unsetenv("JWT_SECRET") }()

	config, err := Load()

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "DATABASE_URL")
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	_ = os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	_ = os.Unsetenv("JWT_SECRET")
	defer func() { _ = os.Unsetenv("DATABASE_URL") }()

	config, err := Load()

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

func TestGetEnv(t *testing.T) {
	_ = os.Setenv("TEST_VAR", "test_value")
	defer func() { _ = os.Unsetenv("TEST_VAR") }()

	value := getEnv("TEST_VAR", "default_value")
	assert.Equal(t, "test_value", value)
}

func TestGetEnv_Default(t *testing.T) {
	_ = os.Unsetenv("NON_EXISTENT_VAR")

	value := getEnv("NON_EXISTENT_VAR", "default_value")
	assert.Equal(t, "default_value", value)
}

func TestGetEnvAsInt(t *testing.T) {
	_ = os.Setenv("TEST_INT", "42")
	defer func() { _ = os.Unsetenv("TEST_INT") }()

	value := getEnvAsInt("TEST_INT", 10)
	assert.Equal(t, 42, value)
}

func TestGetEnvAsInt_Default(t *testing.T) {
	_ = os.Unsetenv("NON_EXISTENT_INT")

	value := getEnvAsInt("NON_EXISTENT_INT", 10)
	assert.Equal(t, 10, value)
}

func TestGetEnvAsInt_Invalid(t *testing.T) {
	_ = os.Setenv("INVALID_INT", "not_a_number")
	defer func() { _ = os.Unsetenv("INVALID_INT") }()

	value := getEnvAsInt("INVALID_INT", 10)
	assert.Equal(t, 10, value)
}

func TestGetEnvAsDuration(t *testing.T) {
	_ = os.Setenv("TEST_DURATION", "1h30m")
	defer func() { _ = os.Unsetenv("TEST_DURATION") }()

	duration := getEnvAsDuration("TEST_DURATION", 15*time.Minute)
	assert.Equal(t, 90*time.Minute, duration)
}

func TestGetEnvAsDuration_Default(t *testing.T) {
	_ = os.Unsetenv("NON_EXISTENT_DURATION")

	duration := getEnvAsDuration("NON_EXISTENT_DURATION", 15*time.Minute)
	assert.Equal(t, 15*time.Minute, duration)
}

func TestGetEnvAsDuration_Invalid(t *testing.T) {
	_ = os.Setenv("INVALID_DURATION", "invalid")
	defer func() { _ = os.Unsetenv("INVALID_DURATION") }()

	duration := getEnvAsDuration("INVALID_DURATION", 15*time.Minute)
	assert.Equal(t, 15*time.Minute, duration)
}

func TestGetEnvAsSlice(t *testing.T) {
	_ = os.Setenv("TEST_SLICE", "val1,val2,val3")
	defer func() { _ = os.Unsetenv("TEST_SLICE") }()

	slice := getEnvAsSlice("TEST_SLICE", []string{})
	assert.Equal(t, []string{"val1", "val2", "val3"}, slice)
}

func TestGetEnvAsSlice_Default(t *testing.T) {
	_ = os.Unsetenv("NON_EXISTENT_SLICE")

	slice := getEnvAsSlice("NON_EXISTENT_SLICE", []string{"default"})
	assert.Equal(t, []string{"default"}, slice)
}

func TestGetEnvAsSlice_Empty(t *testing.T) {
	_ = os.Setenv("EMPTY_SLICE", "")
	defer func() { _ = os.Unsetenv("EMPTY_SLICE") }()

	slice := getEnvAsSlice("EMPTY_SLICE", []string{"default"})
	assert.Equal(t, []string{"default"}, slice)
}

func TestGetEnvAsSlice_SingleValue(t *testing.T) {
	_ = os.Setenv("SINGLE_SLICE", "singlevalue")
	defer func() { _ = os.Unsetenv("SINGLE_SLICE") }()

	slice := getEnvAsSlice("SINGLE_SLICE", []string{})
	assert.Equal(t, []string{"singlevalue"}, slice)
}
