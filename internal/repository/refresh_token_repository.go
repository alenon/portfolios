package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/lenon/portfolios/internal/models"
)

// RefreshTokenRepository defines the interface for refresh token data operations
type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	FindByTokenHash(hash string) (*models.RefreshToken, error)
	RevokeByUserID(userID string) error
	RevokeByTokenHash(hash string) error
	DeleteExpired() error
}

// refreshTokenRepository implements RefreshTokenRepository interface
type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository instance
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// Create creates a new refresh token in the database
func (r *refreshTokenRepository) Create(token *models.RefreshToken) error {
	if token == nil {
		return fmt.Errorf("token cannot be nil")
	}

	if err := r.db.Create(token).Error; err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// FindByTokenHash finds a refresh token by its hash
func (r *refreshTokenRepository) FindByTokenHash(hash string) (*models.RefreshToken, error) {
	if hash == "" {
		return nil, fmt.Errorf("hash cannot be empty")
	}

	var token models.RefreshToken
	err := r.db.Where("token_hash = ?", hash).First(&token).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("refresh token not found")
		}
		return nil, fmt.Errorf("failed to find refresh token: %w", err)
	}

	return &token, nil
}

// RevokeByUserID revokes all refresh tokens for a specific user
func (r *refreshTokenRepository) RevokeByUserID(userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	// Validate UUID format
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	now := time.Now().UTC()
	result := r.db.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", uid).
		Update("revoked_at", now)

	if result.Error != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", result.Error)
	}

	return nil
}

// RevokeByTokenHash revokes a specific refresh token by its hash
func (r *refreshTokenRepository) RevokeByTokenHash(hash string) error {
	if hash == "" {
		return fmt.Errorf("hash cannot be empty")
	}

	now := time.Now().UTC()
	result := r.db.Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", hash).
		Update("revoked_at", now)

	if result.Error != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("refresh token not found or already revoked")
	}

	return nil
}

// DeleteExpired deletes all expired refresh tokens from the database
func (r *refreshTokenRepository) DeleteExpired() error {
	now := time.Now().UTC()
	result := r.db.Where("expires_at < ?", now).Delete(&models.RefreshToken{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete expired refresh tokens: %w", result.Error)
	}

	return nil
}
