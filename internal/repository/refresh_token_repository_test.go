package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRefreshTokenTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	err = db.AutoMigrate(&models.RefreshToken{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

func TestNewRefreshTokenRepository(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	assert.NotNil(t, repo)
}

func TestRefreshTokenRepository_Create_Success(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	token := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "token-hash-123",
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	err := repo.Create(token)

	assert.NoError(t, err)
}

func TestRefreshTokenRepository_Create_NilToken(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	err := repo.Create(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token cannot be nil")
}

func TestRefreshTokenRepository_FindByTokenHash_Success(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	token := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "token-hash-123",
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	err := repo.Create(token)
	assert.NoError(t, err)

	foundToken, err := repo.FindByTokenHash("token-hash-123")

	assert.NoError(t, err)
	assert.NotNil(t, foundToken)
	assert.Equal(t, token.TokenHash, foundToken.TokenHash)
	assert.Equal(t, token.UserID, foundToken.UserID)
}

func TestRefreshTokenRepository_FindByTokenHash_EmptyHash(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	token, err := repo.FindByTokenHash("")

	assert.Error(t, err)
	assert.Nil(t, token)
	assert.Contains(t, err.Error(), "hash cannot be empty")
}

func TestRefreshTokenRepository_FindByTokenHash_NotFound(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	token, err := repo.FindByTokenHash("nonexistent-hash")

	assert.Error(t, err)
	assert.Nil(t, token)
	assert.Contains(t, err.Error(), "refresh token not found")
}

func TestRefreshTokenRepository_RevokeByUserID_Success(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	userID := uuid.New()
	token1 := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: "token-hash-1",
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	token2 := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: "token-hash-2",
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	err := repo.Create(token1)
	assert.NoError(t, err)
	err = repo.Create(token2)
	assert.NoError(t, err)

	err = repo.RevokeByUserID(userID.String())

	assert.NoError(t, err)

	// Verify tokens were revoked
	foundToken1, _ := repo.FindByTokenHash("token-hash-1")
	foundToken2, _ := repo.FindByTokenHash("token-hash-2")
	assert.NotNil(t, foundToken1.RevokedAt)
	assert.NotNil(t, foundToken2.RevokedAt)
}

func TestRefreshTokenRepository_RevokeByUserID_EmptyUserID(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	err := repo.RevokeByUserID("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID cannot be empty")
}

func TestRefreshTokenRepository_RevokeByUserID_InvalidUUID(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	err := repo.RevokeByUserID("invalid-uuid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID format")
}

func TestRefreshTokenRepository_RevokeByUserID_NoTokens(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	err := repo.RevokeByUserID(uuid.New().String())

	// Should not error even if no tokens found
	assert.NoError(t, err)
}

func TestRefreshTokenRepository_RevokeByTokenHash_Success(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	token := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "token-hash-123",
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	err := repo.Create(token)
	assert.NoError(t, err)

	err = repo.RevokeByTokenHash("token-hash-123")

	assert.NoError(t, err)

	// Verify token was revoked
	foundToken, _ := repo.FindByTokenHash("token-hash-123")
	assert.NotNil(t, foundToken.RevokedAt)
}

func TestRefreshTokenRepository_RevokeByTokenHash_EmptyHash(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	err := repo.RevokeByTokenHash("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hash cannot be empty")
}

func TestRefreshTokenRepository_RevokeByTokenHash_NotFound(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	err := repo.RevokeByTokenHash("nonexistent-hash")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh token not found or already revoked")
}

func TestRefreshTokenRepository_RevokeByTokenHash_AlreadyRevoked(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	token := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "token-hash-123",
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	err := repo.Create(token)
	assert.NoError(t, err)

	// Revoke once
	err = repo.RevokeByTokenHash("token-hash-123")
	assert.NoError(t, err)

	// Try to revoke again
	err = repo.RevokeByTokenHash("token-hash-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh token not found or already revoked")
}

func TestRefreshTokenRepository_DeleteExpired_Success(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	// Create expired token
	expiredToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "expired-token",
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
		CreatedAt: time.Now().UTC().Add(-2 * time.Hour),
	}
	err := repo.Create(expiredToken)
	assert.NoError(t, err)

	// Create valid token
	validToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "valid-token",
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	err = repo.Create(validToken)
	assert.NoError(t, err)

	err = repo.DeleteExpired()

	assert.NoError(t, err)

	// Verify expired token was deleted
	_, err = repo.FindByTokenHash("expired-token")
	assert.Error(t, err)

	// Verify valid token still exists
	foundToken, err := repo.FindByTokenHash("valid-token")
	assert.NoError(t, err)
	assert.NotNil(t, foundToken)
}

func TestRefreshTokenRepository_DeleteExpired_NoExpiredTokens(t *testing.T) {
	db := setupRefreshTokenTestDB(t)
	repo := NewRefreshTokenRepository(db)

	// Create only valid tokens
	token := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "valid-token",
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	err := repo.Create(token)
	assert.NoError(t, err)

	err = repo.DeleteExpired()

	// Should not error even if no expired tokens
	assert.NoError(t, err)

	// Verify token still exists
	foundToken, err := repo.FindByTokenHash("valid-token")
	assert.NoError(t, err)
	assert.NotNil(t, foundToken)
}
