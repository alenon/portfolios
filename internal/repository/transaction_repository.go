package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/lenon/portfolios/internal/models"
)

// TransactionRepository defines the interface for transaction data operations
type TransactionRepository interface {
	Create(transaction *models.Transaction) error
	FindByID(id string) (*models.Transaction, error)
	FindByPortfolioID(portfolioID string) ([]*models.Transaction, error)
	FindByPortfolioIDAndSymbol(portfolioID, symbol string) ([]*models.Transaction, error)
	FindByPortfolioIDWithFilters(portfolioID string, symbol *string, startDate, endDate *time.Time) ([]*models.Transaction, error)
	Update(transaction *models.Transaction) error
	Delete(id string) error
	DeleteByImportBatchID(batchID string) error
}

// transactionRepository implements TransactionRepository interface
type transactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new TransactionRepository instance
func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

// Create creates a new transaction in the database
func (r *transactionRepository) Create(transaction *models.Transaction) error {
	if transaction == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	if err := r.db.Create(transaction).Error; err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// FindByID finds a transaction by ID
func (r *transactionRepository) FindByID(id string) (*models.Transaction, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	transactionID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction ID format: %w", err)
	}

	var transaction models.Transaction
	err = r.db.Where("id = ?", transactionID).First(&transaction).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrTransactionNotFound
		}
		return nil, fmt.Errorf("failed to find transaction: %w", err)
	}

	return &transaction, nil
}

// FindByPortfolioID finds all transactions for a specific portfolio
func (r *transactionRepository) FindByPortfolioID(portfolioID string) ([]*models.Transaction, error) {
	if portfolioID == "" {
		return nil, fmt.Errorf("portfolio ID cannot be empty")
	}

	// Validate UUID format
	pid, err := uuid.Parse(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("invalid portfolio ID format: %w", err)
	}

	var transactions []*models.Transaction
	err = r.db.Where("portfolio_id = ?", pid).Order("date DESC, created_at DESC").Find(&transactions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find transactions: %w", err)
	}

	return transactions, nil
}

// FindByPortfolioIDAndSymbol finds all transactions for a specific portfolio and symbol
func (r *transactionRepository) FindByPortfolioIDAndSymbol(portfolioID, symbol string) ([]*models.Transaction, error) {
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

	var transactions []*models.Transaction
	err = r.db.Where("portfolio_id = ? AND symbol = ?", pid, symbol).
		Order("date DESC, created_at DESC").
		Find(&transactions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find transactions: %w", err)
	}

	return transactions, nil
}

// FindByPortfolioIDWithFilters finds transactions with optional filters
func (r *transactionRepository) FindByPortfolioIDWithFilters(
	portfolioID string,
	symbol *string,
	startDate, endDate *time.Time,
) ([]*models.Transaction, error) {
	if portfolioID == "" {
		return nil, fmt.Errorf("portfolio ID cannot be empty")
	}

	// Validate UUID format
	pid, err := uuid.Parse(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("invalid portfolio ID format: %w", err)
	}

	query := r.db.Where("portfolio_id = ?", pid)

	if symbol != nil && *symbol != "" {
		query = query.Where("symbol = ?", *symbol)
	}

	if startDate != nil {
		query = query.Where("date >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("date <= ?", *endDate)
	}

	var transactions []*models.Transaction
	err = query.Order("date DESC, created_at DESC").Find(&transactions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find transactions: %w", err)
	}

	return transactions, nil
}

// Update updates an existing transaction
func (r *transactionRepository) Update(transaction *models.Transaction) error {
	if transaction == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	result := r.db.Model(transaction).Where("id = ?", transaction.ID).Updates(transaction)
	if result.Error != nil {
		return fmt.Errorf("failed to update transaction: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return models.ErrTransactionNotFound
	}

	return nil
}

// Delete deletes a transaction by ID
func (r *transactionRepository) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	// Validate UUID format
	transactionID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid transaction ID format: %w", err)
	}

	result := r.db.Where("id = ?", transactionID).Delete(&models.Transaction{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete transaction: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return models.ErrTransactionNotFound
	}

	return nil
}

// DeleteByImportBatchID deletes all transactions with a specific import batch ID
func (r *transactionRepository) DeleteByImportBatchID(batchID string) error {
	if batchID == "" {
		return fmt.Errorf("batch ID cannot be empty")
	}

	// Validate UUID format
	bid, err := uuid.Parse(batchID)
	if err != nil {
		return fmt.Errorf("invalid batch ID format: %w", err)
	}

	result := r.db.Where("import_batch_id = ?", bid).Delete(&models.Transaction{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete transactions: %w", result.Error)
	}

	return nil
}
