package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestHealthCheck_NoConnection(t *testing.T) {
	// Reset DB
	DB = nil

	err := HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestHealthCheck_WithConnection(t *testing.T) {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	DB = db

	err = HealthCheck()
	assert.NoError(t, err)
}

func TestClose_NoConnection(t *testing.T) {
	DB = nil

	err := Close()
	assert.NoError(t, err) // Should not error if DB is nil
}

func TestClose_WithConnection(t *testing.T) {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	DB = db

	err = Close()
	assert.NoError(t, err)
}

func TestConnect_InvalidDSN(t *testing.T) {
	// Test with invalid DSN
	_, err := Connect("invalid-dsn")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

// Note: Testing successful Connect() would require a real PostgreSQL instance
// or more complex mocking. For unit tests, we test the error cases and
// the health check/close functions with SQLite as a stand-in.
