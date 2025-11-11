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

	// Application methods
	ApplyStockSplit(portfolioID, symbol, userID string, ratio decimal.Decimal, date time.Time) error
	ApplyDividend(portfolioID, symbol, userID string, amount decimal.Decimal, date time.Time) error
	ApplyMerger(portfolioID, oldSymbol, newSymbol, userID string, ratio decimal.Decimal, date time.Time) error
	ApplySpinoff(portfolioID, oldSymbol, newSymbol, userID string, ratio decimal.Decimal, date time.Time) error
	ApplyTickerChange(portfolioID, oldSymbol, newSymbol, userID string, date time.Time) error
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

// ApplySpinoff applies a spinoff to a portfolio
// A spinoff is when a company distributes shares of a subsidiary to existing shareholders
// Example: You own 100 shares of Company A. Company A spins off Company B at 0.5:1 ratio.
// You still have 100 shares of Company A, plus you receive 50 shares of Company B.
func (s *corporateActionService) ApplySpinoff(
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

	// Validate spinoff ratio
	if ratio.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("invalid spinoff ratio: must be greater than 0")
	}

	// 1. Get holding for parent symbol
	parentHolding, err := s.holdingRepo.FindByPortfolioIDAndSymbol(portfolioID, oldSymbol)
	if err != nil {
		return fmt.Errorf("no holding found for symbol %s: %w", oldSymbol, err)
	}

	// 2. Calculate spinoff shares
	// Example: 100 shares of parent at 0.5:1 ratio = 50 shares of spinoff
	spinoffQuantity := parentHolding.Quantity.Mul(ratio)

	// 3. Get or create holding for spinoff symbol
	// For spinoffs, the cost basis of the parent is typically allocated between
	// parent and spinoff based on relative fair market values on distribution date.
	// For simplicity, we'll allocate a small percentage (10%) to the spinoff,
	// but in production this should be configurable or based on actual market values.
	costBasisAllocationRatio := decimal.NewFromFloat(0.10) // 10% to spinoff, 90% stays with parent
	spinoffCostBasis := parentHolding.CostBasis.Mul(costBasisAllocationRatio)

	spinoffHolding, err := s.holdingRepo.FindByPortfolioIDAndSymbol(portfolioID, newSymbol)
	if err != nil {
		// Create new holding if it doesn't exist
		spinoffHolding = &models.Holding{
			PortfolioID:  portfolio.ID,
			Symbol:       newSymbol,
			Quantity:     decimal.Zero,
			CostBasis:    decimal.Zero,
			AvgCostPrice: decimal.Zero,
		}
	}

	// Add the spinoff shares
	spinoffHolding.AddShares(spinoffQuantity, spinoffCostBasis)

	if spinoffHolding.ID.String() == "00000000-0000-0000-0000-000000000000" || spinoffHolding.ID == [16]byte{} {
		if err := s.holdingRepo.Create(spinoffHolding); err != nil {
			return fmt.Errorf("failed to create spinoff holding: %w", err)
		}
	} else {
		if err := s.holdingRepo.Update(spinoffHolding); err != nil {
			return fmt.Errorf("failed to update spinoff holding: %w", err)
		}
	}

	// 4. Reduce parent holding cost basis (parent retains 90% of cost basis)
	parentHolding.CostBasis = parentHolding.CostBasis.Sub(spinoffCostBasis)
	parentHolding.CalculateAvgCostPrice()

	if err := s.holdingRepo.Update(parentHolding); err != nil {
		return fmt.Errorf("failed to update parent holding: %w", err)
	}

	// 5. Create tax lots for spinoff shares
	// Tax lots inherit the purchase dates from the parent company
	parentTaxLots, err := s.taxLotRepo.FindByPortfolioIDAndSymbol(portfolioID, oldSymbol)
	if err != nil {
		return fmt.Errorf("failed to retrieve parent tax lots: %w", err)
	}

	totalParentQuantity := decimal.Zero
	for _, lot := range parentTaxLots {
		totalParentQuantity = totalParentQuantity.Add(lot.Quantity)
	}

	for _, parentLot := range parentTaxLots {
		// Allocate spinoff shares proportionally based on each lot's percentage of total
		lotPercentage := parentLot.Quantity.Div(totalParentQuantity)
		spinoffLotQuantity := spinoffQuantity.Mul(lotPercentage)
		spinoffLotCostBasis := spinoffCostBasis.Mul(lotPercentage)

		// Create new tax lot for spinoff with same purchase date as parent
		spinoffLot := &models.TaxLot{
			PortfolioID:   portfolio.ID,
			Symbol:        newSymbol,
			PurchaseDate:  parentLot.PurchaseDate, // Inherit purchase date for tax purposes
			Quantity:      spinoffLotQuantity,
			CostBasis:     spinoffLotCostBasis,
			TransactionID: parentLot.TransactionID,
		}

		if err := s.taxLotRepo.Create(spinoffLot); err != nil {
			return fmt.Errorf("failed to create spinoff tax lot: %w", err)
		}

		// Reduce parent lot cost basis
		parentLot.CostBasis = parentLot.CostBasis.Sub(spinoffLotCostBasis)
		if err := s.taxLotRepo.Update(parentLot); err != nil {
			return fmt.Errorf("failed to update parent tax lot: %w", err)
		}
	}

	// 6. Create SPINOFF transaction for audit trail
	spinoffTransaction := &models.Transaction{
		PortfolioID: portfolio.ID,
		Type:        models.TransactionTypeSpinoff,
		Symbol:      newSymbol,
		Date:        date,
		Quantity:    spinoffQuantity,
		Price:       nil, // No price for spinoff transactions
		Commission:  decimal.Zero,
		Currency:    portfolio.BaseCurrency,
		Notes:       fmt.Sprintf("Spinoff: received %s shares of %s from %s at %s ratio", spinoffQuantity.String(), newSymbol, oldSymbol, ratio.String()),
	}

	if err := s.transactionRepo.Create(spinoffTransaction); err != nil {
		return fmt.Errorf("failed to create spinoff transaction: %w", err)
	}

	return nil
}

