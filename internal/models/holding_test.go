package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestHolding_BeforeCreate(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	holding := &Holding{
		PortfolioID:  uuid.New(),
		Symbol:       "AAPL",
		Quantity:     decimal.NewFromInt(10),
		CostBasis:    decimal.NewFromFloat(1500.00),
		AvgCostPrice: decimal.NewFromFloat(150.00),
	}

	err := holding.BeforeCreate(db)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, holding.ID)
	assert.False(t, holding.UpdatedAt.IsZero())
}

func TestHolding_BeforeUpdate(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	holding := &Holding{}
	err := holding.BeforeUpdate(db)

	assert.NoError(t, err)
	assert.False(t, holding.UpdatedAt.IsZero())
}

func TestHolding_CalculateAvgCostPrice(t *testing.T) {
	t.Run("normal calculation", func(t *testing.T) {
		holding := &Holding{
			Quantity:  decimal.NewFromInt(10),
			CostBasis: decimal.NewFromFloat(1500.00),
		}

		holding.CalculateAvgCostPrice()

		expected := decimal.NewFromFloat(150.00)
		assert.True(t, expected.Equal(holding.AvgCostPrice))
	})

	t.Run("zero quantity", func(t *testing.T) {
		holding := &Holding{
			Quantity:  decimal.Zero,
			CostBasis: decimal.NewFromFloat(1500.00),
		}

		holding.CalculateAvgCostPrice()

		assert.True(t, decimal.Zero.Equal(holding.AvgCostPrice))
	})
}

func TestHolding_AddShares(t *testing.T) {
	holding := &Holding{
		Quantity:     decimal.NewFromInt(10),
		CostBasis:    decimal.NewFromFloat(1500.00),
		AvgCostPrice: decimal.NewFromFloat(150.00),
	}

	holding.AddShares(decimal.NewFromInt(5), decimal.NewFromFloat(800.00))

	assert.True(t, decimal.NewFromInt(15).Equal(holding.Quantity))
	assert.True(t, decimal.NewFromFloat(2300.00).Equal(holding.CostBasis))
	// Average should be 2300/15 = 153.33...
	expectedAvg := decimal.NewFromFloat(2300.00).Div(decimal.NewFromInt(15))
	assert.True(t, expectedAvg.Equal(holding.AvgCostPrice))
}

func TestHolding_RemoveShares(t *testing.T) {
	t.Run("successful removal", func(t *testing.T) {
		holding := &Holding{
			Quantity:     decimal.NewFromInt(10),
			CostBasis:    decimal.NewFromFloat(1500.00),
			AvgCostPrice: decimal.NewFromFloat(150.00),
		}

		err := holding.RemoveShares(decimal.NewFromInt(5), decimal.NewFromFloat(750.00))

		assert.NoError(t, err)
		assert.True(t, decimal.NewFromInt(5).Equal(holding.Quantity))
		assert.True(t, decimal.NewFromFloat(750.00).Equal(holding.CostBasis))
	})

	t.Run("insufficient shares error", func(t *testing.T) {
		holding := &Holding{
			Quantity:  decimal.NewFromInt(5),
			CostBasis: decimal.NewFromFloat(750.00),
		}

		err := holding.RemoveShares(decimal.NewFromInt(10), decimal.NewFromFloat(1500.00))

		assert.Error(t, err)
		assert.Equal(t, ErrInsufficientShares, err)
	})
}

func TestHolding_CalculateUnrealizedGain(t *testing.T) {
	holding := &Holding{
		Quantity:  decimal.NewFromInt(10),
		CostBasis: decimal.NewFromFloat(1500.00),
	}

	marketPrice := decimal.NewFromFloat(160.00)
	gain := holding.CalculateUnrealizedGain(marketPrice)

	// Current value: 10 * 160 = 1600
	// Cost basis: 1500
	// Gain: 1600 - 1500 = 100
	expected := decimal.NewFromFloat(100.00)
	assert.True(t, expected.Equal(gain))
}

func TestHolding_CalculateUnrealizedGainPercent(t *testing.T) {
	t.Run("positive gain", func(t *testing.T) {
		holding := &Holding{
			Quantity:  decimal.NewFromInt(10),
			CostBasis: decimal.NewFromFloat(1500.00),
		}

		marketPrice := decimal.NewFromFloat(160.00)
		gainPercent := holding.CalculateUnrealizedGainPercent(marketPrice)

		// Gain: 100 / 1500 * 100 = 6.666...%
		expected := decimal.NewFromFloat(100.00).Div(decimal.NewFromFloat(1500.00)).Mul(decimal.NewFromInt(100))
		assert.True(t, expected.Equal(gainPercent))
	})

	t.Run("zero cost basis", func(t *testing.T) {
		holding := &Holding{
			Quantity:  decimal.NewFromInt(10),
			CostBasis: decimal.Zero,
		}

		marketPrice := decimal.NewFromFloat(160.00)
		gainPercent := holding.CalculateUnrealizedGainPercent(marketPrice)

		assert.True(t, decimal.Zero.Equal(gainPercent))
	})
}

func TestHolding_TableName(t *testing.T) {
	holding := Holding{}
	assert.Equal(t, "holdings", holding.TableName())
}
