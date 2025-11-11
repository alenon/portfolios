package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/lenon/portfolios/internal/models"
)

// HoldingRepository defines the interface for holding data operations
type HoldingRepository interface {
	Create(holding *models.Holding) error
	FindByID(id string) (*models.Holding, error)
	FindByPortfolioID(portfolioID string) ([]*models.Holding, error)
	FindByPortfolioIDAndSymbol(portfolioID, symbol string) (*models.Holding, error)
	FindBySymbol(symbol string) ([]*models.Holding, error)
	Update(holding *models.Holding) error
	Upsert(holding *models.Holding) error
	Delete(id string) error
	DeleteByPortfolioIDAndSymbol(portfolioID, symbol string) error
}

// holdingRepository implements HoldingRepository interface
type holdingRepository struct {
	db *gorm.DB
}

// NewHoldingRepository creates a new HoldingRepository instance
func NewHoldingRepository(db *gorm.DB) HoldingRepository {
	return &holdingRepository{db: db}
}

// Create creates a new holding in the database
func (r *holdingRepository) Create(holding *models.Holding) error {
	if holding == nil {
		return fmt.Errorf("holding cannot be nil")
	}

	if err := r.db.Create(holding).Error; err != nil {
		return fmt.Errorf("failed to create holding: %w", err)
	}

	return nil
}

// FindByID finds a holding by ID
func (r *holdingRepository) FindByID(id string) (*models.Holding, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	holdingID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid holding ID format: %w", err)
	}

	var holding models.Holding
	err = r.db.Where("id = ?", holdingID).First(&holding).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrHoldingNotFound
		}
		return nil, fmt.Errorf("failed to find holding: %w", err)
	}

	return &holding, nil
}

// FindByPortfolioID finds all holdings for a specific portfolio
func (r *holdingRepository) FindByPortfolioID(portfolioID string) ([]*models.Holding, error) {
	if portfolioID == "" {
		return nil, fmt.Errorf("portfolio ID cannot be empty")
	}

	// Validate UUID format
	pid, err := uuid.Parse(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("invalid portfolio ID format: %w", err)
	}

	var holdings []*models.Holding
	err = r.db.Where("portfolio_id = ? AND quantity > 0", pid).
		Order("symbol ASC").
		Find(&holdings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find holdings: %w", err)
	}

	return holdings, nil
}

// FindByPortfolioIDAndSymbol finds a holding by portfolio ID and symbol
func (r *holdingRepository) FindByPortfolioIDAndSymbol(portfolioID, symbol string) (*models.Holding, error) {
	if portfolioID == "" {
		return nil, fmt.Errorf("portfolio ID cannot be empty")
	}
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	// Validate UUID format
	pid, err := uuid.Parse(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("invalid portfolio ID format: %w", err)
	}

	var holding models.Holding
	err = r.db.Where("portfolio_id = ? AND symbol = ?", pid, symbol).First(&holding).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrHoldingNotFound
		}
		return nil, fmt.Errorf("failed to find holding: %w", err)
	}

	return &holding, nil
}

// FindBySymbol finds all holdings for a given symbol across all portfolios
func (r *holdingRepository) FindBySymbol(symbol string) ([]*models.Holding, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	var holdings []*models.Holding
	if err := r.db.Preload("Portfolio").
		Where("symbol = ?", symbol).
		Find(&holdings).Error; err != nil {
		return nil, fmt.Errorf("failed to find holdings: %w", err)
	}

	return holdings, nil
}

// Update updates an existing holding
func (r *holdingRepository) Update(holding *models.Holding) error {
	if holding == nil {
		return fmt.Errorf("holding cannot be nil")
	}

	result := r.db.Model(holding).Where("id = ?", holding.ID).Updates(holding)
	if result.Error != nil {
		return fmt.Errorf("failed to update holding: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return models.ErrHoldingNotFound
	}

	return nil
}

// Upsert creates or updates a holding based on portfolio_id and symbol
func (r *holdingRepository) Upsert(holding *models.Holding) error {
	if holding == nil {
		return fmt.Errorf("holding cannot be nil")
	}

	// Try to find existing holding
	existing, err := r.FindByPortfolioIDAndSymbol(holding.PortfolioID.String(), holding.Symbol)
	if err != nil && err != models.ErrHoldingNotFound {
		return fmt.Errorf("failed to check existing holding: %w", err)
	}

	if existing != nil {
		// Update existing holding
		holding.ID = existing.ID
		return r.Update(holding)
	}

	// Create new holding
	return r.Create(holding)
}

// Delete deletes a holding by ID
func (r *holdingRepository) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	holdingID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid holding ID format: %w", err)
	}

	result := r.db.Where("id = ?", holdingID).Delete(&models.Holding{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete holding: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return models.ErrHoldingNotFound
	}

	return nil
}

// DeleteByPortfolioIDAndSymbol deletes a holding by portfolio ID and symbol
func (r *holdingRepository) DeleteByPortfolioIDAndSymbol(portfolioID, symbol string) error {
	if portfolioID == "" {
		return fmt.Errorf("portfolio ID cannot be empty")
	}
	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	// Validate UUID format
	pid, err := uuid.Parse(portfolioID)
	if err != nil {
		return fmt.Errorf("invalid portfolio ID format: %w", err)
	}

	result := r.db.Where("portfolio_id = ? AND symbol = ?", pid, symbol).Delete(&models.Holding{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete holding: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return models.ErrHoldingNotFound
	}

	return nil
}
