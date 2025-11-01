package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&User{})
	require.NoError(t, err)

	return db
}

func TestUser_SetPassword(t *testing.T) {
	user := &User{
		Email: "test@example.com",
	}

	password := "SecurePass123"
	err := user.SetPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, password, user.PasswordHash, "Password should be hashed, not stored as plain text")
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{
		Email: "test@example.com",
	}

	password := "SecurePass123"
	err := user.SetPassword(password)
	require.NoError(t, err)

	// Test correct password
	assert.True(t, user.CheckPassword(password), "Correct password should return true")

	// Test incorrect password
	assert.False(t, user.CheckPassword("WrongPassword"), "Incorrect password should return false")
}

func TestUser_Create(t *testing.T) {
	db := setupTestDB(t)

	user := &User{
		Email: "test@example.com",
	}
	err := user.SetPassword("SecurePass123")
	require.NoError(t, err)

	// Create user
	result := db.Create(user)
	require.NoError(t, result.Error)

	// Verify UUID was generated
	assert.NotEqual(t, uuid.Nil, user.ID, "User ID should be generated")

	// Verify timestamps were set
	assert.False(t, user.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, user.UpdatedAt.IsZero(), "UpdatedAt should be set")
}

func TestUser_UniqueEmail(t *testing.T) {
	db := setupTestDB(t)

	email := "test@example.com"

	// Create first user
	user1 := &User{Email: email}
	err := user1.SetPassword("Password123")
	require.NoError(t, err)
	result := db.Create(user1)
	require.NoError(t, result.Error)

	// Try to create second user with same email
	user2 := &User{Email: email}
	err = user2.SetPassword("Password456")
	require.NoError(t, err)
	result = db.Create(user2)

	// Should fail due to unique constraint
	assert.Error(t, result.Error, "Creating user with duplicate email should fail")
}

func TestUser_UpdateLastLogin(t *testing.T) {
	db := setupTestDB(t)

	// Create user
	user := &User{Email: "test@example.com"}
	err := user.SetPassword("SecurePass123")
	require.NoError(t, err)
	result := db.Create(user)
	require.NoError(t, result.Error)

	// Initially, LastLoginAt should be nil
	assert.Nil(t, user.LastLoginAt)

	// Update last login
	err = user.UpdateLastLogin(db)
	require.NoError(t, err)

	// Verify LastLoginAt is set
	assert.NotNil(t, user.LastLoginAt)
	assert.WithinDuration(t, time.Now().UTC(), *user.LastLoginAt, 2*time.Second)
}

func TestUser_PasswordHashSecurity(t *testing.T) {
	user1 := &User{Email: "user1@example.com"}
	user2 := &User{Email: "user2@example.com"}

	// Same password should produce different hashes
	password := "SamePassword123"
	err1 := user1.SetPassword(password)
	err2 := user2.SetPassword(password)

	require.NoError(t, err1)
	require.NoError(t, err2)

	assert.NotEqual(t, user1.PasswordHash, user2.PasswordHash,
		"Same password should produce different hashes due to salt")

	// Both users should be able to verify the password
	assert.True(t, user1.CheckPassword(password))
	assert.True(t, user2.CheckPassword(password))
}
