package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
)

func setupTransactionTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate schemas
	err = db.AutoMigrate(&models.User{}, &models.Portfolio{}, &models.Transaction{}, &models.Holding{})
	assert.NoError(t, err)

	return db
}

func createTestUserAndPortfolio(t *testing.T, db *gorm.DB) (*models.User, *models.Portfolio) {
	userRepo := repository.NewUserRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)

	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	err := user.SetPassword("password123")
	assert.NoError(t, err)
	err = userRepo.Create(user)
	assert.NoError(t, err)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Test Portfolio",
		Description:     "Test",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	err = portfolioRepo.Create(portfolio)
	assert.NoError(t, err)

	return user, portfolio
}

func TestTransactionService_CreateBuy(t *testing.T) {
	db := setupTransactionTestDB(t)
	transactionRepo := repository.NewTransactionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	service := NewTransactionService(transactionRepo, portfolioRepo, holdingRepo)

	user, portfolio := createTestUserAndPortfolio(t, db)

	t.Run("successful buy transaction", func(t *testing.T) {
		transaction, err := service.Create(
			portfolio.ID.String(),
			user.ID.String(),
			models.TransactionTypeBuy,
			"AAPL",
			time.Now(),
			decimal.NewFromInt(10),
			decimal.NewFromFloat(150.50),
			decimal.NewFromFloat(1.00),
			"USD",
			"Initial purchase",
		)

		assert.NoError(t, err)
		assert.NotNil(t, transaction)
		assert.Equal(t, models.TransactionTypeBuy, transaction.Type)
		assert.Equal(t, "AAPL", transaction.Symbol)
		assert.Equal(t, decimal.NewFromInt(10), transaction.Quantity)

		// Verify holding was created
		holding, err := holdingRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")
		assert.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(10), holding.Quantity)
		// Total cost = 10 shares * $150.50 + $1.00 commission = $1506.00
		expectedCost := decimal.NewFromFloat(1506.00)
		assert.True(t, holding.CostBasis.Equal(expectedCost))
	})

	t.Run("second buy transaction updates holding", func(t *testing.T) {
		transaction, err := service.Create(
			portfolio.ID.String(),
			user.ID.String(),
			models.TransactionTypeBuy,
			"AAPL",
			time.Now(),
			decimal.NewFromInt(5),
			decimal.NewFromFloat(160.00),
			decimal.NewFromFloat(1.00),
			"USD",
			"Additional purchase",
		)

		assert.NoError(t, err)
		assert.NotNil(t, transaction)

		// Verify holding was updated
		holding, err := holdingRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")
		assert.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(15), holding.Quantity) // 10 + 5

		// Total cost = previous $1506.00 + (5 * $160.00 + $1.00) = $2307.00
		expectedCost := decimal.NewFromFloat(2307.00)
		assert.True(t, holding.CostBasis.Equal(expectedCost))
	})

	t.Run("unauthorized access", func(t *testing.T) {
		otherUserID := uuid.New().String()
		_, err := service.Create(
			portfolio.ID.String(),
			otherUserID,
			models.TransactionTypeBuy,
			"MSFT",
			time.Now(),
			decimal.NewFromInt(10),
			decimal.NewFromFloat(200.00),
			decimal.Zero,
			"USD",
			"",
		)

		assert.Error(t, err)
		assert.Equal(t, models.ErrUnauthorizedAccess, err)
	})

	t.Run("invalid quantity", func(t *testing.T) {
		_, err := service.Create(
			portfolio.ID.String(),
			user.ID.String(),
			models.TransactionTypeBuy,
			"MSFT",
			time.Now(),
			decimal.Zero, // Invalid quantity
			decimal.NewFromFloat(200.00),
			decimal.Zero,
			"USD",
			"",
		)

		assert.Error(t, err)
		assert.Equal(t, models.ErrInvalidQuantity, err)
	})
}

