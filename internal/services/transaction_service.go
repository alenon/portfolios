package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
)

// TransactionService defines the interface for transaction operations
type TransactionService interface {
	Create(portfolioID, userID string, transactionType models.TransactionType, symbol string, date time.Time, quantity, price decimal.Decimal, commission decimal.Decimal, currency, notes string) (*models.Transaction, error)
	GetByID(id, userID string) (*models.Transaction, error)
	GetByPortfolioID(portfolioID, userID string) ([]*models.Transaction, error)
	GetByPortfolioIDAndSymbol(portfolioID, symbol, userID string) ([]*models.Transaction, error)
	Update(id, userID string, transactionType models.TransactionType, symbol string, date time.Time, quantity, price decimal.Decimal, commission decimal.Decimal, currency, notes string) (*models.Transaction, error)
	Delete(id, userID string) error
}

// transactionService implements TransactionService interface
type transactionService struct {
	transactionRepo repository.TransactionRepository
	portfolioRepo   repository.PortfolioRepository
	holdingRepo     repository.HoldingRepository
}

// NewTransactionService creates a new TransactionService instance
func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	portfolioRepo repository.PortfolioRepository,
	holdingRepo repository.HoldingRepository,
) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
		portfolioRepo:   portfolioRepo,
		holdingRepo:     holdingRepo,
	}
}

// recalculateHoldingsForSymbol recalculates holdings for a specific symbol in a portfolio
// This is used after update/delete operations to ensure holdings are accurate
func (s *transactionService) recalculateHoldingsForSymbol(portfolioID, symbol string) error {
	// Get all transactions for this symbol, ordered by date
	transactions, err := s.transactionRepo.FindByPortfolioIDAndSymbol(portfolioID, symbol)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	// Calculate new holdings based on all transactions
	var quantity decimal.Decimal
	var costBasis decimal.Decimal

	for _, tx := range transactions {
		switch tx.Type {
		case models.TransactionTypeBuy:
			totalCost := tx.GetTotalCost()
			quantity = quantity.Add(tx.Quantity)
			costBasis = costBasis.Add(totalCost)
		case models.TransactionTypeSell:
			if quantity.IsZero() {
				return models.ErrInsufficientShares
			}
			// Calculate average cost per share before the sale
			avgCostPrice := costBasis.Div(quantity)
			// Calculate cost basis for the shares being sold
			costBasisForSale := avgCostPrice.Mul(tx.Quantity)
			quantity = quantity.Sub(tx.Quantity)
			costBasis = costBasis.Sub(costBasisForSale)
		}
	}

	// Update or delete holding based on final quantity
	if quantity.IsZero() {
		// Delete holding if quantity is zero
		holding, err := s.holdingRepo.FindByPortfolioIDAndSymbol(portfolioID, symbol)
		if err == nil {
			return s.holdingRepo.Delete(holding.ID.String())
		}
		if err == models.ErrHoldingNotFound {
			return nil // Already deleted, that's fine
		}
		return err
	}

	// Update holding with recalculated values
	holding, err := s.holdingRepo.FindByPortfolioIDAndSymbol(portfolioID, symbol)
	if err == models.ErrHoldingNotFound {
		// Create new holding
		pid, err := uuid.Parse(portfolioID)
		if err != nil {
			return fmt.Errorf("invalid portfolio ID: %w", err)
		}
		newHolding := &models.Holding{
			PortfolioID:  pid,
			Symbol:       symbol,
			Quantity:     quantity,
			CostBasis:    costBasis,
			AvgCostPrice: costBasis.Div(quantity),
		}
		return s.holdingRepo.Create(newHolding)
	}
	if err != nil {
		return err
	}

	// Update existing holding
	holding.Quantity = quantity
	holding.CostBasis = costBasis
	holding.AvgCostPrice = costBasis.Div(quantity)
	return s.holdingRepo.Update(holding)
}

