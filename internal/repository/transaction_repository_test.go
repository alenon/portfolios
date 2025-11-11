package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/lenon/portfolios/internal/models"
)

func setupTransactionRepoTestDB(t *testing.T) (*gorm.DB, *models.User, *models.Portfolio) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.User{}, &models.Portfolio{}, &models.Transaction{})
	assert.NoError(t, err)

	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	err = user.SetPassword("password123")
	assert.NoError(t, err)
	err = db.Create(user).Error
	assert.NoError(t, err)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	err = db.Create(portfolio).Error
	assert.NoError(t, err)

	return db, user, portfolio
}

func TestTransactionRepository_Create(t *testing.T) {
	db, _, portfolio := setupTransactionRepoTestDB(t)
	repo := NewTransactionRepository(db)

	price := decimal.NewFromFloat(150.50)
	t.Run("successful creation", func(t *testing.T) {
		transaction := &models.Transaction{
			PortfolioID: portfolio.ID,
			Type:        models.TransactionTypeBuy,
			Symbol:      "AAPL",
			Date:        time.Now(),
			Quantity:    decimal.NewFromInt(10),
			Price:       &price,
			Commission:  decimal.NewFromFloat(1.00),
			Currency:    "USD",
			Notes:       "Test transaction",
		}

		err := repo.Create(transaction)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, transaction.ID)
	})

	t.Run("nil transaction error", func(t *testing.T) {
		err := repo.Create(nil)

		assert.Error(t, err)
	})
}

func TestTransactionRepository_FindByID(t *testing.T) {
	db, _, portfolio := setupTransactionRepoTestDB(t)
	repo := NewTransactionRepository(db)

	price := decimal.NewFromFloat(150.50)
	transaction := &models.Transaction{
		PortfolioID: portfolio.ID,
		Type:        models.TransactionTypeBuy,
		Symbol:      "AAPL",
		Date:        time.Now(),
		Quantity:    decimal.NewFromInt(10),
		Price:       &price,
		Commission:  decimal.Zero,
		Currency:    "USD",
	}
	err := repo.Create(transaction)
	assert.NoError(t, err)

	t.Run("successful find", func(t *testing.T) {
		found, err := repo.FindByID(transaction.ID.String())

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, transaction.ID, found.ID)
		assert.Equal(t, transaction.Symbol, found.Symbol)
	})

	t.Run("not found error", func(t *testing.T) {
		_, err := repo.FindByID(uuid.New().String())

		assert.Error(t, err)
		assert.Equal(t, models.ErrTransactionNotFound, err)
	})

	t.Run("empty id error", func(t *testing.T) {
		_, err := repo.FindByID("")

		assert.Error(t, err)
	})
}

func TestTransactionRepository_FindByPortfolioID(t *testing.T) {
	db, _, portfolio := setupTransactionRepoTestDB(t)
	repo := NewTransactionRepository(db)

	t.Run("empty list", func(t *testing.T) {
		transactions, err := repo.FindByPortfolioID(portfolio.ID.String())

		assert.NoError(t, err)
		assert.NotNil(t, transactions)
		assert.Empty(t, transactions)
	})

	// Create multiple transactions
	price1 := decimal.NewFromFloat(150.00)
	price2 := decimal.NewFromFloat(200.00)
	transactions := []*models.Transaction{
		{
			PortfolioID: portfolio.ID,
			Type:        models.TransactionTypeBuy,
			Symbol:      "AAPL",
			Date:        time.Now(),
			Quantity:    decimal.NewFromInt(10),
			Price:       &price1,
			Currency:    "USD",
		},
		{
			PortfolioID: portfolio.ID,
			Type:        models.TransactionTypeBuy,
			Symbol:      "MSFT",
			Date:        time.Now().Add(time.Hour),
			Quantity:    decimal.NewFromInt(5),
			Price:       &price2,
			Currency:    "USD",
		},
	}

	for _, tx := range transactions {
		err := repo.Create(tx)
		assert.NoError(t, err)
	}

	t.Run("find multiple transactions", func(t *testing.T) {
		found, err := repo.FindByPortfolioID(portfolio.ID.String())

		assert.NoError(t, err)
		assert.Len(t, found, 2)
	})

	t.Run("empty portfolio id error", func(t *testing.T) {
		_, err := repo.FindByPortfolioID("")

		assert.Error(t, err)
	})
}

