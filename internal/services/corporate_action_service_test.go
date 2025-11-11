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

// Removed: TestCorporateActionService_ApplyStockSplit_NotImplemented - no longer needed as functionality is implemented

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

// Removed: TestCorporateActionService_ApplyDividend_NotImplemented - no longer needed as functionality is implemented

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

// Removed: TestCorporateActionService_ApplyMerger_NotImplemented - no longer needed as functionality is implemented
func TestCorporateActionService_ApplySpinoff_Success(t *testing.T) {
	db := setupServiceTestDB(t)
	user, portfolio := createServiceTestData(t, db)

	// Create a holding and tax lots for parent company
	parentHolding := &models.Holding{
		PortfolioID:  portfolio.ID,
		Symbol:       "AAPL",
		Quantity:     decimal.NewFromInt(100),
		CostBasis:    decimal.NewFromInt(10000),
		AvgCostPrice: decimal.NewFromInt(100),
	}
	require.NoError(t, db.Create(parentHolding).Error)

	// Create transaction for tracking
	transaction := &models.Transaction{
		PortfolioID: portfolio.ID,
		Type:        models.TransactionTypeBuy,
		Symbol:      "AAPL",
		Quantity:    decimal.NewFromInt(100),
		Price:       &[]decimal.Decimal{decimal.NewFromInt(100)}[0],
		Date:        time.Now().UTC().Add(-30 * 24 * time.Hour),
	}
	require.NoError(t, db.Create(transaction).Error)

	// Create tax lot
	taxLot := &models.TaxLot{
		PortfolioID:   portfolio.ID,
		Symbol:        "AAPL",
		PurchaseDate:  time.Now().UTC().Add(-30 * 24 * time.Hour),
		Quantity:      decimal.NewFromInt(100),
		CostBasis:     decimal.NewFromInt(10000),
		TransactionID: transaction.ID,
	}
	require.NoError(t, db.Create(taxLot).Error)

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

	// Apply spinoff: 0.5 shares of SPIN for every 1 share of AAPL
	ratio := decimal.NewFromFloat(0.5)
	err := service.ApplySpinoff(
		portfolio.ID.String(),
		"AAPL",
		"SPIN",
		user.ID.String(),
		ratio,
		time.Now().UTC(),
	)

	assert.NoError(t, err)

	// Verify parent holding was updated (cost basis reduced by 10%)
	parentUpdated, err := holdingRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")
	assert.NoError(t, err)
	assert.True(t, parentUpdated.Quantity.Equal(decimal.NewFromInt(100)))   // Quantity unchanged
	assert.True(t, parentUpdated.CostBasis.Equal(decimal.NewFromInt(9000))) // 90% of original

	// Verify spinoff holding was created
	spinoffHolding, err := holdingRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "SPIN")
	assert.NoError(t, err)
	assert.True(t, spinoffHolding.Quantity.Equal(decimal.NewFromInt(50)))    // 100 * 0.5
	assert.True(t, spinoffHolding.CostBasis.Equal(decimal.NewFromInt(1000))) // 10% of original

	// Verify spinoff tax lot was created
	spinoffTaxLots, err := taxLotRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "SPIN")
	assert.NoError(t, err)
	assert.Len(t, spinoffTaxLots, 1)
	assert.True(t, spinoffTaxLots[0].Quantity.Equal(decimal.NewFromInt(50)))
	assert.Equal(t, taxLot.PurchaseDate.Unix(), spinoffTaxLots[0].PurchaseDate.Unix()) // Inherited purchase date

	// Verify parent tax lot was updated (cost basis reduced)
	parentTaxLots, err := taxLotRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")
	assert.NoError(t, err)
	assert.Len(t, parentTaxLots, 1)
	assert.True(t, parentTaxLots[0].CostBasis.Equal(decimal.NewFromInt(9000))) // Reduced by 10%
}

