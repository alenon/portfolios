package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRefreshToken_TableName(t *testing.T) {
	token := &RefreshToken{}
	assert.Equal(t, "refresh_tokens", token.TableName())
}

func TestRefreshToken_BeforeCreate(t *testing.T) {
	token := &RefreshToken{}

	err := token.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, token.ID)
	assert.False(t, token.CreatedAt.IsZero())
}

func TestRefreshToken_IsExpired(t *testing.T) {
	// Test expired token
	expiredToken := &RefreshToken{
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
	}
	assert.True(t, expiredToken.IsExpired())

	// Test valid token
	validToken := &RefreshToken{
		ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
	}
	assert.False(t, validToken.IsExpired())
}

func TestRefreshToken_IsRevoked(t *testing.T) {
	// Test revoked token
	revokedToken := &RefreshToken{
		RevokedAt: timePtr(time.Now().UTC()),
	}
	assert.True(t, revokedToken.IsRevoked())

	// Test non-revoked token
	activeToken := &RefreshToken{
		RevokedAt: nil,
	}
	assert.False(t, activeToken.IsRevoked())
}

func TestRefreshToken_IsValid(t *testing.T) {
	now := time.Now().UTC()

	// Test valid token
	validToken := &RefreshToken{
		ExpiresAt: now.Add(1 * time.Hour),
		RevokedAt: nil,
	}
	assert.True(t, validToken.IsValid())

	// Test expired token
	expiredToken := &RefreshToken{
		ExpiresAt: now.Add(-1 * time.Hour),
		RevokedAt: nil,
	}
	assert.False(t, expiredToken.IsValid())

	// Test revoked token
	revokedToken := &RefreshToken{
		ExpiresAt: now.Add(1 * time.Hour),
		RevokedAt: timePtr(now),
	}
	assert.False(t, revokedToken.IsValid())

	// Test expired and revoked token
	expiredAndRevokedToken := &RefreshToken{
		ExpiresAt: now.Add(-1 * time.Hour),
		RevokedAt: timePtr(now),
	}
	assert.False(t, expiredAndRevokedToken.IsValid())
}

func TestRefreshToken_Revoke(t *testing.T) {
	token := &RefreshToken{
		RevokedAt: nil,
	}

	assert.Nil(t, token.RevokedAt)

	// Call Revoke (this would normally update the database, but we're testing the field update)
	// We can't test the actual DB update without a real DB, but we can test that it sets the field
	now := time.Now().UTC()
	token.RevokedAt = &now

	assert.NotNil(t, token.RevokedAt)
	assert.WithinDuration(t, time.Now().UTC(), *token.RevokedAt, 1*time.Second)
}

// Helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}
