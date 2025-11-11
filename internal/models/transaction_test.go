package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTransaction_BeforeCreate(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	t.Run("sets defaults", func(t *testing.T) {
		transaction := &Transaction{
			PortfolioID: uuid.New(),
			Type:        TransactionTypeBuy,
			Symbol:      "AAPL",
			Date:        time.Now(),
			Quantity:    decimal.NewFromInt(10),
		}

		err := transaction.BeforeCreate(db)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, transaction.ID)
		assert.False(t, transaction.CreatedAt.IsZero())
		assert.False(t, transaction.UpdatedAt.IsZero())
		assert.Equal(t, "USD", transaction.Currency)
		assert.True(t, transaction.Commission.IsZero())
	})
}

func TestTransaction_BeforeUpdate(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	transaction := &Transaction{}
	err := transaction.BeforeUpdate(db)

	assert.NoError(t, err)
	assert.False(t, transaction.UpdatedAt.IsZero())
}

func TestTransaction_Validate(t *testing.T) {
	price := decimal.NewFromFloat(150.00)

	t.Run("valid buy transaction", func(t *testing.T) {
		transaction := &Transaction{
			Type:     TransactionTypeBuy,
			Symbol:   "AAPL",
			Quantity: decimal.NewFromInt(10),
			Price:    &price,
		}

		err := transaction.Validate()

		assert.NoError(t, err)
	})

	t.Run("empty symbol error", func(t *testing.T) {
		transaction := &Transaction{
			Type:     TransactionTypeBuy,
			Symbol:   "",
			Quantity: decimal.NewFromInt(10),
			Price:    &price,
		}

		err := transaction.Validate()

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidSymbol, err)
	})

	t.Run("invalid quantity error", func(t *testing.T) {
		transaction := &Transaction{
			Type:     TransactionTypeBuy,
			Symbol:   "AAPL",
			Quantity: decimal.Zero,
			Price:    &price,
		}

		err := transaction.Validate()

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidQuantity, err)
	})

	t.Run("invalid transaction type error", func(t *testing.T) {
		transaction := &Transaction{
			Type:     "INVALID",
			Symbol:   "AAPL",
			Quantity: decimal.NewFromInt(10),
			Price:    &price,
		}

		err := transaction.Validate()

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidTransactionType, err)
	})

	t.Run("buy without price error", func(t *testing.T) {
		transaction := &Transaction{
			Type:     TransactionTypeBuy,
			Symbol:   "AAPL",
			Quantity: decimal.NewFromInt(10),
		}

		err := transaction.Validate()

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidPrice, err)
	})

	t.Run("negative commission error", func(t *testing.T) {
		transaction := &Transaction{
			Type:       TransactionTypeBuy,
			Symbol:     "AAPL",
			Quantity:   decimal.NewFromInt(10),
			Price:      &price,
			Commission: decimal.NewFromFloat(-1.00),
		}

		err := transaction.Validate()

		assert.Error(t, err)
	})
}

func TestTransaction_isValidTransactionType(t *testing.T) {
	validTypes := []TransactionType{
		TransactionTypeBuy,
		TransactionTypeSell,
		TransactionTypeDividend,
		TransactionTypeSplit,
		TransactionTypeMerger,
		TransactionTypeSpinoff,
		TransactionTypeDividendReinvest,
	}

	for _, tt := range validTypes {
		transaction := &Transaction{Type: tt}
		assert.True(t, transaction.isValidTransactionType(), "Type %s should be valid", tt)
	}

	transaction := &Transaction{Type: "INVALID"}
	assert.False(t, transaction.isValidTransactionType())
}

func TestTransaction_IsBuy(t *testing.T) {
	t.Run("buy is true", func(t *testing.T) {
		transaction := &Transaction{Type: TransactionTypeBuy}
		assert.True(t, transaction.IsBuy())
	})

	t.Run("dividend reinvest is true", func(t *testing.T) {
		transaction := &Transaction{Type: TransactionTypeDividendReinvest}
		assert.True(t, transaction.IsBuy())
	})

	t.Run("sell is false", func(t *testing.T) {
		transaction := &Transaction{Type: TransactionTypeSell}
		assert.False(t, transaction.IsBuy())
	})
}

func TestTransaction_IsSell(t *testing.T) {
	t.Run("sell is true", func(t *testing.T) {
		transaction := &Transaction{Type: TransactionTypeSell}
		assert.True(t, transaction.IsSell())
	})

	t.Run("buy is false", func(t *testing.T) {
		transaction := &Transaction{Type: TransactionTypeBuy}
		assert.False(t, transaction.IsSell())
	})
}

func TestTransaction_GetTotalCost(t *testing.T) {
	t.Run("with price", func(t *testing.T) {
		price := decimal.NewFromFloat(150.00)
		transaction := &Transaction{
			Quantity:   decimal.NewFromInt(10),
			Price:      &price,
			Commission: decimal.NewFromFloat(1.00),
		}

		totalCost := transaction.GetTotalCost()

		expected := decimal.NewFromFloat(1501.00) // (10 * 150.00) + 1.00
		assert.True(t, expected.Equal(totalCost))
	})

	t.Run("without price", func(t *testing.T) {
		transaction := &Transaction{
			Quantity:   decimal.NewFromInt(10),
			Commission: decimal.NewFromFloat(1.00),
		}

		totalCost := transaction.GetTotalCost()

		assert.True(t, decimal.NewFromFloat(1.00).Equal(totalCost))
	})
}

func TestTransaction_GetProceeds(t *testing.T) {
	t.Run("with price", func(t *testing.T) {
		price := decimal.NewFromFloat(160.00)
		transaction := &Transaction{
			Quantity:   decimal.NewFromInt(10),
			Price:      &price,
			Commission: decimal.NewFromFloat(1.00),
		}

		proceeds := transaction.GetProceeds()

		expected := decimal.NewFromFloat(1599.00) // (10 * 160.00) - 1.00
		assert.True(t, expected.Equal(proceeds))
	})

	t.Run("without price", func(t *testing.T) {
		transaction := &Transaction{
			Quantity:   decimal.NewFromInt(10),
			Commission: decimal.NewFromFloat(1.00),
		}

		proceeds := transaction.GetProceeds()

		assert.True(t, decimal.Zero.Equal(proceeds))
	})
}

func TestTransaction_TableName(t *testing.T) {
	transaction := Transaction{}
	assert.Equal(t, "transactions", transaction.TableName())
}
