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

	// Validate split ratio
	if ratio.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("invalid split ratio: must be greater than 0")
	}

	// 1. Get holding for the symbol
	holding, err := s.holdingRepo.FindByPortfolioIDAndSymbol(portfolioID, symbol)
	if err != nil {
		return fmt.Errorf("no holding found for symbol %s: %w", symbol, err)
	}

	// 2. Multiply quantity by split ratio, keep total cost basis the same
	// Example: 4:1 split of 100 shares @ $180/share
	// Before: 100 shares, $18,000 cost basis, $180 avg cost
	// After: 400 shares, $18,000 cost basis, $45 avg cost
	oldQuantity := holding.Quantity
	newQuantity := holding.Quantity.Mul(ratio)

	holding.Quantity = newQuantity
	// Cost basis stays the same (total investment doesn't change)
	// But avg cost per share decreases
	holding.CalculateAvgCostPrice()

	if err := s.holdingRepo.Update(holding); err != nil {
		return fmt.Errorf("failed to update holding: %w", err)
	}

	// 3. Update all tax lots for this symbol
	taxLots, err := s.taxLotRepo.FindByPortfolioIDAndSymbol(portfolioID, symbol)
	if err != nil {
		return fmt.Errorf("failed to retrieve tax lots: %w", err)
	}

	for _, lot := range taxLots {
		// Multiply quantity by split ratio
		lot.Quantity = lot.Quantity.Mul(ratio)
		// Cost basis stays the same for each lot
		// This automatically adjusts the per-share cost

		if err := s.taxLotRepo.Update(lot); err != nil {
			return fmt.Errorf("failed to update tax lot: %w", err)
		}
	}

	// 4. Create a SPLIT transaction for audit trail
	splitTransaction := &models.Transaction{
		PortfolioID: portfolio.ID,
		Type:        models.TransactionTypeSplit,
		Symbol:      symbol,
		Date:        date,
		Quantity:    newQuantity.Sub(oldQuantity), // Additional shares received
		Price:       nil,                          // No price for splits
		Commission:  decimal.Zero,
		Currency:    portfolio.BaseCurrency,
		Notes:       fmt.Sprintf("Stock split: %s ratio applied", ratio.String()),
	}

	if err := s.transactionRepo.Create(splitTransaction); err != nil {
		return fmt.Errorf("failed to create split transaction: %w", err)
	}

	return nil
}

// ApplyDividend applies a cash dividend to a portfolio
// Note: For dividend reinvestment (DRIP), use the transaction service to create a buy transaction
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

	// Validate dividend amount
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("invalid dividend amount: must be greater than 0")
	}

	// 1. Get holding for the symbol to verify it exists
	_, err = s.holdingRepo.FindByPortfolioIDAndSymbol(portfolioID, symbol)
	if err != nil {
		return fmt.Errorf("no holding found for symbol %s: %w", symbol, err)
	}

	// 2. Create a DIVIDEND transaction for record keeping
	// Note: Cash dividends don't affect share count or cost basis,
	// they just represent income received
	dividendTransaction := &models.Transaction{
		PortfolioID: portfolio.ID,
		Type:        models.TransactionTypeDividend,
		Symbol:      symbol,
		Date:        date,
		Quantity:    amount, // For dividends, quantity represents the total amount received
		Price:       nil,    // No price for dividend transactions
		Commission:  decimal.Zero,
		Currency:    portfolio.BaseCurrency,
		Notes:       fmt.Sprintf("Cash dividend: %s", amount.String()),
	}

	if err := s.transactionRepo.Create(dividendTransaction); err != nil {
		return fmt.Errorf("failed to create dividend transaction: %w", err)
	}

	// Note: For dividend reinvestment (DRIP), the calling code should:
	// 1. Calculate shares purchased: dividend amount / current share price
	// 2. Call transaction service to create a DIVIDEND_REINVEST buy transaction
	// 3. That will automatically create tax lots and update holdings

	return nil
}

