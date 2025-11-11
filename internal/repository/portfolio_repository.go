package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/lenon/portfolios/internal/models"
)

// PortfolioRepository defines the interface for portfolio data operations
type PortfolioRepository interface {
	Create(portfolio *models.Portfolio) error
	FindByID(id string) (*models.Portfolio, error)
	FindByUserID(userID string) ([]*models.Portfolio, error)
	FindByUserIDAndName(userID, name string) (*models.Portfolio, error)
	Update(portfolio *models.Portfolio) error
	Delete(id string) error
	ExistsByUserIDAndName(userID, name string) (bool, error)
}

// portfolioRepository implements PortfolioRepository interface
type portfolioRepository struct {
	db *gorm.DB
}

// NewPortfolioRepository creates a new PortfolioRepository instance
func NewPortfolioRepository(db *gorm.DB) PortfolioRepository {
	return &portfolioRepository{db: db}
}

// Create creates a new portfolio in the database
func (r *portfolioRepository) Create(portfolio *models.Portfolio) error {
	if portfolio == nil {
		return fmt.Errorf("portfolio cannot be nil")
	}

	if err := r.db.Create(portfolio).Error; err != nil {
		return fmt.Errorf("failed to create portfolio: %w", err)
	}

	return nil
}

// FindByID finds a portfolio by ID
func (r *portfolioRepository) FindByID(id string) (*models.Portfolio, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	portfolioID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid portfolio ID format: %w", err)
	}

	var portfolio models.Portfolio
	err = r.db.Where("id = ?", portfolioID).First(&portfolio).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrPortfolioNotFound
		}
		return nil, fmt.Errorf("failed to find portfolio: %w", err)
	}

	return &portfolio, nil
}

// FindByUserID finds all portfolios for a specific user
func (r *portfolioRepository) FindByUserID(userID string) ([]*models.Portfolio, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	// Validate UUID format
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	var portfolios []*models.Portfolio
	err = r.db.Where("user_id = ?", uid).Order("created_at DESC").Find(&portfolios).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find portfolios for user: %w", err)
	}

	return portfolios, nil
}

// FindByUserIDAndName finds a portfolio by user ID and name
func (r *portfolioRepository) FindByUserIDAndName(userID, name string) (*models.Portfolio, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	// Validate UUID format
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	var portfolio models.Portfolio
	err = r.db.Where("user_id = ? AND name = ?", uid, name).First(&portfolio).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrPortfolioNotFound
		}
		return nil, fmt.Errorf("failed to find portfolio: %w", err)
	}

	return &portfolio, nil
}

// Update updates an existing portfolio
func (r *portfolioRepository) Update(portfolio *models.Portfolio) error {
	if portfolio == nil {
		return fmt.Errorf("portfolio cannot be nil")
	}

	result := r.db.Model(portfolio).Where("id = ?", portfolio.ID).Updates(portfolio)
	if result.Error != nil {
		return fmt.Errorf("failed to update portfolio: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return models.ErrPortfolioNotFound
	}

	return nil
}

// Delete deletes a portfolio by ID
func (r *portfolioRepository) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	portfolioID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid portfolio ID format: %w", err)
	}

	result := r.db.Where("id = ?", portfolioID).Delete(&models.Portfolio{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete portfolio: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return models.ErrPortfolioNotFound
	}

	return nil
}

// ExistsByUserIDAndName checks if a portfolio with the given name exists for a user
func (r *portfolioRepository) ExistsByUserIDAndName(userID, name string) (bool, error) {
	if userID == "" {
		return false, fmt.Errorf("user ID cannot be empty")
	}
	if name == "" {
		return false, fmt.Errorf("name cannot be empty")
	}

	// Validate UUID format
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user ID format: %w", err)
	}

	var count int64
	err = r.db.Model(&models.Portfolio{}).Where("user_id = ? AND name = ?", uid, name).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check portfolio existence: %w", err)
	}

	return count > 0, nil
}