func TestCorporateActionService_ApplySpinoff_PortfolioNotFound(t *testing.T) {
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

	ratio := decimal.NewFromFloat(0.5)
	err := service.ApplySpinoff(
		"00000000-0000-0000-0000-000000000000",
		"AAPL",
		"SPIN",
		user.ID.String(),
		ratio,
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Equal(t, models.ErrPortfolioNotFound, err)
}

func TestCorporateActionService_ApplySpinoff_Unauthorized(t *testing.T) {
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

	ratio := decimal.NewFromFloat(0.5)
	err := service.ApplySpinoff(
		portfolio.ID.String(),
		"AAPL",
		"SPIN",
		otherUser.ID.String(),
		ratio,
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)
}

func TestCorporateActionService_ApplySpinoff_InvalidRatio(t *testing.T) {
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

	// Test with zero ratio
	err := service.ApplySpinoff(
		portfolio.ID.String(),
		"AAPL",
		"SPIN",
		user.ID.String(),
		decimal.Zero,
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid spinoff ratio")

	// Test with negative ratio
	err = service.ApplySpinoff(
		portfolio.ID.String(),
		"AAPL",
		"SPIN",
		user.ID.String(),
		decimal.NewFromInt(-1),
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid spinoff ratio")
}

func TestCorporateActionService_ApplySpinoff_NoHolding(t *testing.T) {
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

	ratio := decimal.NewFromFloat(0.5)
	err := service.ApplySpinoff(
		portfolio.ID.String(),
		"AAPL",
		"SPIN",
		user.ID.String(),
		ratio,
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no holding found")
}

func TestCorporateActionService_ApplyTickerChange_Success(t *testing.T) {
	db := setupServiceTestDB(t)
	user, portfolio := createServiceTestData(t, db)

	// Create holding with old ticker
	oldHolding := &models.Holding{
		PortfolioID:  portfolio.ID,
		Symbol:       "FB",
		Quantity:     decimal.NewFromInt(100),
		CostBasis:    decimal.NewFromInt(20000),
		AvgCostPrice: decimal.NewFromInt(200),
	}
	require.NoError(t, db.Create(oldHolding).Error)

	// Create transaction
	transaction := &models.Transaction{
		PortfolioID: portfolio.ID,
		Type:        models.TransactionTypeBuy,
		Symbol:      "FB",
		Quantity:    decimal.NewFromInt(100),
		Price:       &[]decimal.Decimal{decimal.NewFromInt(200)}[0],
		Date:        time.Now().UTC().Add(-60 * 24 * time.Hour),
	}
	require.NoError(t, db.Create(transaction).Error)

	// Create tax lot
	taxLot := &models.TaxLot{
		PortfolioID:   portfolio.ID,
		Symbol:        "FB",
		PurchaseDate:  time.Now().UTC().Add(-60 * 24 * time.Hour),
		Quantity:      decimal.NewFromInt(100),
		CostBasis:     decimal.NewFromInt(20000),
		TransactionID: transaction.ID,
	}
	require.NoError(t, db.Create(taxLot).Error)

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

	// Apply ticker change from FB to META
	err := service.ApplyTickerChange(
		portfolio.ID.String(),
		"FB",
		"META",
		user.ID.String(),
		time.Now().UTC(),
	)

	assert.NoError(t, err)

	// Verify new holding was created with same values
	newHolding, err := holdingRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "META")
	assert.NoError(t, err)
	assert.True(t, newHolding.Quantity.Equal(decimal.NewFromInt(100)))     // Same quantity
	assert.True(t, newHolding.CostBasis.Equal(decimal.NewFromInt(20000)))  // Same cost basis
	assert.True(t, newHolding.AvgCostPrice.Equal(decimal.NewFromInt(200))) // Same avg price

	// Verify tax lot was created with new symbol
	newTaxLots, err := taxLotRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "META")
	assert.NoError(t, err)
	assert.Len(t, newTaxLots, 1)
	assert.True(t, newTaxLots[0].Quantity.Equal(decimal.NewFromInt(100)))
	assert.True(t, newTaxLots[0].CostBasis.Equal(decimal.NewFromInt(20000)))
	assert.Equal(t, taxLot.PurchaseDate.Unix(), newTaxLots[0].PurchaseDate.Unix()) // Same purchase date

	// Verify transactions were updated (original BUY transaction + new TICKER_CHANGE transaction)
	updatedTransactions, err := transactionRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "META")
	assert.NoError(t, err)
	assert.Len(t, updatedTransactions, 2) // Original BUY + new TICKER_CHANGE
	// Find the original BUY transaction
	var buyTxn *models.Transaction
	for _, txn := range updatedTransactions {
		if txn.Type == models.TransactionTypeBuy {
			buyTxn = txn
			break
		}
	}
	assert.NotNil(t, buyTxn)
	assert.Equal(t, "META", buyTxn.Symbol)

	// Verify old tax lots were deleted
	oldTaxLots, err := taxLotRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "FB")
	assert.NoError(t, err)
	assert.Len(t, oldTaxLots, 0) // Old lots should be deleted

	// Verify old holding was deleted
	_, err = holdingRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "FB")
	assert.Error(t, err)
	assert.Equal(t, models.ErrHoldingNotFound, err)
}

func TestCorporateActionService_ApplyTickerChange_PortfolioNotFound(t *testing.T) {
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

	err := service.ApplyTickerChange(
		"00000000-0000-0000-0000-000000000000",
		"FB",
		"META",
		user.ID.String(),
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Equal(t, models.ErrPortfolioNotFound, err)
}

func TestCorporateActionService_ApplyTickerChange_Unauthorized(t *testing.T) {
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

	err := service.ApplyTickerChange(
		portfolio.ID.String(),
		"FB",
		"META",
		otherUser.ID.String(),
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)
}

func TestCorporateActionService_ApplyTickerChange_NoHolding(t *testing.T) {
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

	err := service.ApplyTickerChange(
		portfolio.ID.String(),
		"FB",
		"META",
		user.ID.String(),
		time.Now().UTC(),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no holding found")
}
