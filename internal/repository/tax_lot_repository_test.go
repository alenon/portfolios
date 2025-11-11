package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTaxLotTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.User{}, &models.Portfolio{}, &models.Transaction{}, &models.TaxLot{})
	require.NoError(t, err)

	return db
}

func createTestUserForTaxLot(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func createTestPortfolioForTaxLot(t *testing.T, db *gorm.DB, userID uuid.UUID) *models.Portfolio {
	portfolio := &models.Portfolio{
		UserID:          userID,
		Name:            "Test Portfolio",
		Description:     "Test Description",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	err := db.Create(portfolio).Error
	require.NoError(t, err)
	return portfolio
}

func createTestTransactionForTaxLot(t *testing.T, db *gorm.DB, portfolioID uuid.UUID) *models.Transaction {
	price := decimal.NewFromFloat(100.50)
	transaction := &models.Transaction{
		PortfolioID: portfolioID,
		Type:        models.TransactionTypeBuy,
		Symbol:      "AAPL",
		Date:        time.Now().UTC(),
		Quantity:    decimal.NewFromInt(10),
		Price:       &price,
		Commission:  decimal.NewFromFloat(1.00),
		Currency:    "USD",
	}
	err := db.Create(transaction).Error
	require.NoError(t, err)
	return transaction
}

func TestTaxLotRepository_Create(t *testing.T) {
	db := setupTaxLotTestDB(t)
	repo := NewTaxLotRepository(db)

	user := createTestUserForTaxLot(t, db)
	portfolio := createTestPortfolioForTaxLot(t, db, user.ID)
	transaction := createTestTransactionForTaxLot(t, db, portfolio.ID)

	taxLot := &models.TaxLot{
		PortfolioID:   portfolio.ID,
		Symbol:        "AAPL",
		PurchaseDate:  time.Now().UTC(),
		Quantity:      decimal.NewFromInt(10),
		CostBasis:     decimal.NewFromFloat(1005.00),
		TransactionID: transaction.ID,
	}

	err := repo.Create(taxLot)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, taxLot.ID)
}

func TestTaxLotRepository_Create_NilTaxLot(t *testing.T) {
	db := setupTaxLotTestDB(t)
	repo := NewTaxLotRepository(db)

	err := repo.Create(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestTaxLotRepository_FindByID(t *testing.T) {
	db := setupTaxLotTestDB(t)
	repo := NewTaxLotRepository(db)

	user := createTestUserForTaxLot(t, db)
	portfolio := createTestPortfolioForTaxLot(t, db, user.ID)
	transaction := createTestTransactionForTaxLot(t, db, portfolio.ID)

	taxLot := &models.TaxLot{
		PortfolioID:   portfolio.ID,
		Symbol:        "AAPL",
		PurchaseDate:  time.Now().UTC(),
		Quantity:      decimal.NewFromInt(10),
		CostBasis:     decimal.NewFromFloat(1005.00),
		TransactionID: transaction.ID,
	}
	err := repo.Create(taxLot)
	require.NoError(t, err)

	found, err := repo.FindByID(taxLot.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, taxLot.ID, found.ID)
	assert.Equal(t, taxLot.Symbol, found.Symbol)
	assert.Equal(t, taxLot.Quantity.String(), found.Quantity.String())
}

func TestTaxLotRepository_FindByID_NotFound(t *testing.T) {
	db := setupTaxLotTestDB(t)
	repo := NewTaxLotRepository(db)

	_, err := repo.FindByID(uuid.New().String())
	assert.Error(t, err)
	assert.Equal(t, models.ErrTaxLotNotFound, err)
}

func TestTaxLotRepository_FindByID_InvalidID(t *testing.T) {
	db := setupTaxLotTestDB(t)
	repo := NewTaxLotRepository(db)

	_, err := repo.FindByID("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestTaxLotRepository_FindByPortfolioID(t *testing.T) {
	db := setupTaxLotTestDB(t)
	repo := NewTaxLotRepository(db)

	user := createTestUserForTaxLot(t, db)
	portfolio := createTestPortfolioForTaxLot(t, db, user.ID)
	transaction := createTestTransactionForTaxLot(t, db, portfolio.ID)

	// Create multiple tax lots
	taxLot1 := &models.TaxLot{
		PortfolioID:   portfolio.ID,
		Symbol:        "AAPL",
		PurchaseDate:  time.Now().UTC().AddDate(0, 0, -2),
		Quantity:      decimal.NewFromInt(10),
		CostBasis:     decimal.NewFromFloat(1005.00),
		TransactionID: transaction.ID,
	}
	err := repo.Create(taxLot1)
	require.NoError(t, err)

	taxLot2 := &models.TaxLot{
		PortfolioID:   portfolio.ID,
		Symbol:        "MSFT",
		PurchaseDate:  time.Now().UTC().AddDate(0, 0, -1),
		Quantity:      decimal.NewFromInt(5),
		CostBasis:     decimal.NewFromFloat(500.00),
		TransactionID: transaction.ID,
	}
	err = repo.Create(taxLot2)
	require.NoError(t, err)

	taxLots, err := repo.FindByPortfolioID(portfolio.ID.String())
	assert.NoError(t, err)
	assert.Len(t, taxLots, 2)
	// Verify they are ordered by purchase date
	assert.Equal(t, "AAPL", taxLots[0].Symbol)
	assert.Equal(t, "MSFT", taxLots[1].Symbol)
}

func TestTaxLotRepository_FindByPortfolioIDAndSymbol(t *testing.T) {
	db := setupTaxLotTestDB(t)
	repo := NewTaxLotRepository(db)

	user := createTestUserForTaxLot(t, db)
	portfolio := createTestPortfolioForTaxLot(t, db, user.ID)
	transaction := createTestTransactionForTaxLot(t, db, portfolio.ID)

	// Create tax lots for different symbols
	taxLot1 := &models.TaxLot{
		PortfolioID:   portfolio.ID,
		Symbol:        "AAPL",
		PurchaseDate:  time.Now().UTC().AddDate(0, 0, -2),
		Quantity:      decimal.NewFromInt(10),
		CostBasis:     decimal.NewFromFloat(1005.00),
		TransactionID: transaction.ID,
	}
	err := repo.Create(taxLot1)
	require.NoError(t, err)

	taxLot2 := &models.TaxLot{
		PortfolioID:   portfolio.ID,
		Symbol:        "AAPL",
		PurchaseDate:  time.Now().UTC().AddDate(0, 0, -1),
		Quantity:      decimal.NewFromInt(5),
		CostBasis:     decimal.NewFromFloat(502.50),
		TransactionID: transaction.ID,
	}
	err = repo.Create(taxLot2)
	require.NoError(t, err)

	taxLot3 := &models.TaxLot{
		PortfolioID:   portfolio.ID,
		Symbol:        "MSFT",
		PurchaseDate:  time.Now().UTC(),
		Quantity:      decimal.NewFromInt(3),
		CostBasis:     decimal.NewFromFloat(300.00),
		TransactionID: transaction.ID,
	}
	err = repo.Create(taxLot3)
	require.NoError(t, err)

	taxLots, err := repo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")
	assert.NoError(t, err)
	assert.Len(t, taxLots, 2)
	// Verify they are FIFO ordered (oldest first)
	assert.True(t, taxLots[0].PurchaseDate.Before(taxLots[1].PurchaseDate))
}

func TestTaxLotRepository_FindByTransactionID(t *testing.T) {
	db := setupTaxLotTestDB(t)
	repo := NewTaxLotRepository(db)

	user := createTestUserForTaxLot(t, db)
	portfolio := createTestPortfolioForTaxLot(t, db, user.ID)
	transaction := createTestTransactionForTaxLot(t, db, portfolio.ID)

	taxLot := &models.TaxLot{
		PortfolioID:   portfolio.ID,
		Symbol:        "AAPL",
		PurchaseDate:  time.Now().UTC(),
		Quantity:      decimal.NewFromInt(10),
		CostBasis:     decimal.NewFromFloat(1005.00),
		TransactionID: transaction.ID,
	}
	err := repo.Create(taxLot)
	require.NoError(t, err)

	taxLots, err := repo.FindByTransactionID(transaction.ID.String())
	assert.NoError(t, err)
	assert.Len(t, taxLots, 1)
	assert.Equal(t, taxLot.ID, taxLots[0].ID)
}

func TestTaxLotRepository_Update(t *testing.T) {
	db := setupTaxLotTestDB(t)
	repo := NewTaxLotRepository(db)

	user := createTestUserForTaxLot(t, db)
	portfolio := createTestPortfolioForTaxLot(t, db, user.ID)
	transaction := createTestTransactionForTaxLot(t, db, portfolio.ID)

	taxLot := &models.TaxLot{
		PortfolioID:   portfolio.ID,
		Symbol:        "AAPL",
		PurchaseDate:  time.Now().UTC(),
		Quantity:      decimal.NewFromInt(10),
		CostBasis:     decimal.NewFromFloat(1005.00),
		TransactionID: transaction.ID,
	}
	err := repo.Create(taxLot)
	require.NoError(t, err)

	// Update quantity
	taxLot.Quantity = decimal.NewFromInt(5)
	taxLot.CostBasis = decimal.NewFromFloat(502.50)

	err = repo.Update(taxLot)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.FindByID(taxLot.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, "5", found.Quantity.String())
	assert.Equal(t, "502.5", found.CostBasis.String())
}

func TestTaxLotRepository_Delete(t *testing.T) {
	db := setupTaxLotTestDB(t)
	repo := NewTaxLotRepository(db)

	user := createTestUserForTaxLot(t, db)
	portfolio := createTestPortfolioForTaxLot(t, db, user.ID)
	transaction := createTestTransactionForTaxLot(t, db, portfolio.ID)

	taxLot := &models.TaxLot{
		PortfolioID:   portfolio.ID,
		Symbol:        "AAPL",
		PurchaseDate:  time.Now().UTC(),
		Quantity:      decimal.NewFromInt(10),
		CostBasis:     decimal.NewFromFloat(1005.00),
		TransactionID: transaction.ID,
	}
	err := repo.Create(taxLot)
	require.NoError(t, err)

	err = repo.Delete(taxLot.ID.String())
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.FindByID(taxLot.ID.String())
	assert.Error(t, err)
	assert.Equal(t, models.ErrTaxLotNotFound, err)
}

func TestTaxLotRepository_Delete_NotFound(t *testing.T) {
	db := setupTaxLotTestDB(t)
	repo := NewTaxLotRepository(db)

	err := repo.Delete(uuid.New().String())
	assert.Error(t, err)
	assert.Equal(t, models.ErrTaxLotNotFound, err)
}

func TestTaxLotRepository_DeleteByPortfolioIDAndSymbol(t *testing.T) {
	db := setupTaxLotTestDB(t)
	repo := NewTaxLotRepository(db)

	user := createTestUserForTaxLot(t, db)
	portfolio := createTestPortfolioForTaxLot(t, db, user.ID)
	transaction := createTestTransactionForTaxLot(t, db, portfolio.ID)

	// Create multiple tax lots
	for i := 0; i < 3; i++ {
		taxLot := &models.TaxLot{
			PortfolioID:   portfolio.ID,
			Symbol:        "AAPL",
			PurchaseDate:  time.Now().UTC(),
			Quantity:      decimal.NewFromInt(10),
			CostBasis:     decimal.NewFromFloat(1005.00),
			TransactionID: transaction.ID,
		}
		err := repo.Create(taxLot)
		require.NoError(t, err)
	}

	// Create one for a different symbol
	taxLot := &models.TaxLot{
		PortfolioID:   portfolio.ID,
		Symbol:        "MSFT",
		PurchaseDate:  time.Now().UTC(),
		Quantity:      decimal.NewFromInt(5),
		CostBasis:     decimal.NewFromFloat(500.00),
		TransactionID: transaction.ID,
	}
	err := repo.Create(taxLot)
	require.NoError(t, err)

	// Delete AAPL tax lots
	err = repo.DeleteByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")
	assert.NoError(t, err)

	// Verify AAPL lots are deleted
	aaplLots, err := repo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")
	assert.NoError(t, err)
	assert.Len(t, aaplLots, 0)

	// Verify MSFT lot still exists
	msftLots, err := repo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "MSFT")
	assert.NoError(t, err)
	assert.Len(t, msftLots, 1)
}
