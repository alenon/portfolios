package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeBuy              TransactionType = "BUY"
	TransactionTypeSell             TransactionType = "SELL"
	TransactionTypeDividend         TransactionType = "DIVIDEND"
	TransactionTypeSplit            TransactionType = "SPLIT"
	TransactionTypeMerger           TransactionType = "MERGER"
	TransactionTypeSpinoff          TransactionType = "SPINOFF"
	TransactionTypeDividendReinvest TransactionType = "DIVIDEND_REINVEST"
	TransactionTypeTickerChange     TransactionType = "TICKER_CHANGE"
)

// Transaction represents a portfolio transaction
type Transaction struct {
	ID            uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	PortfolioID   uuid.UUID        `gorm:"type:uuid;not null;index" json:"portfolio_id" validate:"required"`
	Type          TransactionType  `gorm:"type:varchar(20);not null" json:"type" validate:"required"`
	Symbol        string           `gorm:"type:varchar(20);not null;index" json:"symbol" validate:"required"`
	Date          time.Time        `gorm:"not null;index" json:"date" validate:"required"`
	Quantity      decimal.Decimal  `gorm:"type:numeric(20,8);not null" json:"quantity" validate:"required"`
	Price         *decimal.Decimal `gorm:"type:numeric(20,8)" json:"price,omitempty"`
	Commission    decimal.Decimal  `gorm:"type:numeric(20,8);not null;default:0" json:"commission"`
	Currency      string           `gorm:"type:varchar(3);not null;default:'USD'" json:"currency"`
	Notes         string           `gorm:"type:text" json:"notes,omitempty"`
	ImportBatchID *uuid.UUID       `gorm:"type:uuid" json:"import_batch_id,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	Portfolio     *Portfolio       `gorm:"foreignKey:PortfolioID" json:"portfolio,omitempty"`
}

// TableName specifies the table name for the Transaction model
func (Transaction) TableName() string {
	return "transactions"
}

// BeforeCreate hook to generate UUID before creating a new transaction
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = time.Now().UTC()
	}
	if t.Currency == "" {
		t.Currency = "USD"
	}
	if t.Commission.IsZero() {
		t.Commission = decimal.Zero
	}
	return nil
}

// BeforeUpdate hook to update the UpdatedAt timestamp
func (t *Transaction) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// Validate checks if the transaction has valid data
func (t *Transaction) Validate() error {
	if t.Symbol == "" {
		return ErrInvalidSymbol
	}
	if t.Quantity.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidQuantity
	}
	if !t.isValidTransactionType() {
		return ErrInvalidTransactionType
	}
	// Price is required for BUY and SELL transactions
	if (t.Type == TransactionTypeBuy || t.Type == TransactionTypeSell) && (t.Price == nil || t.Price.LessThanOrEqual(decimal.Zero)) {
		return ErrInvalidPrice
	}
	if t.Commission.IsNegative() {
		return ErrInvalidPrice
	}
	return nil
}

// isValidTransactionType checks if the transaction type is valid
func (t *Transaction) isValidTransactionType() bool {
	switch t.Type {
	case TransactionTypeBuy, TransactionTypeSell, TransactionTypeDividend,
		TransactionTypeSplit, TransactionTypeMerger, TransactionTypeSpinoff,
		TransactionTypeDividendReinvest, TransactionTypeTickerChange:
		return true
	default:
		return false
	}
}

// IsBuy returns true if the transaction is a buy
func (t *Transaction) IsBuy() bool {
	return t.Type == TransactionTypeBuy || t.Type == TransactionTypeDividendReinvest
}

// IsSell returns true if the transaction is a sell
func (t *Transaction) IsSell() bool {
	return t.Type == TransactionTypeSell
}

// GetTotalCost returns the total cost of the transaction including commission
func (t *Transaction) GetTotalCost() decimal.Decimal {
	if t.Price == nil {
		return t.Commission
	}
	return t.Price.Mul(t.Quantity).Add(t.Commission)
}

// GetProceeds returns the proceeds from a sell transaction minus commission
func (t *Transaction) GetProceeds() decimal.Decimal {
	if t.Price == nil {
		return decimal.Zero
	}
	return t.Price.Mul(t.Quantity).Sub(t.Commission)
}
