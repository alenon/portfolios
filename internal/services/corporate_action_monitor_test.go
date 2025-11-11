package services

import (
	"context"
	"testing"
	"time"

	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMonitorTestDB(t *testing.T) *gorm.DB {
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

func createMonitorTestData(t *testing.T, db *gorm.DB) (*models.Portfolio, *models.Holding, *models.CorporateAction) {
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

	return portfolio, holding, action
}

func TestNewCorporateActionMonitor(t *testing.T) {
	db := setupMonitorTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.corporateActionRepo)
	assert.NotNil(t, monitor.portfolioRepo)
	assert.NotNil(t, monitor.holdingRepo)
	assert.NotNil(t, monitor.portfolioActionRepo)
}

func TestDetectAndSuggestActions_NoActions(t *testing.T) {
	db := setupMonitorTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	ctx := context.Background()
	err := monitor.DetectAndSuggestActions(ctx)

	assert.NoError(t, err)
}

func TestDetectAndSuggestActions_WithAction(t *testing.T) {
	db := setupMonitorTestDB(t)
	_, _, action := createMonitorTestData(t, db)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	ctx := context.Background()
	err := monitor.DetectAndSuggestActions(ctx)
	assert.NoError(t, err)

	// Verify portfolio action was created
	actions, err := portfolioActionRepo.FindPendingByCorporateActionID(action.ID.String())
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
	assert.Equal(t, "AAPL", actions[0].AffectedSymbol)
	assert.Equal(t, int64(100), actions[0].SharesAffected)
}

func TestDetectAndSuggestActions_ContextCancellation(t *testing.T) {
	db := setupMonitorTestDB(t)
	createMonitorTestData(t, db)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := monitor.DetectAndSuggestActions(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestDetectAndSuggestActions_NoDuplicates(t *testing.T) {
	db := setupMonitorTestDB(t)
	portfolio, _, action := createMonitorTestData(t, db)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	// Create existing pending action
	existingAction := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: action.ID,
		Status:            models.PortfolioActionStatusPending,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC(),
	}
	require.NoError(t, portfolioActionRepo.Create(existingAction))

	// Run detection
	ctx := context.Background()
	err := monitor.DetectAndSuggestActions(ctx)
	assert.NoError(t, err)

	// Should still only have one action (no duplicate created)
	actions, err := portfolioActionRepo.FindPendingByCorporateActionID(action.ID.String())
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
}

func TestProcessAction_NoHoldings(t *testing.T) {
	db := setupMonitorTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	// Create action for symbol with no holdings
	ratio := decimal.NewFromFloat(2.0)
	action := &models.CorporateAction{
		Symbol:  "TSLA",
		Type:    models.CorporateActionTypeSplit,
		Date:    time.Now().UTC(),
		Ratio:   &ratio,
		Applied: false,
	}
	require.NoError(t, db.Create(action).Error)

	err := monitor.processAction(action)
	assert.NoError(t, err)

	// No portfolio actions should be created
	actions, err := portfolioActionRepo.FindPendingByCorporateActionID(action.ID.String())
	assert.NoError(t, err)
	assert.Len(t, actions, 0)
}

func TestFindPortfoliosWithSymbol(t *testing.T) {
	db := setupMonitorTestDB(t)
	createMonitorTestData(t, db)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	portfolios, err := monitor.findPortfoliosWithSymbol("AAPL")
	assert.NoError(t, err)
	assert.Len(t, portfolios, 1)
	assert.Equal(t, "Test Portfolio", portfolios[0].Name)
}

func TestFindPortfoliosWithSymbol_NotFound(t *testing.T) {
	db := setupMonitorTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	portfolios, err := monitor.findPortfoliosWithSymbol("TSLA")
	assert.NoError(t, err)
	assert.Len(t, portfolios, 0)
}

func TestGenerateActionDescription_Split(t *testing.T) {
	db := setupMonitorTestDB(t)
	_, holding, _ := createMonitorTestData(t, db)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	ratio := decimal.NewFromFloat(2.0)
	action := &models.CorporateAction{
		Symbol: "AAPL",
		Type:   models.CorporateActionTypeSplit,
		Date:   time.Now().UTC(),
		Ratio:  &ratio,
	}

	desc := monitor.generateActionDescription(action, holding)
	assert.Contains(t, desc, "Stock split")
	assert.Contains(t, desc, "AAPL")
	assert.Contains(t, desc, "100")
}

func TestGenerateActionDescription_Dividend(t *testing.T) {
	db := setupMonitorTestDB(t)
	_, holding, _ := createMonitorTestData(t, db)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	amount := decimal.NewFromFloat(0.25)
	action := &models.CorporateAction{
		Symbol: "AAPL",
		Type:   models.CorporateActionTypeDividend,
		Date:   time.Now().UTC(),
		Amount: &amount,
	}

	desc := monitor.generateActionDescription(action, holding)
	assert.Contains(t, desc, "Dividend")
	assert.Contains(t, desc, "AAPL")
	assert.Contains(t, desc, "0.25")
}

func TestGenerateActionDescription_Merger(t *testing.T) {
	db := setupMonitorTestDB(t)
	_, holding, _ := createMonitorTestData(t, db)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	ratio := decimal.NewFromFloat(1.5)
	newSymbol := "ABC"
	action := &models.CorporateAction{
		Symbol:    "AAPL",
		Type:      models.CorporateActionTypeMerger,
		Date:      time.Now().UTC(),
		Ratio:     &ratio,
		NewSymbol: &newSymbol,
	}

	desc := monitor.generateActionDescription(action, holding)
	assert.Contains(t, desc, "Merger")
	assert.Contains(t, desc, "AAPL")
	assert.Contains(t, desc, "ABC")
}

func TestGenerateActionDescription_Spinoff(t *testing.T) {
	db := setupMonitorTestDB(t)
	_, holding, _ := createMonitorTestData(t, db)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	ratio := decimal.NewFromFloat(0.5)
	newSymbol := "SPIN"
	action := &models.CorporateAction{
		Symbol:    "AAPL",
		Type:      models.CorporateActionTypeSpinoff,
		Date:      time.Now().UTC(),
		Ratio:     &ratio,
		NewSymbol: &newSymbol,
	}

	desc := monitor.generateActionDescription(action, holding)
	assert.Contains(t, desc, "Spinoff")
	assert.Contains(t, desc, "AAPL")
	assert.Contains(t, desc, "SPIN")
}

func TestGenerateActionDescription_TickerChange(t *testing.T) {
	db := setupMonitorTestDB(t)
	_, holding, _ := createMonitorTestData(t, db)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	newSymbol := "APPL"
	action := &models.CorporateAction{
		Symbol:    "AAPL",
		Type:      models.CorporateActionTypeTickerChange,
		Date:      time.Now().UTC(),
		NewSymbol: &newSymbol,
	}

	desc := monitor.generateActionDescription(action, holding)
	assert.Contains(t, desc, "Ticker change")
	assert.Contains(t, desc, "AAPL")
	assert.Contains(t, desc, "APPL")
}

func TestSimulateDetection_CreateNew(t *testing.T) {
	db := setupMonitorTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	ratio := decimal.NewFromFloat(2.0)
	action, err := monitor.SimulateDetection(
		"AAPL",
		models.CorporateActionTypeSplit,
		time.Now().UTC(),
		&ratio,
		nil,
		nil,
		"2:1 stock split",
	)

	assert.NoError(t, err)
	assert.NotNil(t, action)
	assert.Equal(t, "AAPL", action.Symbol)
	assert.Equal(t, models.CorporateActionTypeSplit, action.Type)
	assert.False(t, action.Applied)
}

func TestSimulateDetection_Duplicate(t *testing.T) {
	db := setupMonitorTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	ratio := decimal.NewFromFloat(2.0)
	date := time.Now().UTC()

	// Create first action
	action1, err := monitor.SimulateDetection(
		"AAPL",
		models.CorporateActionTypeSplit,
		date,
		&ratio,
		nil,
		nil,
		"2:1 stock split",
	)
	assert.NoError(t, err)

	// Try to create duplicate
	action2, err := monitor.SimulateDetection(
		"AAPL",
		models.CorporateActionTypeSplit,
		date,
		&ratio,
		nil,
		nil,
		"2:1 stock split",
	)
	assert.NoError(t, err)
	assert.Equal(t, action1.ID, action2.ID) // Should return existing action
}

func TestSimulateDetection_InvalidAction(t *testing.T) {
	db := setupMonitorTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)

	monitor := NewCorporateActionMonitor(
		corporateActionRepo,
		portfolioRepo,
		holdingRepo,
		portfolioActionRepo,
	)

	// Split without ratio should fail validation
	_, err := monitor.SimulateDetection(
		"AAPL",
		models.CorporateActionTypeSplit,
		time.Now().UTC(),
		nil, // Missing ratio
		nil,
		nil,
		"Invalid split",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}
