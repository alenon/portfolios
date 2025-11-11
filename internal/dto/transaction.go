package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/models"
)

// CreateTransactionRequest represents the request to create a new transaction
type CreateTransactionRequest struct {
	Type       models.TransactionType `json:"type" binding:"required,oneof=BUY SELL DIVIDEND SPLIT MERGER SPINOFF DIVIDEND_REINVEST"`
	Symbol     string                 `json:"symbol" binding:"required,min=1,max=20"`
	Date       time.Time              `json:"date" binding:"required"`
	Quantity   decimal.Decimal        `json:"quantity" binding:"required"`
	Price      *decimal.Decimal       `json:"price,omitempty"`
	Commission decimal.Decimal        `json:"commission"`
	Currency   string                 `json:"currency,omitempty" binding:"omitempty,len=3"`
	Notes      string                 `json:"notes,omitempty"`
}

// UpdateTransactionRequest represents the request to update a transaction
type UpdateTransactionRequest struct {
	Type       models.TransactionType `json:"type" binding:"required,oneof=BUY SELL DIVIDEND SPLIT MERGER SPINOFF DIVIDEND_REINVEST"`
	Symbol     string                 `json:"symbol" binding:"required,min=1,max=20"`
	Date       time.Time              `json:"date" binding:"required"`
	Quantity   decimal.Decimal        `json:"quantity" binding:"required"`
	Price      *decimal.Decimal       `json:"price,omitempty"`
	Commission decimal.Decimal        `json:"commission"`
	Currency   string                 `json:"currency,omitempty" binding:"omitempty,len=3"`
	Notes      string                 `json:"notes,omitempty"`
}

// TransactionResponse represents a transaction in API responses
type TransactionResponse struct {
	ID            uuid.UUID              `json:"id"`
	PortfolioID   uuid.UUID              `json:"portfolio_id"`
	Type          models.TransactionType `json:"type"`
	Symbol        string                 `json:"symbol"`
	Date          time.Time              `json:"date"`
	Quantity      decimal.Decimal        `json:"quantity"`
	Price         *decimal.Decimal       `json:"price,omitempty"`
	Commission    decimal.Decimal        `json:"commission"`
	Currency      string                 `json:"currency"`
	Notes         string                 `json:"notes,omitempty"`
	ImportBatchID *uuid.UUID             `json:"import_batch_id,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// TransactionListResponse represents a list of transactions
type TransactionListResponse struct {
	Transactions []*TransactionResponse `json:"transactions"`
	Total        int                    `json:"total"`
}

// ToTransactionResponse converts a Transaction model to TransactionResponse DTO
func ToTransactionResponse(transaction *models.Transaction) *TransactionResponse {
	if transaction == nil {
		return nil
	}

	return &TransactionResponse{
		ID:            transaction.ID,
		PortfolioID:   transaction.PortfolioID,
		Type:          transaction.Type,
		Symbol:        transaction.Symbol,
		Date:          transaction.Date,
		Quantity:      transaction.Quantity,
		Price:         transaction.Price,
		Commission:    transaction.Commission,
		Currency:      transaction.Currency,
		Notes:         transaction.Notes,
		ImportBatchID: transaction.ImportBatchID,
		CreatedAt:     transaction.CreatedAt,
		UpdatedAt:     transaction.UpdatedAt,
	}
}

// ToTransactionListResponse converts a list of Transaction models to TransactionListResponse DTO
func ToTransactionListResponse(transactions []*models.Transaction) *TransactionListResponse {
	response := &TransactionListResponse{
		Transactions: make([]*TransactionResponse, 0, len(transactions)),
		Total:        len(transactions),
	}

	for _, transaction := range transactions {
		response.Transactions = append(response.Transactions, ToTransactionResponse(transaction))
	}

	return response
}
