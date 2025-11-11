package services

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
)

// CorporateActionService defines the interface for corporate action operations
type CorporateActionService interface {
	// Core operations
	Create(symbol string, actionType models.CorporateActionType, date time.Time, ratio, amount *decimal.Decimal, newSymbol, currency, description *string) (*models.CorporateAction, error)
	GetByID(id string) (*models.CorporateAction, error)
	GetBySymbol(symbol string) ([]*models.CorporateAction, error)
	GetBySymbolAndDateRange(symbol string, startDate, endDate time.Time) ([]*models.CorporateAction, error)
	GetUnapplied() ([]*models.CorporateAction, error)
	MarkAsApplied(id string) error
	Delete(id string) error

	// Application methods (placeholder for future implementation)
	ApplyStockSplit(portfolioID, symbol, userID string, ratio decimal.Decimal, date time.Time) error
	ApplyDividend(portfolioID, symbol, userID string, amount decimal.Decimal, date time.Time) error
	ApplyMerger(portfolioID, oldSymbol, newSymbol, userID string, ratio decimal.Decimal, date time.Time) error
}

// corporateActionService implements CorporateActionService interface
type corporateActionService struct {
	corporateActionRepo repository.CorporateActionRepository
	portfolioRepo       repository.PortfolioRepository
	transactionRepo     repository.TransactionRepository
	holdingRepo         repository.HoldingRepository
	taxLotRepo          repository.TaxLotRepository
}

// NewCorporateActionService creates a new CorporateActionService instance
func NewCorporateActionService(
	corporateActionRepo repository.CorporateActionRepository,
	portfolioRepo repository.PortfolioRepository,
	transactionRepo repository.TransactionRepository,
	holdingRepo repository.HoldingRepository,
	taxLotRepo repository.TaxLotRepository,
) CorporateActionService {
	return &corporateActionService{
		corporateActionRepo: corporateActionRepo,
		portfolioRepo:       portfolioRepo,
		transactionRepo:     transactionRepo,
		holdingRepo:         holdingRepo,
		taxLotRepo:          taxLotRepo,
	}
}

// Create creates a new corporate action
func (s *corporateActionService) Create(
	symbol string,
	actionType models.CorporateActionType,
	date time.Time,
	ratio, amount *decimal.Decimal,
	newSymbol, currency, description *string,
) (*models.CorporateAction, error) {
	action := &models.CorporateAction{
		Symbol:      symbol,
		Type:        actionType,
		Date:        date,
		Ratio:       ratio,
		Amount:      amount,
		NewSymbol:   newSymbol,
		Currency:    currency,
		Description: *description,
		Applied:     false,
	}

	if err := action.Validate(); err != nil {
		return nil, err
	}

	if err := s.corporateActionRepo.Create(action); err != nil {
		return nil, fmt.Errorf("failed to create corporate action: %w", err)
	}

	return action, nil
}

// GetByID retrieves a corporate action by ID
func (s *corporateActionService) GetByID(id string) (*models.CorporateAction, error) {
	return s.corporateActionRepo.FindByID(id)
}

// GetBySymbol retrieves all corporate actions for a symbol
func (s *corporateActionService) GetBySymbol(symbol string) ([]*models.CorporateAction, error) {
	return s.corporateActionRepo.FindBySymbol(symbol)
}

// GetBySymbolAndDateRange retrieves corporate actions for a symbol within a date range
func (s *corporateActionService) GetBySymbolAndDateRange(
	symbol string,
	startDate, endDate time.Time,
) ([]*models.CorporateAction, error) {
	return s.corporateActionRepo.FindBySymbolAndDateRange(symbol, startDate, endDate)
}

// GetUnapplied retrieves all unapplied corporate actions
func (s *corporateActionService) GetUnapplied() ([]*models.CorporateAction, error) {
	return s.corporateActionRepo.FindUnapplied()
}

// MarkAsApplied marks a corporate action as applied
func (s *corporateActionService) MarkAsApplied(id string) error {
	action, err := s.corporateActionRepo.FindByID(id)
	if err != nil {
		return err
	}

	action.Applied = true
	return s.corporateActionRepo.Update(action)
}

// Delete deletes a corporate action
func (s *corporateActionService) Delete(id string) error {
	return s.corporateActionRepo.Delete(id)
}

// ApplyStockSplit applies a stock split to a portfolio
// TODO: Implement full stock split logic
func (s *corporateActionService) ApplyStockSplit(
	portfolioID, symbol, userID string,
	ratio decimal.Decimal,
	date time.Time,
) error {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return models.ErrUnauthorizedAccess
	}

	// TODO: Implement stock split logic:
	// 1. Get all holdings for the symbol
	// 2. Multiply quantities by split ratio
	// 3. Divide cost basis per share by split ratio
	// 4. Update all tax lots
	// 5. Create a SPLIT transaction for audit trail

	return fmt.Errorf("stock split application not yet implemented")
}

// ApplyDividend applies a dividend to a portfolio
// TODO: Implement full dividend logic
func (s *corporateActionService) ApplyDividend(
	portfolioID, symbol, userID string,
	amount decimal.Decimal,
	date time.Time,
) error {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return models.ErrUnauthorizedAccess
	}

	// TODO: Implement dividend logic:
	// 1. Get holding for the symbol
	// 2. Calculate total dividend amount (quantity * per-share dividend)
	// 3. Create a DIVIDEND transaction
	// 4. If DRIP, create additional shares and tax lots

	return fmt.Errorf("dividend application not yet implemented")
}

// ApplyMerger applies a merger/acquisition to a portfolio
// TODO: Implement full merger logic
func (s *corporateActionService) ApplyMerger(
	portfolioID, oldSymbol, newSymbol, userID string,
	ratio decimal.Decimal,
	date time.Time,
) error {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return models.ErrUnauthorizedAccess
	}

	// TODO: Implement merger logic:
	// 1. Get all holdings for old symbol
	// 2. Close out old symbol holdings
	// 3. Create new holdings in new symbol based on merger ratio
	// 4. Transfer cost basis proportionally
	// 5. Update/create tax lots for new symbol
	// 6. Create MERGER transaction for audit trail

	return fmt.Errorf("merger application not yet implemented")
}