// Create creates a new transaction
func (s *transactionService) Create(
	portfolioID, userID string,
	transactionType models.TransactionType,
	symbol string,
	date time.Time,
	quantity, price, commission decimal.Decimal,
	currency, notes string,
) (*models.Transaction, error) {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return nil, models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	// Parse portfolio ID
	pid, err := uuid.Parse(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("invalid portfolio ID: %w", err)
	}

	// Set defaults
	if currency == "" {
		currency = portfolio.BaseCurrency
	}

	// Create transaction
	var pricePtr *decimal.Decimal
	if !price.IsZero() {
		pricePtr = &price
	}

	transaction := &models.Transaction{
		PortfolioID: pid,
		Type:        transactionType,
		Symbol:      symbol,
		Date:        date,
		Quantity:    quantity,
		Price:       pricePtr,
		Commission:  commission,
		Currency:    currency,
		Notes:       notes,
	}

	// Validate transaction
	if err := transaction.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.transactionRepo.Create(transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update holdings based on transaction type
	// Note: In a production system, this should be wrapped in a database transaction
	// to ensure atomicity. For now, we handle errors by attempting to delete the transaction
	// if holdings update fails.
	if err := s.updateHoldings(transaction, portfolio); err != nil {
		// Attempt to rollback by deleting the transaction
		if deleteErr := s.transactionRepo.Delete(transaction.ID.String()); deleteErr != nil {
			return nil, fmt.Errorf("failed to update holdings and rollback transaction: holdings error=%w, rollback error=%v", err, deleteErr)
		}
		return nil, fmt.Errorf("failed to update holdings (transaction rolled back): %w", err)
	}

	return transaction, nil
}

// GetByID retrieves a transaction by ID, ensuring it belongs to the user
func (s *transactionService) GetByID(id, userID string) (*models.Transaction, error) {
	transaction, err := s.transactionRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Verify the portfolio belongs to the user
	portfolio, err := s.portfolioRepo.FindByID(transaction.PortfolioID.String())
	if err != nil {
		return nil, models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	return transaction, nil
}

// GetByPortfolioID retrieves all transactions for a portfolio
func (s *transactionService) GetByPortfolioID(portfolioID, userID string) ([]*models.Transaction, error) {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return nil, models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	transactions, err := s.transactionRepo.FindByPortfolioID(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transactions: %w", err)
	}

	return transactions, nil
}

// GetByPortfolioIDAndSymbol retrieves all transactions for a portfolio and symbol
func (s *transactionService) GetByPortfolioIDAndSymbol(portfolioID, symbol, userID string) ([]*models.Transaction, error) {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return nil, models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	transactions, err := s.transactionRepo.FindByPortfolioIDAndSymbol(portfolioID, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transactions: %w", err)
	}

	return transactions, nil
}

// Update updates a transaction
func (s *transactionService) Update(
	id, userID string,
	transactionType models.TransactionType,
	symbol string,
	date time.Time,
	quantity, price, commission decimal.Decimal,
	currency, notes string,
) (*models.Transaction, error) {
	// Get existing transaction and verify ownership
	transaction, err := s.GetByID(id, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	transaction.Type = transactionType
	transaction.Symbol = symbol
	transaction.Date = date
	transaction.Quantity = quantity
	if !price.IsZero() {
		transaction.Price = &price
	}
	transaction.Commission = commission
	transaction.Currency = currency
	transaction.Notes = notes

	// Validate updated transaction
	if err := transaction.Validate(); err != nil {
		return nil, err
	}

	// Save changes
	if err := s.transactionRepo.Update(transaction); err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	// Recalculate holdings for affected symbol
	if err := s.recalculateHoldingsForSymbol(transaction.PortfolioID.String(), transaction.Symbol); err != nil {
		return nil, fmt.Errorf("failed to recalculate holdings: %w", err)
	}

	return transaction, nil
}

// Delete deletes a transaction, ensuring it belongs to the user
func (s *transactionService) Delete(id, userID string) error {
	// Get transaction and verify ownership
	transaction, err := s.GetByID(id, userID)
	if err != nil {
		return err
	}

	// Store symbol and portfolio ID before deletion
	symbol := transaction.Symbol
	portfolioID := transaction.PortfolioID.String()

	// Delete the transaction
	if err := s.transactionRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	// Recalculate holdings for affected symbol
	if err := s.recalculateHoldingsForSymbol(portfolioID, symbol); err != nil {
		return fmt.Errorf("failed to recalculate holdings: %w", err)
	}

	return nil
}

// updateHoldings updates the holdings table based on a transaction
func (s *transactionService) updateHoldings(transaction *models.Transaction, portfolio *models.Portfolio) error {
	// Only update holdings for BUY and SELL transactions
	if transaction.Type != models.TransactionTypeBuy && transaction.Type != models.TransactionTypeSell {
		return nil
	}

	// Get current holding for this symbol
	holding, err := s.holdingRepo.FindByPortfolioIDAndSymbol(
		transaction.PortfolioID.String(),
		transaction.Symbol,
	)

	// If no holding exists and this is a buy, create new holding
	if err == models.ErrHoldingNotFound {
		if transaction.Type == models.TransactionTypeBuy {
			totalCost := transaction.GetTotalCost()
			newHolding := &models.Holding{
				PortfolioID:  transaction.PortfolioID,
				Symbol:       transaction.Symbol,
				Quantity:     transaction.Quantity,
				CostBasis:    totalCost,
				AvgCostPrice: totalCost.Div(transaction.Quantity),
			}
			return s.holdingRepo.Create(newHolding)
		}
		return models.ErrInsufficientShares
	}

	if err != nil {
		return fmt.Errorf("failed to get holding: %w", err)
	}

	// Update existing holding
	switch transaction.Type {
	case models.TransactionTypeBuy:
		totalCost := transaction.GetTotalCost()
		holding.AddShares(transaction.Quantity, totalCost)
	case models.TransactionTypeSell:
		// For FIFO, use average cost basis
		costBasisForSale := holding.AvgCostPrice.Mul(transaction.Quantity)
		if err := holding.RemoveShares(transaction.Quantity, costBasisForSale); err != nil {
			return err
		}
	}

	// If quantity is zero, delete the holding
	if holding.Quantity.IsZero() {
		return s.holdingRepo.Delete(holding.ID.String())
	}

	return s.holdingRepo.Update(holding)
}