func TestTransactionService_CreateSell(t *testing.T) {
	db := setupTransactionTestDB(t)
	transactionRepo := repository.NewTransactionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	service := NewTransactionService(transactionRepo, portfolioRepo, holdingRepo)

	user, portfolio := createTestUserAndPortfolio(t, db)

	// First, buy some shares
	_, err := service.Create(
		portfolio.ID.String(),
		user.ID.String(),
		models.TransactionTypeBuy,
		"AAPL",
		time.Now(),
		decimal.NewFromInt(10),
		decimal.NewFromFloat(150.00),
		decimal.Zero,
		"USD",
		"Initial purchase",
	)
	assert.NoError(t, err)

	t.Run("successful sell transaction", func(t *testing.T) {
		transaction, err := service.Create(
			portfolio.ID.String(),
			user.ID.String(),
			models.TransactionTypeSell,
			"AAPL",
			time.Now(),
			decimal.NewFromInt(5),
			decimal.NewFromFloat(160.00),
			decimal.NewFromFloat(1.00),
			"USD",
			"Partial sale",
		)

		assert.NoError(t, err)
		assert.NotNil(t, transaction)
		assert.Equal(t, models.TransactionTypeSell, transaction.Type)

		// Verify holding was updated
		holding, err := holdingRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")
		assert.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(5), holding.Quantity) // 10 - 5
	})

	t.Run("insufficient shares error", func(t *testing.T) {
		_, err := service.Create(
			portfolio.ID.String(),
			user.ID.String(),
			models.TransactionTypeSell,
			"AAPL",
			time.Now(),
			decimal.NewFromInt(100), // More than available
			decimal.NewFromFloat(160.00),
			decimal.Zero,
			"USD",
			"",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient shares")
	})

	t.Run("sell without holding", func(t *testing.T) {
		_, err := service.Create(
			portfolio.ID.String(),
			user.ID.String(),
			models.TransactionTypeSell,
			"MSFT", // Never bought
			time.Now(),
			decimal.NewFromInt(10),
			decimal.NewFromFloat(200.00),
			decimal.Zero,
			"USD",
			"",
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient shares")
	})
}

func TestTransactionService_GetByID(t *testing.T) {
	db := setupTransactionTestDB(t)
	transactionRepo := repository.NewTransactionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	service := NewTransactionService(transactionRepo, portfolioRepo, holdingRepo)

	user, portfolio := createTestUserAndPortfolio(t, db)

	// Create a transaction
	transaction, err := service.Create(
		portfolio.ID.String(),
		user.ID.String(),
		models.TransactionTypeBuy,
		"AAPL",
		time.Now(),
		decimal.NewFromInt(10),
		decimal.NewFromFloat(150.00),
		decimal.Zero,
		"USD",
		"Test",
	)
	assert.NoError(t, err)

	t.Run("successful retrieval", func(t *testing.T) {
		retrieved, err := service.GetByID(transaction.ID.String(), user.ID.String())

		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, transaction.ID, retrieved.ID)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		otherUserID := uuid.New().String()
		_, err := service.GetByID(transaction.ID.String(), otherUserID)

		assert.Error(t, err)
		assert.Equal(t, models.ErrUnauthorizedAccess, err)
	})
}

func TestTransactionService_GetByPortfolioID(t *testing.T) {
	db := setupTransactionTestDB(t)
	transactionRepo := repository.NewTransactionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	service := NewTransactionService(transactionRepo, portfolioRepo, holdingRepo)

	user, portfolio := createTestUserAndPortfolio(t, db)

	// Create multiple transactions
	_, err := service.Create(
		portfolio.ID.String(),
		user.ID.String(),
		models.TransactionTypeBuy,
		"AAPL",
		time.Now(),
		decimal.NewFromInt(10),
		decimal.NewFromFloat(150.00),
		decimal.Zero,
		"USD",
		"",
	)
	assert.NoError(t, err)

	_, err = service.Create(
		portfolio.ID.String(),
		user.ID.String(),
		models.TransactionTypeBuy,
		"MSFT",
		time.Now(),
		decimal.NewFromInt(5),
		decimal.NewFromFloat(200.00),
		decimal.Zero,
		"USD",
		"",
	)
	assert.NoError(t, err)

	t.Run("retrieve all transactions", func(t *testing.T) {
		transactions, err := service.GetByPortfolioID(portfolio.ID.String(), user.ID.String())

		assert.NoError(t, err)
		assert.Len(t, transactions, 2)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		otherUserID := uuid.New().String()
		_, err := service.GetByPortfolioID(portfolio.ID.String(), otherUserID)

		assert.Error(t, err)
		assert.Equal(t, models.ErrUnauthorizedAccess, err)
	})
}

func TestTransactionService_GetByPortfolioIDAndSymbol(t *testing.T) {
	db := setupTransactionTestDB(t)
	transactionRepo := repository.NewTransactionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	service := NewTransactionService(transactionRepo, portfolioRepo, holdingRepo)

	user, portfolio := createTestUserAndPortfolio(t, db)

	// Create transactions for different symbols
	_, err := service.Create(
		portfolio.ID.String(),
		user.ID.String(),
		models.TransactionTypeBuy,
		"AAPL",
		time.Now(),
		decimal.NewFromInt(10),
		decimal.NewFromFloat(150.00),
		decimal.Zero,
		"USD",
		"",
	)
	assert.NoError(t, err)

	_, err = service.Create(
		portfolio.ID.String(),
		user.ID.String(),
		models.TransactionTypeBuy,
		"AAPL",
		time.Now(),
		decimal.NewFromInt(5),
		decimal.NewFromFloat(160.00),
		decimal.Zero,
		"USD",
		"",
	)
	assert.NoError(t, err)

	_, err = service.Create(
		portfolio.ID.String(),
		user.ID.String(),
		models.TransactionTypeBuy,
		"MSFT",
		time.Now(),
		decimal.NewFromInt(3),
		decimal.NewFromFloat(200.00),
		decimal.Zero,
		"USD",
		"",
	)
	assert.NoError(t, err)

	t.Run("filter by symbol", func(t *testing.T) {
		transactions, err := service.GetByPortfolioIDAndSymbol(
			portfolio.ID.String(),
			"AAPL",
			user.ID.String(),
		)

		assert.NoError(t, err)
		assert.Len(t, transactions, 2)
		for _, tx := range transactions {
			assert.Equal(t, "AAPL", tx.Symbol)
		}
	})
}

