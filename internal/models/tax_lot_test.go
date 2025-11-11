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

func TestTaxLot_BeforeCreate(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	taxLot := &TaxLot{
		PortfolioID:   uuid.New(),
		Symbol:        "AAPL",
		PurchaseDate:  time.Now(),
		Quantity:      decimal.NewFromInt(10),
		CostBasis:     decimal.NewFromFloat(1500.00),
		TransactionID: uuid.New(),
	}

	err := taxLot.BeforeCreate(db)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, taxLot.ID)
	assert.False(t, taxLot.CreatedAt.IsZero())
	assert.False(t, taxLot.UpdatedAt.IsZero())
}

func TestTaxLot_BeforeUpdate(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	taxLot := &TaxLot{}
	err := taxLot.BeforeUpdate(db)

	assert.NoError(t, err)
	assert.False(t, taxLot.UpdatedAt.IsZero())
}

func TestTaxLot_GetCostPerShare(t *testing.T) {
	t.Run("normal calculation", func(t *testing.T) {
		taxLot := &TaxLot{
			Quantity:  decimal.NewFromInt(10),
			CostBasis: decimal.NewFromFloat(1500.00),
		}

		costPerShare := taxLot.GetCostPerShare()

		expected := decimal.NewFromFloat(150.00)
		assert.True(t, expected.Equal(costPerShare))
	})

	t.Run("zero quantity", func(t *testing.T) {
		taxLot := &TaxLot{
			Quantity:  decimal.Zero,
			CostBasis: decimal.NewFromFloat(1500.00),
		}

		costPerShare := taxLot.GetCostPerShare()

		assert.True(t, decimal.Zero.Equal(costPerShare))
	})
}

func TestTaxLot_CalculateGain(t *testing.T) {
	taxLot := &TaxLot{
		Quantity:  decimal.NewFromInt(10),
		CostBasis: decimal.NewFromFloat(1500.00),
	}

	salePrice := decimal.NewFromFloat(160.00)
	quantity := decimal.NewFromInt(5)

	gain := taxLot.CalculateGain(salePrice, quantity)

	// Cost basis for 5 shares: 150 * 5 = 750
	// Proceeds: 160 * 5 = 800
	// Gain: 800 - 750 = 50
	expected := decimal.NewFromFloat(50.00)
	assert.True(t, expected.Equal(gain))
}

func TestTaxLot_IsLongTerm(t *testing.T) {
	t.Run("short term holding", func(t *testing.T) {
		taxLot := &TaxLot{
			PurchaseDate: time.Now().AddDate(0, -6, 0), // 6 months ago
		}

		isLongTerm := taxLot.IsLongTerm(time.Now())

		assert.False(t, isLongTerm)
	})

	t.Run("exactly one year", func(t *testing.T) {
		purchaseDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		saleDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		taxLot := &TaxLot{
			PurchaseDate: purchaseDate,
		}

		isLongTerm := taxLot.IsLongTerm(saleDate)

		assert.True(t, isLongTerm)
	})

	t.Run("long term holding", func(t *testing.T) {
		taxLot := &TaxLot{
			PurchaseDate: time.Now().AddDate(-2, 0, 0), // 2 years ago
		}

		isLongTerm := taxLot.IsLongTerm(time.Now())

		assert.True(t, isLongTerm)
	})
}

func TestTaxLot_ReduceQuantity(t *testing.T) {
	t.Run("successful reduction", func(t *testing.T) {
		taxLot := &TaxLot{
			Quantity:  decimal.NewFromInt(10),
			CostBasis: decimal.NewFromFloat(1500.00),
		}

		err := taxLot.ReduceQuantity(decimal.NewFromInt(5))

		assert.NoError(t, err)
		assert.True(t, decimal.NewFromInt(5).Equal(taxLot.Quantity))
		assert.True(t, decimal.NewFromFloat(750.00).Equal(taxLot.CostBasis))
	})

	t.Run("insufficient shares error", func(t *testing.T) {
		taxLot := &TaxLot{
			Quantity:  decimal.NewFromInt(5),
			CostBasis: decimal.NewFromFloat(750.00),
		}

		err := taxLot.ReduceQuantity(decimal.NewFromInt(10))

		assert.Error(t, err)
		assert.Equal(t, ErrInsufficientShares, err)
	})
}

func TestTaxLot_TableName(t *testing.T) {
	taxLot := TaxLot{}
	assert.Equal(t, "tax_lots", taxLot.TableName())
}
