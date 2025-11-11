package services

import (
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

func setupServiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.User{},
		&models.Portfolio{},
		&models.Transaction{},
		&models.Holding{},
		&models.TaxLot{},
		&models.CorporateAction{},
	)
	require.NoError(t, err)

	return db
}

func createServiceTestData(t *testing.T, db *gorm.DB) (*models.User, *models.Portfolio) {
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

	return user, portfolio
}

func TestNewCorporateActionService(t *testing.T) {
	db := setupServiceTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	assert.NotNil(t, service)
}

func TestCorporateActionService_Create_Success(t *testing.T) {
	db := setupServiceTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	ratio := decimal.NewFromFloat(2.0)
	desc := "2:1 stock split"
	action, err := service.Create(
		"AAPL",
		models.CorporateActionTypeSplit,
		time.Now().UTC(),
		&ratio,
		nil,
		nil,
		nil,
		&desc,
	)

	assert.NoError(t, err)
	assert.NotNil(t, action)
	assert.Equal(t, "AAPL", action.Symbol)
	assert.Equal(t, models.CorporateActionTypeSplit, action.Type)
	assert.False(t, action.Applied)
}

func TestCorporateActionService_Create_ValidationError(t *testing.T) {
	db := setupServiceTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	desc := "Invalid split"
	// Split without ratio should fail validation
	_, err := service.Create(
		"AAPL",
		models.CorporateActionTypeSplit,
		time.Now().UTC(),
		nil, // Missing required ratio
		nil,
		nil,
		nil,
		&desc,
	)

	assert.Error(t, err)
}

func TestCorporateActionService_GetByID_Success(t *testing.T) {
	db := setupServiceTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	ratio := decimal.NewFromFloat(2.0)
	desc := "2:1 stock split"
	created, err := service.Create(
		"AAPL",
		models.CorporateActionTypeSplit,
		time.Now().UTC(),
		&ratio,
		nil,
		nil,
		nil,
		&desc,
	)
	require.NoError(t, err)

	found, err := service.GetByID(created.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "AAPL", found.Symbol)
}

func TestCorporateActionService_GetByID_NotFound(t *testing.T) {
	db := setupServiceTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	_, err := service.GetByID("00000000-0000-0000-0000-000000000000")
	assert.Error(t, err)
}

func TestCorporateActionService_GetBySymbol(t *testing.T) {
	db := setupServiceTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	ratio := decimal.NewFromFloat(2.0)
	desc := "2:1 stock split"
	_, err := service.Create(
		"AAPL",
		models.CorporateActionTypeSplit,
		time.Now().UTC(),
		&ratio,
		nil,
		nil,
		nil,
		&desc,
	)
	require.NoError(t, err)

	actions, err := service.GetBySymbol("AAPL")
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
	assert.Equal(t, "AAPL", actions[0].Symbol)
}

func TestCorporateActionService_GetBySymbolAndDateRange(t *testing.T) {
	db := setupServiceTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	now := time.Now().UTC()
	ratio := decimal.NewFromFloat(2.0)
	desc := "2:1 stock split"
	_, err := service.Create(
		"AAPL",
		models.CorporateActionTypeSplit,
		now,
		&ratio,
		nil,
		nil,
		nil,
		&desc,
	)
	require.NoError(t, err)

	// Query within range
	startDate := now.AddDate(0, 0, -1)
	endDate := now.AddDate(0, 0, 1)
	actions, err := service.GetBySymbolAndDateRange("AAPL", startDate, endDate)
	assert.NoError(t, err)
	assert.Len(t, actions, 1)

	// Query outside range
	startDate = now.AddDate(0, 0, -10)
	endDate = now.AddDate(0, 0, -5)
	actions, err = service.GetBySymbolAndDateRange("AAPL", startDate, endDate)
	assert.NoError(t, err)
	assert.Len(t, actions, 0)
}

func TestCorporateActionService_GetUnapplied(t *testing.T) {
	db := setupServiceTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	ratio := decimal.NewFromFloat(2.0)
	desc := "2:1 stock split"
	action, err := service.Create(
		"AAPL",
		models.CorporateActionTypeSplit,
		time.Now().UTC(),
		&ratio,
		nil,
		nil,
		nil,
		&desc,
	)
	require.NoError(t, err)

	unapplied, err := service.GetUnapplied()
	assert.NoError(t, err)
	assert.Len(t, unapplied, 1)
	assert.Equal(t, action.ID, unapplied[0].ID)
	assert.False(t, unapplied[0].Applied)
}

func TestCorporateActionService_MarkAsApplied(t *testing.T) {
	db := setupServiceTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	ratio := decimal.NewFromFloat(2.0)
	desc := "2:1 stock split"
	action, err := service.Create(
		"AAPL",
		models.CorporateActionTypeSplit,
		time.Now().UTC(),
		&ratio,
		nil,
		nil,
		nil,
		&desc,
	)
	require.NoError(t, err)
	assert.False(t, action.Applied)

	err = service.MarkAsApplied(action.ID.String())
	assert.NoError(t, err)

	// Verify it was marked
	updated, err := service.GetByID(action.ID.String())
	assert.NoError(t, err)
	assert.True(t, updated.Applied)

	// Should not be in unapplied list
	unapplied, err := service.GetUnapplied()
	assert.NoError(t, err)
	assert.Len(t, unapplied, 0)
}