// ApplyMerger applies a merger/acquisition to a portfolio
// This handles stock-for-stock mergers where oldSymbol is converted to newSymbol
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

	// Validate merger ratio
	if ratio.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("invalid merger ratio: must be greater than 0")
	}

	// 1. Get holding for old symbol
	oldHolding, err := s.holdingRepo.FindByPortfolioIDAndSymbol(portfolioID, oldSymbol)
	if err != nil {
		return fmt.Errorf("no holding found for symbol %s: %w", oldSymbol, err)
	}

	// 2. Calculate new position
	// Example: 1.5:1 ratio means 100 shares of A become 150 shares of B
	oldQuantity := oldHolding.Quantity
	newQuantity := oldQuantity.Mul(ratio)
	totalCostBasis := oldHolding.CostBasis // Cost basis transfers to new symbol

	// 3. Get or create holding for new symbol
	newHolding, err := s.holdingRepo.FindByPortfolioIDAndSymbol(portfolioID, newSymbol)
	if err != nil {
		// Create new holding if it doesn't exist
		newHolding = &models.Holding{
			PortfolioID:  portfolio.ID,
			Symbol:       newSymbol,
			Quantity:     decimal.Zero,
			CostBasis:    decimal.Zero,
			AvgCostPrice: decimal.Zero,
		}
	}

	// Add the converted shares to the new holding
	newHolding.AddShares(newQuantity, totalCostBasis)

	if newHolding.ID.String() == "00000000-0000-0000-0000-000000000000" || newHolding.ID == [16]byte{} {
		if err := s.holdingRepo.Create(newHolding); err != nil {
			return fmt.Errorf("failed to create new holding: %w", err)
		}
	} else {
		if err := s.holdingRepo.Update(newHolding); err != nil {
			return fmt.Errorf("failed to update new holding: %w", err)
		}
	}

	// 4. Update tax lots - convert old symbol lots to new symbol
	oldTaxLots, err := s.taxLotRepo.FindByPortfolioIDAndSymbol(portfolioID, oldSymbol)
	if err != nil {
		return fmt.Errorf("failed to retrieve tax lots: %w", err)
	}

	for _, lot := range oldTaxLots {
		// Calculate new quantity based on merger ratio
		newLotQuantity := lot.Quantity.Mul(ratio)

		// Create new tax lot for new symbol with preserved purchase date
		newLot := &models.TaxLot{
			PortfolioID:   portfolio.ID,
			Symbol:        newSymbol,
			PurchaseDate:  lot.PurchaseDate, // Preserve original purchase date for tax purposes
			Quantity:      newLotQuantity,
			CostBasis:     lot.CostBasis, // Cost basis transfers
			TransactionID: lot.TransactionID,
		}

		if err := s.taxLotRepo.Create(newLot); err != nil {
			return fmt.Errorf("failed to create new tax lot: %w", err)
		}
	}

	// 5. Delete old holding and tax lots
	if err := s.holdingRepo.DeleteByPortfolioIDAndSymbol(portfolioID, oldSymbol); err != nil {
		return fmt.Errorf("failed to delete old holding: %w", err)
	}

	if err := s.taxLotRepo.DeleteByPortfolioIDAndSymbol(portfolioID, oldSymbol); err != nil {
		return fmt.Errorf("failed to delete old tax lots: %w", err)
	}

	// 6. Create MERGER transaction for audit trail
	mergerTransaction := &models.Transaction{
		PortfolioID: portfolio.ID,
		Type:        models.TransactionTypeMerger,
		Symbol:      newSymbol,
		Date:        date,
		Quantity:    newQuantity,
		Price:       nil, // No price for merger transactions
		Commission:  decimal.Zero,
		Currency:    portfolio.BaseCurrency,
		Notes:       fmt.Sprintf("Merger: %s converted to %s at %s ratio (%s shares became %s shares)", oldSymbol, newSymbol, ratio.String(), oldQuantity.String(), newQuantity.String()),
	}

	if err := s.transactionRepo.Create(mergerTransaction); err != nil {
		return fmt.Errorf("failed to create merger transaction: %w", err)
	}

	return nil
}
