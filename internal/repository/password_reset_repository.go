package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/lenon/portfolios/internal/models"
)

// PasswordResetRepository defines the interface for password reset token data operations
type PasswordResetRepository interface {
	Create(token *models.PasswordResetToken) error
	FindByTokenHash(hash string) (*models.PasswordResetToken, error)
	MarkAsUsed(id string) error
	DeleteExpired() error
}

// passwordResetRepository implements PasswordResetRepository interface
type passwordResetRepository struct {
	db *gorm.DB
}

// NewPasswordResetRepository creates a new PasswordResetRepository instance
func NewPasswordResetRepository(db *gorm.DB) PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

// Create creates a new password reset token in the database
func (r *passwordResetRepository) Create(token *models.PasswordResetToken) error {
	if token == nil {
		return fmt.Errorf("token cannot be nil")
	}

	if err := r.db.Create(token).Error; err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	return nil
}

// FindByTokenHash finds a password reset token by its hash
func (r *passwordResetRepository) FindByTokenHash(hash string) (*models.PasswordResetToken, error) {
	if hash == "" {
		return nil, fmt.Errorf("hash cannot be empty")
	}

	var token models.PasswordResetToken
	err := r.db.Where("token_hash = ?", hash).First(&token).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("password reset token not found")
		}
		return nil, fmt.Errorf("failed to find password reset token: %w", err)
	}

	return &token, nil
}

// MarkAsUsed marks a password reset token as used
func (r *passwordResetRepository) MarkAsUsed(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	now := time.Now().UTC()
	result := r.db.Model(&models.PasswordResetToken{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", now)

	if result.Error != nil {
		return fmt.Errorf("failed to mark token as used: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("password reset token not found or already used")
	}

	return nil
}

// DeleteExpired deletes all expired password reset tokens from the database
func (r *passwordResetRepository) DeleteExpired() error {
	now := time.Now().UTC()
	result := r.db.Where("expires_at < ?", now).Delete(&models.PasswordResetToken{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete expired password reset tokens: %w", result.Error)
	}

	return nil
}