func TestTransactionService_Delete(t *testing.T) {
	db := setupTransactionTestDB(t)
	transactionRepo := repository.NewTransactionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	service := NewTransactionService(transactionRepo, portfolioRepo, holdingRepo)

	user, portfolio := createTestUserAndPortfolio(t, db)

	// Create a transaction
	transaction, err := service.Create(
		portfolio.ID.String(),
		user.ID.String(),
		models.TransactionTypeBuy,
		"AAPL",
		time.Now(),
		decimal.NewFromInt(10),
		decimal.NewFromFloat(150.00),
		decimal.Zero,
		"USD",
		"",
	)
	assert.NoError(t, err)

	t.Run("successful deletion", func(t *testing.T) {
		err := service.Delete(transaction.ID.String(), user.ID.String())

		assert.NoError(t, err)

		// Verify it's deleted
		_, err = service.GetByID(transaction.ID.String(), user.ID.String())
		assert.Error(t, err)
	})

	t.Run("unauthorized deletion", func(t *testing.T) {
		// Create another transaction
		transaction2, err := service.Create(
			portfolio.ID.String(),
			user.ID.String(),
			models.TransactionTypeBuy,
			"MSFT",
			time.Now(),
			decimal.NewFromInt(5),
			decimal.NewFromFloat(200.00),
			decimal.Zero,
			"USD",
			"",
		)
		assert.NoError(t, err)

		otherUserID := uuid.New().String()
		err = service.Delete(transaction2.ID.String(), otherUserID)

		assert.Error(t, err)
		assert.Equal(t, models.ErrUnauthorizedAccess, err)
	})
}

func TestTransactionService_Update(t *testing.T) {
	db := setupTransactionTestDB(t)
	transactionRepo := repository.NewTransactionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	service := NewTransactionService(transactionRepo, portfolioRepo, holdingRepo)

	user, portfolio := createTestUserAndPortfolio(t, db)

	// Create initial transaction
	transaction, err := service.Create(
		portfolio.ID.String(),
		user.ID.String(),
		models.TransactionTypeBuy,
		"AAPL",
		time.Now(),
		decimal.NewFromInt(10),
		decimal.NewFromFloat(150.00),
		decimal.Zero,
		"USD",
		"Initial purchase",
	)
	assert.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		updated, err := service.Update(
			transaction.ID.String(),
			user.ID.String(),
			models.TransactionTypeBuy,
			"AAPL",
			time.Now(),
			decimal.NewFromInt(15),       // Changed quantity
			decimal.NewFromFloat(155.00), // Changed price
			decimal.Zero,
			"USD",
			"Updated purchase",
		)

		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, decimal.NewFromInt(15), updated.Quantity)
		assert.Equal(t, decimal.NewFromFloat(155.00), *updated.Price)
		assert.Equal(t, "Updated purchase", updated.Notes)

		// Verify holdings were recalculated
		holding, err := holdingRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")
		assert.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(15), holding.Quantity)
	})

	t.Run("unauthorized update", func(t *testing.T) {
		otherUserID := uuid.New().String()
		_, err := service.Update(
			transaction.ID.String(),
			otherUserID,
			models.TransactionTypeBuy,
			"AAPL",
			time.Now(),
			decimal.NewFromInt(20),
			decimal.NewFromFloat(160.00),
			decimal.Zero,
			"USD",
			"Unauthorized update",
		)

		assert.Error(t, err)
		assert.Equal(t, models.ErrUnauthorizedAccess, err)
	})

	t.Run("validation error", func(t *testing.T) {
		// Try to update with zero quantity (should fail validation)
		_, err := service.Update(
			transaction.ID.String(),
			user.ID.String(),
			models.TransactionTypeBuy,
			"AAPL",
			time.Now(),
			decimal.Zero, // Invalid: zero quantity
			decimal.NewFromFloat(150.00),
			decimal.Zero,
			"USD",
			"Invalid update",
		)

		assert.Error(t, err)
	})
}