func TestTransactionRepository_FindByPortfolioIDAndSymbol(t *testing.T) {
	db, _, portfolio := setupTransactionRepoTestDB(t)
	repo := NewTransactionRepository(db)

	price1 := decimal.NewFromFloat(150.00)
	price2 := decimal.NewFromFloat(160.00)
	price3 := decimal.NewFromFloat(200.00)

	// Create transactions for different symbols
	transactions := []*models.Transaction{
		{
			PortfolioID: portfolio.ID,
			Type:        models.TransactionTypeBuy,
			Symbol:      "AAPL",
			Date:        time.Now(),
			Quantity:    decimal.NewFromInt(10),
			Price:       &price1,
			Currency:    "USD",
		},
		{
			PortfolioID: portfolio.ID,
			Type:        models.TransactionTypeBuy,
			Symbol:      "AAPL",
			Date:        time.Now().Add(time.Hour),
			Quantity:    decimal.NewFromInt(5),
			Price:       &price2,
			Currency:    "USD",
		},
		{
			PortfolioID: portfolio.ID,
			Type:        models.TransactionTypeBuy,
			Symbol:      "MSFT",
			Date:        time.Now(),
			Quantity:    decimal.NewFromInt(3),
			Price:       &price3,
			Currency:    "USD",
		},
	}

	for _, tx := range transactions {
		err := repo.Create(tx)
		assert.NoError(t, err)
	}

	t.Run("filter by symbol", func(t *testing.T) {
		found, err := repo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")

		assert.NoError(t, err)
		assert.Len(t, found, 2)
		for _, tx := range found {
			assert.Equal(t, "AAPL", tx.Symbol)
		}
	})

	t.Run("empty parameters error", func(t *testing.T) {
		_, err := repo.FindByPortfolioIDAndSymbol("", "AAPL")
		assert.Error(t, err)

		_, err = repo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "")
		assert.Error(t, err)
	})
}

func TestTransactionRepository_Update(t *testing.T) {
	db, _, portfolio := setupTransactionRepoTestDB(t)
	repo := NewTransactionRepository(db)

	price := decimal.NewFromFloat(150.00)
	transaction := &models.Transaction{
		PortfolioID: portfolio.ID,
		Type:        models.TransactionTypeBuy,
		Symbol:      "AAPL",
		Date:        time.Now(),
		Quantity:    decimal.NewFromInt(10),
		Price:       &price,
		Currency:    "USD",
		Notes:       "Original notes",
	}
	err := repo.Create(transaction)
	assert.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		transaction.Notes = "Updated notes"
		newQuantity := decimal.NewFromInt(15)
		transaction.Quantity = newQuantity

		err := repo.Update(transaction)

		assert.NoError(t, err)

		// Verify update
		found, err := repo.FindByID(transaction.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, "Updated notes", found.Notes)
		assert.True(t, newQuantity.Equal(found.Quantity))
	})

	t.Run("nil transaction error", func(t *testing.T) {
		err := repo.Update(nil)

		assert.Error(t, err)
	})

	t.Run("not found error", func(t *testing.T) {
		nonExistent := &models.Transaction{
			ID:          uuid.New(),
			PortfolioID: portfolio.ID,
			Type:        models.TransactionTypeBuy,
			Symbol:      "TEST",
			Date:        time.Now(),
			Quantity:    decimal.NewFromInt(1),
			Currency:    "USD",
		}

		err := repo.Update(nonExistent)

		assert.Error(t, err)
		assert.Equal(t, models.ErrTransactionNotFound, err)
	})
}

func TestTransactionRepository_Delete(t *testing.T) {
	db, _, portfolio := setupTransactionRepoTestDB(t)
	repo := NewTransactionRepository(db)

	price := decimal.NewFromFloat(150.00)
	transaction := &models.Transaction{
		PortfolioID: portfolio.ID,
		Type:        models.TransactionTypeBuy,
		Symbol:      "AAPL",
		Date:        time.Now(),
		Quantity:    decimal.NewFromInt(10),
		Price:       &price,
		Currency:    "USD",
	}
	err := repo.Create(transaction)
	assert.NoError(t, err)

	t.Run("successful deletion", func(t *testing.T) {
		err := repo.Delete(transaction.ID.String())

		assert.NoError(t, err)

		// Verify deletion
		_, err = repo.FindByID(transaction.ID.String())
		assert.Error(t, err)
	})

	t.Run("not found error", func(t *testing.T) {
		err := repo.Delete(uuid.New().String())

		assert.Error(t, err)
		assert.Equal(t, models.ErrTransactionNotFound, err)
	})

	t.Run("empty id error", func(t *testing.T) {
		err := repo.Delete("")

		assert.Error(t, err)
	})
}
