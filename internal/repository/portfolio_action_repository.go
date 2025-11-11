package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/lenon/portfolios/internal/models"
)

// PortfolioActionRepository defines the interface for portfolio action data operations
type PortfolioActionRepository interface {
	Create(action *models.PortfolioAction) error
	FindByID(id string) (*models.PortfolioAction, error)
	FindByPortfolioID(portfolioID string) ([]*models.PortfolioAction, error)
	FindPendingByPortfolioID(portfolioID string) ([]*models.PortfolioAction, error)
	FindByPortfolioIDAndStatus(portfolioID string, status models.PortfolioActionStatus) ([]*models.PortfolioAction, error)
	FindPendingByCorporateActionID(corporateActionID string) ([]*models.PortfolioAction, error)
	Update(action *models.PortfolioAction) error
	Delete(id string) error
	ExistsPendingForPortfolioAndAction(portfolioID, corporateActionID string) (bool, error)
}

// portfolioActionRepository implements PortfolioActionRepository interface
type portfolioActionRepository struct {
	db *gorm.DB
}

// NewPortfolioActionRepository creates a new PortfolioActionRepository instance
func NewPortfolioActionRepository(db *gorm.DB) PortfolioActionRepository {
	return &portfolioActionRepository{db: db}
}

// Create creates a new portfolio action in the database
func (r *portfolioActionRepository) Create(action *models.PortfolioAction) error {
	if action == nil {
		return fmt.Errorf("portfolio action cannot be nil")
	}

	if err := r.db.Create(action).Error; err != nil {
		return fmt.Errorf("failed to create portfolio action: %w", err)
	}

	return nil
}

// FindByID finds a portfolio action by ID with preloaded relationships
func (r *portfolioActionRepository) FindByID(id string) (*models.PortfolioAction, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	actionID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid portfolio action ID format: %w", err)
	}

	var action models.PortfolioAction
	if err := r.db.Preload("Portfolio").
		Preload("CorporateAction").
		Where("id = ?", actionID).
		First(&action).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("portfolio action not found")
		}
		return nil, fmt.Errorf("failed to find portfolio action: %w", err)
	}

	return &action, nil
}

// FindByPortfolioID finds all portfolio actions for a portfolio
func (r *portfolioActionRepository) FindByPortfolioID(portfolioID string) ([]*models.PortfolioAction, error) {
	if portfolioID == "" {
		return nil, fmt.Errorf("portfolio ID cannot be empty")
	}

	pid, err := uuid.Parse(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("invalid portfolio ID format: %w", err)
	}

	var actions []*models.PortfolioAction
	if err := r.db.Preload("CorporateAction").
		Where("portfolio_id = ?", pid).
		Order("detected_at DESC").
		Find(&actions).Error; err != nil {
		return nil, fmt.Errorf("failed to find portfolio actions: %w", err)
	}

	return actions, nil
}

// FindPendingByPortfolioID finds all pending portfolio actions for a portfolio
func (r *portfolioActionRepository) FindPendingByPortfolioID(portfolioID string) ([]*models.PortfolioAction, error) {
	return r.FindByPortfolioIDAndStatus(portfolioID, models.PortfolioActionStatusPending)
}

// FindByPortfolioIDAndStatus finds portfolio actions by portfolio ID and status
func (r *portfolioActionRepository) FindByPortfolioIDAndStatus(
	portfolioID string,
	status models.PortfolioActionStatus,
) ([]*models.PortfolioAction, error) {
	if portfolioID == "" {
		return nil, fmt.Errorf("portfolio ID cannot be empty")
	}

	pid, err := uuid.Parse(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("invalid portfolio ID format: %w", err)
	}

	var actions []*models.PortfolioAction
	if err := r.db.Preload("CorporateAction").
		Where("portfolio_id = ? AND status = ?", pid, status).
		Order("detected_at DESC").
		Find(&actions).Error; err != nil {
		return nil, fmt.Errorf("failed to find portfolio actions: %w", err)
	}

	return actions, nil
}

// FindPendingByCorporateActionID finds all pending portfolio actions for a corporate action
func (r *portfolioActionRepository) FindPendingByCorporateActionID(corporateActionID string) ([]*models.PortfolioAction, error) {
	if corporateActionID == "" {
		return nil, fmt.Errorf("corporate action ID cannot be empty")
	}

	caid, err := uuid.Parse(corporateActionID)
	if err != nil {
		return nil, fmt.Errorf("invalid corporate action ID format: %w", err)
	}

	var actions []*models.PortfolioAction
	if err := r.db.Preload("Portfolio").
		Where("corporate_action_id = ? AND status = ?", caid, models.PortfolioActionStatusPending).
		Find(&actions).Error; err != nil {
		return nil, fmt.Errorf("failed to find portfolio actions: %w", err)
	}

	return actions, nil
}

// Update updates an existing portfolio action
func (r *portfolioActionRepository) Update(action *models.PortfolioAction) error {
	if action == nil {
		return fmt.Errorf("portfolio action cannot be nil")
	}

	action.UpdatedAt = time.Now().UTC()

	if err := r.db.Save(action).Error; err != nil {
		return fmt.Errorf("failed to update portfolio action: %w", err)
	}

	return nil
}

// Delete deletes a portfolio action by ID
func (r *portfolioActionRepository) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	actionID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid portfolio action ID format: %w", err)
	}

	result := r.db.Where("id = ?", actionID).Delete(&models.PortfolioAction{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete portfolio action: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("portfolio action not found")
	}

	return nil
}

// ExistsPendingForPortfolioAndAction checks if a pending action already exists
func (r *portfolioActionRepository) ExistsPendingForPortfolioAndAction(
	portfolioID, corporateActionID string,
) (bool, error) {
	if portfolioID == "" || corporateActionID == "" {
		return false, fmt.Errorf("portfolio ID and corporate action ID cannot be empty")
	}

	pid, err := uuid.Parse(portfolioID)
	if err != nil {
		return false, fmt.Errorf("invalid portfolio ID format: %w", err)
	}

	caid, err := uuid.Parse(corporateActionID)
	if err != nil {
		return false, fmt.Errorf("invalid corporate action ID format: %w", err)
	}

	var count int64
	if err := r.db.Model(&models.PortfolioAction{}).
		Where("portfolio_id = ? AND corporate_action_id = ? AND status = ?",
			pid, caid, models.PortfolioActionStatusPending).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check portfolio action existence: %w", err)
	}

	return count > 0, nil
}
