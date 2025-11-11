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

func TestPerformanceSnapshot_TableName(t *testing.T) {
	snapshot := &PerformanceSnapshot{}
	assert.Equal(t, "performance_snapshots", snapshot.TableName())
}

func TestPerformanceSnapshot_BeforeCreate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	snapshot := &PerformanceSnapshot{
		PortfolioID:    uuid.New(),
		Date:           time.Now().UTC(),
		TotalValue:     decimal.NewFromInt(10000),
		TotalCostBasis: decimal.NewFromInt(8000),
		TotalReturn:    decimal.NewFromInt(2000),
		TotalReturnPct: decimal.NewFromFloat(25.0),
	}

	// ID and CreatedAt should be set
	err = snapshot.BeforeCreate(db)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, snapshot.ID)
	assert.False(t, snapshot.CreatedAt.IsZero())
}

func TestPerformanceSnapshot_Validate(t *testing.T) {
	portfolioID := uuid.New()

	t.Run("valid snapshot", func(t *testing.T) {
		snapshot := &PerformanceSnapshot{
			PortfolioID:    portfolioID,
			Date:           time.Now().UTC(),
			TotalValue:     decimal.NewFromInt(10000),
			TotalCostBasis: decimal.NewFromInt(8000),
			TotalReturn:    decimal.NewFromInt(2000),
			TotalReturnPct: decimal.NewFromFloat(25.0),
		}
		assert.NoError(t, snapshot.Validate())
	})

	t.Run("invalid portfolio ID", func(t *testing.T) {
		snapshot := &PerformanceSnapshot{
			PortfolioID:    uuid.Nil,
			Date:           time.Now().UTC(),
			TotalValue:     decimal.NewFromInt(10000),
			TotalCostBasis: decimal.NewFromInt(8000),
		}
		assert.Equal(t, ErrInvalidPortfolioID, snapshot.Validate())
	})

	t.Run("invalid date", func(t *testing.T) {
		snapshot := &PerformanceSnapshot{
			PortfolioID:    portfolioID,
			Date:           time.Time{},
			TotalValue:     decimal.NewFromInt(10000),
			TotalCostBasis: decimal.NewFromInt(8000),
		}
		assert.Equal(t, ErrInvalidDate, snapshot.Validate())
	})

	t.Run("negative total value", func(t *testing.T) {
		snapshot := &PerformanceSnapshot{
			PortfolioID:    portfolioID,
			Date:           time.Now().UTC(),
			TotalValue:     decimal.NewFromInt(-1000),
			TotalCostBasis: decimal.NewFromInt(8000),
		}
		assert.Equal(t, ErrInvalidValue, snapshot.Validate())
	})

	t.Run("negative cost basis", func(t *testing.T) {
		snapshot := &PerformanceSnapshot{
			PortfolioID:    portfolioID,
			Date:           time.Now().UTC(),
			TotalValue:     decimal.NewFromInt(10000),
			TotalCostBasis: decimal.NewFromInt(-8000),
		}
		assert.Equal(t, ErrInvalidValue, snapshot.Validate())
	})
}

