package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/lenon/portfolios/internal/models"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id string) (*models.User, error)
	UpdateLastLogin(id string) error
	UpdatePassword(id string, passwordHash string) error
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user in the database
func (r *userRepository) Create(user *models.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByEmail finds a user by email address
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return &user, nil
}

// FindByID finds a user by ID
func (r *userRepository) FindByID(id string) (*models.User, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	var user models.User
	err = r.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with id: %s", id)
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return &user, nil
}

// UpdateLastLogin updates the last login timestamp for a user
func (r *userRepository) UpdateLastLogin(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	userID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	now := time.Now().UTC()
	result := r.db.Model(&models.User{}).Where("id = ?", userID).Update("last_login_at", now)
	if result.Error != nil {
		return fmt.Errorf("failed to update last login: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found with id: %s", id)
	}

	return nil
}

// UpdatePassword updates the password hash for a user
func (r *userRepository) UpdatePassword(id string, passwordHash string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}
	if passwordHash == "" {
		return fmt.Errorf("password hash cannot be empty")
	}

	// Validate UUID format
	userID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	now := time.Now().UTC()
	result := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"password_hash": passwordHash,
			"updated_at":    now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update password: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found with id: %s", id)
	}

	return nil
}