// ApplyTickerChange applies a ticker symbol change to a portfolio
// This is when a company changes its ticker symbol but remains the same company
// Example: Facebook (FB) changed to Meta (META)
func (s *corporateActionService) ApplyTickerChange(
	portfolioID, oldSymbol, newSymbol, userID string,
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

	// 1. Get holding for old symbol
	oldHolding, err := s.holdingRepo.FindByPortfolioIDAndSymbol(portfolioID, oldSymbol)
	if err != nil {
		return fmt.Errorf("no holding found for symbol %s: %w", oldSymbol, err)
	}

	// 2. For ticker changes, everything stays the same except the symbol
	// Quantity, cost basis, and average cost price all remain unchanged
	quantity := oldHolding.Quantity
	costBasis := oldHolding.CostBasis
	avgCostPrice := oldHolding.AvgCostPrice

	// 3. Get or create holding for new symbol
	newHolding, err := s.holdingRepo.FindByPortfolioIDAndSymbol(portfolioID, newSymbol)
	if err != nil {
		// Create new holding with same values
		newHolding = &models.Holding{
			PortfolioID:  portfolio.ID,
			Symbol:       newSymbol,
			Quantity:     quantity,
			CostBasis:    costBasis,
			AvgCostPrice: avgCostPrice,
		}
		if err := s.holdingRepo.Create(newHolding); err != nil {
			return fmt.Errorf("failed to create new holding: %w", err)
		}
	} else {
		// If holding already exists, add to it
		newHolding.AddShares(quantity, costBasis)
		if err := s.holdingRepo.Update(newHolding); err != nil {
			return fmt.Errorf("failed to update new holding: %w", err)
		}
	}

	// 4. Update all tax lots - change symbol but keep everything else
	oldTaxLots, err := s.taxLotRepo.FindByPortfolioIDAndSymbol(portfolioID, oldSymbol)
	if err != nil {
		return fmt.Errorf("failed to retrieve tax lots: %w", err)
	}

	for _, lot := range oldTaxLots {
		// Create new tax lot with new symbol but all other data preserved
		newLot := &models.TaxLot{
			PortfolioID:   portfolio.ID,
			Symbol:        newSymbol,
			PurchaseDate:  lot.PurchaseDate, // Preserve original purchase date
			Quantity:      lot.Quantity,     // Same quantity
			CostBasis:     lot.CostBasis,    // Same cost basis
			TransactionID: lot.TransactionID,
		}

		if err := s.taxLotRepo.Create(newLot); err != nil {
			return fmt.Errorf("failed to create new tax lot: %w", err)
		}
	}

	// 5. Update all old transactions with the old symbol to use the new symbol
	// This maintains historical accuracy
	oldTransactions, err := s.transactionRepo.FindByPortfolioIDAndSymbol(portfolioID, oldSymbol)
	if err != nil {
		return fmt.Errorf("failed to retrieve transactions: %w", err)
	}

	for _, txn := range oldTransactions {
		txn.Symbol = newSymbol
		if err := s.transactionRepo.Update(txn); err != nil {
			return fmt.Errorf("failed to update transaction: %w", err)
		}
	}

	// 6. Delete old holding and tax lots
	if err := s.holdingRepo.DeleteByPortfolioIDAndSymbol(portfolioID, oldSymbol); err != nil {
		return fmt.Errorf("failed to delete old holding: %w", err)
	}

	if err := s.taxLotRepo.DeleteByPortfolioIDAndSymbol(portfolioID, oldSymbol); err != nil {
		return fmt.Errorf("failed to delete old tax lots: %w", err)
	}

	// 7. Create TICKER_CHANGE transaction for audit trail
	tickerChangeTransaction := &models.Transaction{
		PortfolioID: portfolio.ID,
		Type:        models.TransactionTypeTickerChange,
		Symbol:      newSymbol,
		Date:        date,
		Quantity:    quantity,
		Price:       nil, // No price for ticker change
		Commission:  decimal.Zero,
		Currency:    portfolio.BaseCurrency,
		Notes:       fmt.Sprintf("Ticker change: %s changed to %s", oldSymbol, newSymbol),
	}

	if err := s.transactionRepo.Create(tickerChangeTransaction); err != nil {
		return fmt.Errorf("failed to create ticker change transaction: %w", err)
	}

	return nil
}