func TestPerformanceSnapshot_CalculateMetrics(t *testing.T) {
	t.Run("positive return", func(t *testing.T) {
		snapshot := &PerformanceSnapshot{
			TotalValue:     decimal.NewFromInt(12000),
			TotalCostBasis: decimal.NewFromInt(10000),
		}
		snapshot.CalculateMetrics()

		expectedReturn := decimal.NewFromInt(2000)
		expectedReturnPct := decimal.NewFromInt(20)

		assert.True(t, snapshot.TotalReturn.Equal(expectedReturn), "Expected return %s, got %s", expectedReturn, snapshot.TotalReturn)
		assert.True(t, snapshot.TotalReturnPct.Equal(expectedReturnPct), "Expected return %% %s, got %s", expectedReturnPct, snapshot.TotalReturnPct)
	})

	t.Run("negative return", func(t *testing.T) {
		snapshot := &PerformanceSnapshot{
			TotalValue:     decimal.NewFromInt(8000),
			TotalCostBasis: decimal.NewFromInt(10000),
		}
		snapshot.CalculateMetrics()

		expectedReturn := decimal.NewFromInt(-2000)
		expectedReturnPct := decimal.NewFromInt(-20)

		assert.True(t, snapshot.TotalReturn.Equal(expectedReturn))
		assert.True(t, snapshot.TotalReturnPct.Equal(expectedReturnPct))
	})

	t.Run("zero cost basis", func(t *testing.T) {
		snapshot := &PerformanceSnapshot{
			TotalValue:     decimal.NewFromInt(10000),
			TotalCostBasis: decimal.Zero,
		}
		snapshot.CalculateMetrics()

		assert.True(t, snapshot.TotalReturnPct.Equal(decimal.Zero))
	})

	t.Run("exact breakeven", func(t *testing.T) {
		snapshot := &PerformanceSnapshot{
			TotalValue:     decimal.NewFromInt(10000),
			TotalCostBasis: decimal.NewFromInt(10000),
		}
		snapshot.CalculateMetrics()

		assert.True(t, snapshot.TotalReturn.Equal(decimal.Zero))
		assert.True(t, snapshot.TotalReturnPct.Equal(decimal.Zero))
	})
}

func TestPerformanceSnapshot_CalculateDayChange(t *testing.T) {
	t.Run("positive day change", func(t *testing.T) {
		snapshot := &PerformanceSnapshot{
			TotalValue: decimal.NewFromInt(11000),
		}
		previousValue := decimal.NewFromInt(10000)
		snapshot.CalculateDayChange(previousValue)

		assert.NotNil(t, snapshot.DayChange)
		assert.NotNil(t, snapshot.DayChangePct)

		expectedChange := decimal.NewFromInt(1000)
		expectedChangePct := decimal.NewFromInt(10)

		assert.True(t, snapshot.DayChange.Equal(expectedChange))
		assert.True(t, snapshot.DayChangePct.Equal(expectedChangePct))
	})

	t.Run("negative day change", func(t *testing.T) {
		snapshot := &PerformanceSnapshot{
			TotalValue: decimal.NewFromInt(9000),
		}
		previousValue := decimal.NewFromInt(10000)
		snapshot.CalculateDayChange(previousValue)

		assert.NotNil(t, snapshot.DayChange)
		assert.NotNil(t, snapshot.DayChangePct)

		expectedChange := decimal.NewFromInt(-1000)
		expectedChangePct := decimal.NewFromInt(-10)

		assert.True(t, snapshot.DayChange.Equal(expectedChange))
		assert.True(t, snapshot.DayChangePct.Equal(expectedChangePct))
	})

	t.Run("zero previous value", func(t *testing.T) {
		snapshot := &PerformanceSnapshot{
			TotalValue: decimal.NewFromInt(10000),
		}
		previousValue := decimal.Zero
		snapshot.CalculateDayChange(previousValue)

		assert.NotNil(t, snapshot.DayChange)
		assert.NotNil(t, snapshot.DayChangePct)

		assert.True(t, snapshot.DayChange.Equal(decimal.NewFromInt(10000)))
		assert.True(t, snapshot.DayChangePct.Equal(decimal.Zero))
	})

	t.Run("no change", func(t *testing.T) {
		snapshot := &PerformanceSnapshot{
			TotalValue: decimal.NewFromInt(10000),
		}
		previousValue := decimal.NewFromInt(10000)
		snapshot.CalculateDayChange(previousValue)

		assert.NotNil(t, snapshot.DayChange)
		assert.NotNil(t, snapshot.DayChangePct)

		assert.True(t, snapshot.DayChange.Equal(decimal.Zero))
		assert.True(t, snapshot.DayChangePct.Equal(decimal.Zero))
	})
}
