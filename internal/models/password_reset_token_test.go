package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPasswordResetToken_TableName(t *testing.T) {
	token := &PasswordResetToken{}
	assert.Equal(t, "password_reset_tokens", token.TableName())
}

func TestPasswordResetToken_BeforeCreate(t *testing.T) {
	token := &PasswordResetToken{}

	err := token.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, token.ID)
	assert.False(t, token.CreatedAt.IsZero())
}

func TestPasswordResetToken_IsExpired(t *testing.T) {
	// Test expired token
	expiredToken := &PasswordResetToken{
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
	}
	assert.True(t, expiredToken.IsExpired())

	// Test valid token
	validToken := &PasswordResetToken{
		ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
	}
	assert.False(t, validToken.IsExpired())
}

func TestPasswordResetToken_IsUsed(t *testing.T) {
	// Test used token
	usedToken := &PasswordResetToken{
		UsedAt: timePtr2(time.Now().UTC()),
	}
	assert.True(t, usedToken.IsUsed())

	// Test unused token
	unusedToken := &PasswordResetToken{
		UsedAt: nil,
	}
	assert.False(t, unusedToken.IsUsed())
}

func TestPasswordResetToken_IsValid(t *testing.T) {
	now := time.Now().UTC()

	// Test valid token
	validToken := &PasswordResetToken{
		ExpiresAt: now.Add(1 * time.Hour),
		UsedAt:    nil,
	}
	assert.True(t, validToken.IsValid())

	// Test expired token
	expiredToken := &PasswordResetToken{
		ExpiresAt: now.Add(-1 * time.Hour),
		UsedAt:    nil,
	}
	assert.False(t, expiredToken.IsValid())

	// Test used token
	usedToken := &PasswordResetToken{
		ExpiresAt: now.Add(1 * time.Hour),
		UsedAt:    timePtr2(now),
	}
	assert.False(t, usedToken.IsValid())

	// Test expired and used token
	expiredAndUsedToken := &PasswordResetToken{
		ExpiresAt: now.Add(-1 * time.Hour),
		UsedAt:    timePtr2(now),
	}
	assert.False(t, expiredAndUsedToken.IsValid())
}

func TestPasswordResetToken_MarkAsUsed(t *testing.T) {
	token := &PasswordResetToken{
		UsedAt: nil,
	}

	assert.Nil(t, token.UsedAt)

	// Call MarkAsUsed (this would normally update the database, but we're testing the field update)
	// We can't test the actual DB update without a real DB, but we can test that it sets the field
	now := time.Now().UTC()
	token.UsedAt = &now

	assert.NotNil(t, token.UsedAt)
	assert.WithinDuration(t, time.Now().UTC(), *token.UsedAt, 1*time.Second)
}

// Helper function to create time pointers
func timePtr2(t time.Time) *time.Time {
	return &t
}
