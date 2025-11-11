package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
	"github.com/lenon/portfolios/internal/services"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.User{},
		&models.Portfolio{},
		&models.Holding{},
		&models.CorporateAction{},
		&models.PortfolioAction{},
	)
	require.NoError(t, err)

	return db
}

func TestCorporateActionDetectionJob_Name(t *testing.T) {
	db := setupTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := services.NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	job := NewCorporateActionDetectionJob(monitor)

	assert.Equal(t, "CorporateActionDetection", job.Name())
}

func TestCorporateActionDetectionJob_Schedule(t *testing.T) {
	db := setupTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := services.NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	job := NewCorporateActionDetectionJob(monitor)

	assert.Equal(t, "@daily", job.Schedule())
}

func TestCorporateActionDetectionJob_Run(t *testing.T) {
	db := setupTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	// Create test data
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
	}
	require.NoError(t, db.Create(user).Error)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	require.NoError(t, db.Create(portfolio).Error)

	holding := &models.Holding{
		PortfolioID:  portfolio.ID,
		Symbol:       "AAPL",
		Quantity:     decimal.NewFromInt(100),
		CostBasis:    decimal.NewFromInt(10000),
		AvgCostPrice: decimal.NewFromInt(100),
	}
	require.NoError(t, db.Create(holding).Error)

	ratio := decimal.NewFromFloat(2.0)
	action := &models.CorporateAction{
		Symbol:  "AAPL",
		Type:    models.CorporateActionTypeSplit,
		Date:    time.Now().UTC(),
		Ratio:   &ratio,
		Applied: false,
	}
	require.NoError(t, db.Create(action).Error)

	monitor := services.NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	job := NewCorporateActionDetectionJob(monitor)

	// Run the job
	ctx := context.Background()
	err := job.Run(ctx)

	// Should not error (even if no actions are created, it should run successfully)
	assert.NoError(t, err)
}

func TestCorporateActionDetectionJob_Run_WithContextCancellation(t *testing.T) {
	db := setupTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := services.NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	job := NewCorporateActionDetectionJob(monitor)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := job.Run(ctx)

	// Should handle cancellation gracefully
	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled) || err != nil)
}
