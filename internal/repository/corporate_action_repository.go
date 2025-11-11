package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/lenon/portfolios/internal/models"
)

// CorporateActionRepository defines the interface for corporate action data operations
type CorporateActionRepository interface {
	Create(action *models.CorporateAction) error
	FindByID(id string) (*models.CorporateAction, error)
	FindBySymbol(symbol string) ([]*models.CorporateAction, error)
	FindBySymbolAndDateRange(symbol string, startDate, endDate time.Time) ([]*models.CorporateAction, error)
	FindUnapplied() ([]*models.CorporateAction, error)
	Update(action *models.CorporateAction) error
	Delete(id string) error
}

// corporateActionRepository implements CorporateActionRepository interface
type corporateActionRepository struct {
	db *gorm.DB
}

// NewCorporateActionRepository creates a new CorporateActionRepository instance
func NewCorporateActionRepository(db *gorm.DB) CorporateActionRepository {
	return &corporateActionRepository{db: db}
}

// Create creates a new corporate action in the database
func (r *corporateActionRepository) Create(action *models.CorporateAction) error {
	if action == nil {
		return fmt.Errorf("corporate action cannot be nil")
	}

	if err := r.db.Create(action).Error; err != nil {
		return fmt.Errorf("failed to create corporate action: %w", err)
	}

	return nil
}

// FindByID finds a corporate action by ID
func (r *corporateActionRepository) FindByID(id string) (*models.CorporateAction, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	actionID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid corporate action ID format: %w", err)
	}

	var action models.CorporateAction
	if err := r.db.Where("id = ?", actionID).First(&action).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrCorporateActionNotFound
		}
		return nil, fmt.Errorf("failed to find corporate action: %w", err)
	}

	return &action, nil
}

// FindBySymbol finds all corporate actions for a symbol, ordered by date descending
func (r *corporateActionRepository) FindBySymbol(symbol string) ([]*models.CorporateAction, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	var actions []*models.CorporateAction
	if err := r.db.Where("symbol = ?", symbol).
		Order("date DESC").
		Find(&actions).Error; err != nil {
		return nil, fmt.Errorf("failed to find corporate actions: %w", err)
	}

	return actions, nil
}

// FindBySymbolAndDateRange finds corporate actions for a symbol within a date range
func (r *corporateActionRepository) FindBySymbolAndDateRange(
	symbol string,
	startDate, endDate time.Time,
) ([]*models.CorporateAction, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	var actions []*models.CorporateAction
	if err := r.db.Where("symbol = ? AND date >= ? AND date <= ?", symbol, startDate, endDate).
		Order("date ASC").
		Find(&actions).Error; err != nil {
		return nil, fmt.Errorf("failed to find corporate actions: %w", err)
	}

	return actions, nil
}

// FindUnapplied finds all corporate actions that haven't been applied yet
func (r *corporateActionRepository) FindUnapplied() ([]*models.CorporateAction, error) {
	var actions []*models.CorporateAction
	if err := r.db.Where("applied = ?", false).
		Order("date ASC").
		Find(&actions).Error; err != nil {
		return nil, fmt.Errorf("failed to find unapplied corporate actions: %w", err)
	}

	return actions, nil
}

// Update updates an existing corporate action
func (r *corporateActionRepository) Update(action *models.CorporateAction) error {
	if action == nil {
		return fmt.Errorf("corporate action cannot be nil")
	}

	if err := r.db.Save(action).Error; err != nil {
		return fmt.Errorf("failed to update corporate action: %w", err)
	}

	return nil
}

// Delete deletes a corporate action by ID
func (r *corporateActionRepository) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	actionID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid corporate action ID format: %w", err)
	}

	result := r.db.Where("id = ?", actionID).Delete(&models.CorporateAction{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete corporate action: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return models.ErrCorporateActionNotFound
	}

	return nil
}