func TestTransactionService_RecalculateHoldingsForSymbol(t *testing.T) {
	db := setupTransactionTestDB(t)
	transactionRepo := repository.NewTransactionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	holdingRepo := repository.NewHoldingRepository(db)
	service := NewTransactionService(transactionRepo, portfolioRepo, holdingRepo)

	user, portfolio := createTestUserAndPortfolio(t, db)

	t.Run("recalculate after multiple buys", func(t *testing.T) {
		// Create first buy
		_, err := service.Create(
			portfolio.ID.String(),
			user.ID.String(),
			models.TransactionTypeBuy,
			"TSLA",
			time.Now().AddDate(0, 0, -2),
			decimal.NewFromInt(10),
			decimal.NewFromFloat(100.00),
			decimal.Zero,
			"USD",
			"",
		)
		assert.NoError(t, err)

		// Create second buy
		_, err = service.Create(
			portfolio.ID.String(),
			user.ID.String(),
			models.TransactionTypeBuy,
			"TSLA",
			time.Now().AddDate(0, 0, -1),
			decimal.NewFromInt(5),
			decimal.NewFromFloat(110.00),
			decimal.Zero,
			"USD",
			"",
		)
		assert.NoError(t, err)

		// Verify holdings
		holding, err := holdingRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "TSLA")
		assert.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(15), holding.Quantity)
		// Cost basis = (10 * 100) + (5 * 110) = 1000 + 550 = 1550
		assert.Equal(t, decimal.NewFromInt(1550), holding.CostBasis)
	})

	t.Run("recalculate after buy and sell", func(t *testing.T) {
		// Create buy
		_, err := service.Create(
			portfolio.ID.String(),
			user.ID.String(),
			models.TransactionTypeBuy,
			"MSFT",
			time.Now().AddDate(0, 0, -2),
			decimal.NewFromInt(20),
			decimal.NewFromFloat(200.00),
			decimal.Zero,
			"USD",
			"",
		)
		assert.NoError(t, err)

		// Create sell
		_, err = service.Create(
			portfolio.ID.String(),
			user.ID.String(),
			models.TransactionTypeSell,
			"MSFT",
			time.Now().AddDate(0, 0, -1),
			decimal.NewFromInt(5),
			decimal.NewFromFloat(210.00),
			decimal.Zero,
			"USD",
			"",
		)
		assert.NoError(t, err)

		// Verify holdings
		holding, err := holdingRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "MSFT")
		assert.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(15), holding.Quantity)
		// Cost basis = (20 * 200) - (5 * 200) = 4000 - 1000 = 3000
		assert.Equal(t, decimal.NewFromInt(3000), holding.CostBasis)
	})

	t.Run("delete holding when quantity reaches zero", func(t *testing.T) {
		// Create buy
		_, err := service.Create(
			portfolio.ID.String(),
			user.ID.String(),
			models.TransactionTypeBuy,
			"GOOG",
			time.Now().AddDate(0, 0, -2),
			decimal.NewFromInt(10),
			decimal.NewFromFloat(1000.00),
			decimal.Zero,
			"USD",
			"",
		)
		assert.NoError(t, err)

		// Verify holding exists
		holding, err := holdingRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "GOOG")
		assert.NoError(t, err)
		assert.NotNil(t, holding)

		// Sell all shares
		_, err = service.Create(
			portfolio.ID.String(),
			user.ID.String(),
			models.TransactionTypeSell,
			"GOOG",
			time.Now().AddDate(0, 0, -1),
			decimal.NewFromInt(10),
			decimal.NewFromFloat(1100.00),
			decimal.Zero,
			"USD",
			"",
		)
		assert.NoError(t, err)

		// Verify holding is deleted
		_, err = holdingRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "GOOG")
		assert.Error(t, err)
		assert.Equal(t, models.ErrHoldingNotFound, err)
	})

	t.Run("create holding if not exists", func(t *testing.T) {
		// Manually delete a holding to test recreation
		holding, err := holdingRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "TSLA")
		assert.NoError(t, err)
		err = holdingRepo.Delete(holding.ID.String())
		assert.NoError(t, err)

		// Trigger recalculation via service (accessing private method through Update)
		// Create a transaction to trigger recalculation
		tx, err := service.Create(
			portfolio.ID.String(),
			user.ID.String(),
			models.TransactionTypeBuy,
			"TSLA",
			time.Now(),
			decimal.NewFromInt(1),
			decimal.NewFromFloat(120.00),
			decimal.Zero,
			"USD",
			"",
		)
		assert.NoError(t, err)

		// Holding should be recreated through recalculation
		recreatedHolding, err := holdingRepo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "TSLA")
		assert.NoError(t, err)
		assert.NotNil(t, recreatedHolding)

		// Clean up the test transaction
		_ = tx
	})
}
