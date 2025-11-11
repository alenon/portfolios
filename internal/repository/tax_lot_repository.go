package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/lenon/portfolios/internal/models"
)

// TaxLotRepository defines the interface for tax lot data operations
type TaxLotRepository interface {
	Create(taxLot *models.TaxLot) error
	FindByID(id string) (*models.TaxLot, error)
	FindByPortfolioID(portfolioID string) ([]*models.TaxLot, error)
	FindByPortfolioIDAndSymbol(portfolioID, symbol string) ([]*models.TaxLot, error)
	FindByTransactionID(transactionID string) ([]*models.TaxLot, error)
	Update(taxLot *models.TaxLot) error
	Delete(id string) error
	DeleteByPortfolioIDAndSymbol(portfolioID, symbol string) error
}

// taxLotRepository implements TaxLotRepository interface
type taxLotRepository struct {
	db *gorm.DB
}

// NewTaxLotRepository creates a new TaxLotRepository instance
func NewTaxLotRepository(db *gorm.DB) TaxLotRepository {
	return &taxLotRepository{db: db}
}

// Create creates a new tax lot in the database
func (r *taxLotRepository) Create(taxLot *models.TaxLot) error {
	if taxLot == nil {
		return fmt.Errorf("tax lot cannot be nil")
	}

	if err := r.db.Create(taxLot).Error; err != nil {
		return fmt.Errorf("failed to create tax lot: %w", err)
	}

	return nil
}

// FindByID finds a tax lot by ID
func (r *taxLotRepository) FindByID(id string) (*models.TaxLot, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	taxLotID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid tax lot ID format: %w", err)
	}

	var taxLot models.TaxLot
	if err := r.db.Where("id = ?", taxLotID).First(&taxLot).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrTaxLotNotFound
		}
		return nil, fmt.Errorf("failed to find tax lot: %w", err)
	}

	return &taxLot, nil
}

// FindByPortfolioID finds all tax lots for a portfolio, ordered by purchase date
func (r *taxLotRepository) FindByPortfolioID(portfolioID string) ([]*models.TaxLot, error) {
	if portfolioID == "" {
		return nil, fmt.Errorf("portfolio ID cannot be empty")
	}

	pid, err := uuid.Parse(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("invalid portfolio ID format: %w", err)
	}

	var taxLots []*models.TaxLot
	if err := r.db.Where("portfolio_id = ?", pid).
		Order("purchase_date ASC, created_at ASC").
		Find(&taxLots).Error; err != nil {
		return nil, fmt.Errorf("failed to find tax lots: %w", err)
	}

	return taxLots, nil
}

// FindByPortfolioIDAndSymbol finds all tax lots for a portfolio and symbol, ordered by purchase date (FIFO)
func (r *taxLotRepository) FindByPortfolioIDAndSymbol(portfolioID, symbol string) ([]*models.TaxLot, error) {
	if portfolioID == "" {
		return nil, fmt.Errorf("portfolio ID cannot be empty")
	}
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	pid, err := uuid.Parse(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("invalid portfolio ID format: %w", err)
	}

	var taxLots []*models.TaxLot
	if err := r.db.Where("portfolio_id = ? AND symbol = ?", pid, symbol).
		Order("purchase_date ASC, created_at ASC").
		Find(&taxLots).Error; err != nil {
		return nil, fmt.Errorf("failed to find tax lots: %w", err)
	}

	return taxLots, nil
}

// FindByTransactionID finds all tax lots associated with a transaction
func (r *taxLotRepository) FindByTransactionID(transactionID string) ([]*models.TaxLot, error) {
	if transactionID == "" {
		return nil, fmt.Errorf("transaction ID cannot be empty")
	}

	tid, err := uuid.Parse(transactionID)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction ID format: %w", err)
	}

	var taxLots []*models.TaxLot
	if err := r.db.Where("transaction_id = ?", tid).
		Find(&taxLots).Error; err != nil {
		return nil, fmt.Errorf("failed to find tax lots: %w", err)
	}

	return taxLots, nil
}

// Update updates an existing tax lot
func (r *taxLotRepository) Update(taxLot *models.TaxLot) error {
	if taxLot == nil {
		return fmt.Errorf("tax lot cannot be nil")
	}

	taxLot.UpdatedAt = time.Now().UTC()

	if err := r.db.Save(taxLot).Error; err != nil {
		return fmt.Errorf("failed to update tax lot: %w", err)
	}

	return nil
}

// Delete deletes a tax lot by ID
func (r *taxLotRepository) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	taxLotID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid tax lot ID format: %w", err)
	}

	result := r.db.Where("id = ?", taxLotID).Delete(&models.TaxLot{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete tax lot: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return models.ErrTaxLotNotFound
	}

	return nil
}

// DeleteByPortfolioIDAndSymbol deletes all tax lots for a portfolio and symbol
func (r *taxLotRepository) DeleteByPortfolioIDAndSymbol(portfolioID, symbol string) error {
	if portfolioID == "" {
		return fmt.Errorf("portfolio ID cannot be empty")
	}
	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	pid, err := uuid.Parse(portfolioID)
	if err != nil {
		return fmt.Errorf("invalid portfolio ID format: %w", err)
	}

	if err := r.db.Where("portfolio_id = ? AND symbol = ?", pid, symbol).
		Delete(&models.TaxLot{}).Error; err != nil {
		return fmt.Errorf("failed to delete tax lots: %w", err)
	}

	return nil
}