func TestCorporateActionService_Delete(t *testing.T) {
	db := setupServiceTestDB(t)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	ratio := decimal.NewFromFloat(2.0)
	desc := "2:1 stock split"
	action, err := service.Create(
		"AAPL",
		models.CorporateActionTypeSplit,
		time.Now().UTC(),
		&ratio,
		nil,
		nil,
		nil,
		&desc,
	)
	require.NoError(t, err)

	err = service.Delete(action.ID.String())
	assert.NoError(t, err)

	// Should not be findable
	_, err = service.GetByID(action.ID.String())
	assert.Error(t, err)
}

func TestCorporateActionService_ApplyStockSplit_PortfolioNotFound(t *testing.T) {
	db := setupServiceTestDB(t)
	user, _ := createServiceTestData(t, db)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	ratio := decimal.NewFromFloat(2.0)
	err := service.ApplyStockSplit(
		"00000000-0000-0000-0000-000000000000",
		"AAPL",
		user.ID.String(),
		ratio,
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Equal(t, models.ErrPortfolioNotFound, err)
}

func TestCorporateActionService_ApplyStockSplit_Unauthorized(t *testing.T) {
	db := setupServiceTestDB(t)
	_, portfolio := createServiceTestData(t, db)

	// Create another user
	otherUser := &models.User{
		Email:        "other@example.com",
		PasswordHash: "hash",
	}
	require.NoError(t, db.Create(otherUser).Error)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	ratio := decimal.NewFromFloat(2.0)
	err := service.ApplyStockSplit(
		portfolio.ID.String(),
		"AAPL",
		otherUser.ID.String(),
		ratio,
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)
}

func TestCorporateActionService_ApplyStockSplit_NotImplemented(t *testing.T) {
	db := setupServiceTestDB(t)
	user, portfolio := createServiceTestData(t, db)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	ratio := decimal.NewFromFloat(2.0)
	err := service.ApplyStockSplit(
		portfolio.ID.String(),
		"AAPL",
		user.ID.String(),
		ratio,
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}

func TestCorporateActionService_ApplyDividend_PortfolioNotFound(t *testing.T) {
	db := setupServiceTestDB(t)
	user, _ := createServiceTestData(t, db)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	amount := decimal.NewFromFloat(0.25)
	err := service.ApplyDividend(
		"00000000-0000-0000-0000-000000000000",
		"AAPL",
		user.ID.String(),
		amount,
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Equal(t, models.ErrPortfolioNotFound, err)
}

func TestCorporateActionService_ApplyDividend_Unauthorized(t *testing.T) {
	db := setupServiceTestDB(t)
	_, portfolio := createServiceTestData(t, db)

	otherUser := &models.User{
		Email:        "other@example.com",
		PasswordHash: "hash",
	}
	require.NoError(t, db.Create(otherUser).Error)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	amount := decimal.NewFromFloat(0.25)
	err := service.ApplyDividend(
		portfolio.ID.String(),
		"AAPL",
		otherUser.ID.String(),
		amount,
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)
}

func TestCorporateActionService_ApplyDividend_NotImplemented(t *testing.T) {
	db := setupServiceTestDB(t)
	user, portfolio := createServiceTestData(t, db)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	amount := decimal.NewFromFloat(0.25)
	err := service.ApplyDividend(
		portfolio.ID.String(),
		"AAPL",
		user.ID.String(),
		amount,
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}

func TestCorporateActionService_ApplyMerger_PortfolioNotFound(t *testing.T) {
	db := setupServiceTestDB(t)
	user, _ := createServiceTestData(t, db)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	ratio := decimal.NewFromFloat(1.5)
	err := service.ApplyMerger(
		"00000000-0000-0000-0000-000000000000",
		"AAPL",
		"XYZ",
		user.ID.String(),
		ratio,
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Equal(t, models.ErrPortfolioNotFound, err)
}

func TestCorporateActionService_ApplyMerger_Unauthorized(t *testing.T) {
	db := setupServiceTestDB(t)
	_, portfolio := createServiceTestData(t, db)

	otherUser := &models.User{
		Email:        "other@example.com",
		PasswordHash: "hash",
	}
	require.NoError(t, db.Create(otherUser).Error)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	ratio := decimal.NewFromFloat(1.5)
	err := service.ApplyMerger(
		portfolio.ID.String(),
		"AAPL",
		"XYZ",
		otherUser.ID.String(),
		ratio,
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)
}

func TestCorporateActionService_ApplyMerger_NotImplemented(t *testing.T) {
	db := setupServiceTestDB(t)
	user, portfolio := createServiceTestData(t, db)

	corporateActionRepo := repository.NewCorporateActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	taxLotRepo := repository.NewTaxLotRepository(db)

	service := NewCorporateActionService(
		corporateActionRepo,
		portfolioRepo,
		transactionRepo,
		holdingRepo,
		taxLotRepo,
	)

	ratio := decimal.NewFromFloat(1.5)
	err := service.ApplyMerger(
		portfolio.ID.String(),
		"AAPL",
		"XYZ",
		user.ID.String(),
		ratio,
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}
