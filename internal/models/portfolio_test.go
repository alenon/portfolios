package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPortfolio_BeforeCreate(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	t.Run("sets defaults", func(t *testing.T) {
		portfolio := &Portfolio{
			UserID: uuid.New(),
			Name:   "Test",
		}

		err := portfolio.BeforeCreate(db)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, portfolio.ID)
		assert.False(t, portfolio.CreatedAt.IsZero())
		assert.False(t, portfolio.UpdatedAt.IsZero())
		assert.Equal(t, "USD", portfolio.BaseCurrency)
		assert.Equal(t, CostBasisFIFO, portfolio.CostBasisMethod)
	})

	t.Run("does not override existing values", func(t *testing.T) {
		id := uuid.New()
		portfolio := &Portfolio{
			ID:              id,
			UserID:          uuid.New(),
			Name:            "Test",
			BaseCurrency:    "EUR",
			CostBasisMethod: CostBasisLIFO,
		}

		err := portfolio.BeforeCreate(db)

		assert.NoError(t, err)
		assert.Equal(t, id, portfolio.ID)
		assert.Equal(t, "EUR", portfolio.BaseCurrency)
		assert.Equal(t, CostBasisLIFO, portfolio.CostBasisMethod)
	})
}

func TestPortfolio_BeforeUpdate(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	portfolio := &Portfolio{
		UserID: uuid.New(),
		Name:   "Test",
	}

	err := portfolio.BeforeUpdate(db)

	assert.NoError(t, err)
	assert.False(t, portfolio.UpdatedAt.IsZero())
}

func TestPortfolio_Validate(t *testing.T) {
	t.Run("valid portfolio", func(t *testing.T) {
		portfolio := &Portfolio{
			Name:            "Test",
			BaseCurrency:    "USD",
			CostBasisMethod: CostBasisFIFO,
		}

		err := portfolio.Validate()

		assert.NoError(t, err)
	})

	t.Run("empty name error", func(t *testing.T) {
		portfolio := &Portfolio{
			Name:            "",
			BaseCurrency:    "USD",
			CostBasisMethod: CostBasisFIFO,
		}

		err := portfolio.Validate()

		assert.Error(t, err)
		assert.Equal(t, ErrPortfolioNameRequired, err)
	})

	t.Run("invalid currency error", func(t *testing.T) {
		portfolio := &Portfolio{
			Name:            "Test",
			BaseCurrency:    "US", // Too short
			CostBasisMethod: CostBasisFIFO,
		}

		err := portfolio.Validate()

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCurrency, err)
	})

	t.Run("invalid cost basis method error", func(t *testing.T) {
		portfolio := &Portfolio{
			Name:            "Test",
			BaseCurrency:    "USD",
			CostBasisMethod: "INVALID",
		}

		err := portfolio.Validate()

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCostBasisMethod, err)
	})
}

func TestPortfolio_isValidCostBasisMethod(t *testing.T) {
	t.Run("FIFO is valid", func(t *testing.T) {
		portfolio := &Portfolio{CostBasisMethod: CostBasisFIFO}
		assert.True(t, portfolio.isValidCostBasisMethod())
	})

	t.Run("LIFO is valid", func(t *testing.T) {
		portfolio := &Portfolio{CostBasisMethod: CostBasisLIFO}
		assert.True(t, portfolio.isValidCostBasisMethod())
	})

	t.Run("SPECIFIC_LOT is valid", func(t *testing.T) {
		portfolio := &Portfolio{CostBasisMethod: CostBasisSpecificLot}
		assert.True(t, portfolio.isValidCostBasisMethod())
	})

	t.Run("invalid method returns false", func(t *testing.T) {
		portfolio := &Portfolio{CostBasisMethod: "INVALID"}
		assert.False(t, portfolio.isValidCostBasisMethod())
	})
}

func TestPortfolio_TableName(t *testing.T) {
	portfolio := Portfolio{}
	assert.Equal(t, "portfolios", portfolio.TableName())
}
