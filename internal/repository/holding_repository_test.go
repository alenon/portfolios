package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/lenon/portfolios/internal/models"
)

func setupHoldingRepoTestDB(t *testing.T) (*gorm.DB, *models.User, *models.Portfolio) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.User{}, &models.Portfolio{}, &models.Holding{})
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

func TestHoldingRepository_Create(t *testing.T) {
	db, _, portfolio := setupHoldingRepoTestDB(t)
	repo := NewHoldingRepository(db)

	t.Run("successful creation", func(t *testing.T) {
		holding := &models.Holding{
			PortfolioID:  portfolio.ID,
			Symbol:       "AAPL",
			Quantity:     decimal.NewFromInt(10),
			CostBasis:    decimal.NewFromFloat(1500.00),
			AvgCostPrice: decimal.NewFromFloat(150.00),
		}

		err := repo.Create(holding)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, holding.ID)
	})

	t.Run("nil holding error", func(t *testing.T) {
		err := repo.Create(nil)

		assert.Error(t, err)
	})
}

func TestHoldingRepository_FindByID(t *testing.T) {
	db, _, portfolio := setupHoldingRepoTestDB(t)
	repo := NewHoldingRepository(db)

	holding := &models.Holding{
		PortfolioID:  portfolio.ID,
		Symbol:       "AAPL",
		Quantity:     decimal.NewFromInt(10),
		CostBasis:    decimal.NewFromFloat(1500.00),
		AvgCostPrice: decimal.NewFromFloat(150.00),
	}
	err := repo.Create(holding)
	assert.NoError(t, err)

	t.Run("successful find", func(t *testing.T) {
		found, err := repo.FindByID(holding.ID.String())

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, holding.ID, found.ID)
		assert.Equal(t, holding.Symbol, found.Symbol)
	})

	t.Run("not found error", func(t *testing.T) {
		_, err := repo.FindByID(uuid.New().String())

		assert.Error(t, err)
		assert.Equal(t, models.ErrHoldingNotFound, err)
	})

	t.Run("empty id error", func(t *testing.T) {
		_, err := repo.FindByID("")

		assert.Error(t, err)
	})
}

func TestHoldingRepository_FindByPortfolioID(t *testing.T) {
	db, _, portfolio := setupHoldingRepoTestDB(t)
	repo := NewHoldingRepository(db)

	t.Run("empty list", func(t *testing.T) {
		holdings, err := repo.FindByPortfolioID(portfolio.ID.String())

		assert.NoError(t, err)
		assert.NotNil(t, holdings)
		assert.Empty(t, holdings)
	})

	// Create multiple holdings
	holdings := []*models.Holding{
		{
			PortfolioID:  portfolio.ID,
			Symbol:       "AAPL",
			Quantity:     decimal.NewFromInt(10),
			CostBasis:    decimal.NewFromFloat(1500.00),
			AvgCostPrice: decimal.NewFromFloat(150.00),
		},
		{
			PortfolioID:  portfolio.ID,
			Symbol:       "MSFT",
			Quantity:     decimal.NewFromInt(5),
			CostBasis:    decimal.NewFromFloat(1000.00),
			AvgCostPrice: decimal.NewFromFloat(200.00),
		},
	}

	for _, h := range holdings {
		err := repo.Create(h)
		assert.NoError(t, err)
	}

	t.Run("find multiple holdings", func(t *testing.T) {
		found, err := repo.FindByPortfolioID(portfolio.ID.String())

		assert.NoError(t, err)
		assert.Len(t, found, 2)
	})

	t.Run("empty portfolio id error", func(t *testing.T) {
		_, err := repo.FindByPortfolioID("")

		assert.Error(t, err)
	})
}

func TestHoldingRepository_FindByPortfolioIDAndSymbol(t *testing.T) {
	db, _, portfolio := setupHoldingRepoTestDB(t)
	repo := NewHoldingRepository(db)

	holding := &models.Holding{
		PortfolioID:  portfolio.ID,
		Symbol:       "AAPL",
		Quantity:     decimal.NewFromInt(10),
		CostBasis:    decimal.NewFromFloat(1500.00),
		AvgCostPrice: decimal.NewFromFloat(150.00),
	}
	err := repo.Create(holding)
	assert.NoError(t, err)

	t.Run("successful find", func(t *testing.T) {
		found, err := repo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, holding.ID, found.ID)
	})

	t.Run("not found error", func(t *testing.T) {
		_, err := repo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "MSFT")

		assert.Error(t, err)
		assert.Equal(t, models.ErrHoldingNotFound, err)
	})

	t.Run("empty parameters error", func(t *testing.T) {
		_, err := repo.FindByPortfolioIDAndSymbol("", "AAPL")
		assert.Error(t, err)

		_, err = repo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "")
		assert.Error(t, err)
	})
}

func TestHoldingRepository_Update(t *testing.T) {
	db, _, portfolio := setupHoldingRepoTestDB(t)
	repo := NewHoldingRepository(db)

	holding := &models.Holding{
		PortfolioID:  portfolio.ID,
		Symbol:       "AAPL",
		Quantity:     decimal.NewFromInt(10),
		CostBasis:    decimal.NewFromFloat(1500.00),
		AvgCostPrice: decimal.NewFromFloat(150.00),
	}
	err := repo.Create(holding)
	assert.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		newQuantity := decimal.NewFromInt(15)
		newCostBasis := decimal.NewFromFloat(2250.00)
		holding.Quantity = newQuantity
		holding.CostBasis = newCostBasis
		holding.CalculateAvgCostPrice()

		err := repo.Update(holding)

		assert.NoError(t, err)

		// Verify update
		found, err := repo.FindByID(holding.ID.String())
		assert.NoError(t, err)
		assert.True(t, newQuantity.Equal(found.Quantity))
		assert.True(t, newCostBasis.Equal(found.CostBasis))
	})

	t.Run("nil holding error", func(t *testing.T) {
		err := repo.Update(nil)

		assert.Error(t, err)
	})

	t.Run("not found error", func(t *testing.T) {
		nonExistent := &models.Holding{
			ID:           uuid.New(),
			PortfolioID:  portfolio.ID,
			Symbol:       "TEST",
			Quantity:     decimal.NewFromInt(1),
			CostBasis:    decimal.NewFromFloat(100.00),
			AvgCostPrice: decimal.NewFromFloat(100.00),
		}

		err := repo.Update(nonExistent)

		assert.Error(t, err)
		assert.Equal(t, models.ErrHoldingNotFound, err)
	})
}

func TestHoldingRepository_Upsert(t *testing.T) {
	db, _, portfolio := setupHoldingRepoTestDB(t)
	repo := NewHoldingRepository(db)

	t.Run("create new holding", func(t *testing.T) {
		holding := &models.Holding{
			PortfolioID:  portfolio.ID,
			Symbol:       "AAPL",
			Quantity:     decimal.NewFromInt(10),
			CostBasis:    decimal.NewFromFloat(1500.00),
			AvgCostPrice: decimal.NewFromFloat(150.00),
		}

		err := repo.Upsert(holding)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, holding.ID)

		// Verify it was created
		found, err := repo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")
		assert.NoError(t, err)
		assert.NotNil(t, found)
	})

	t.Run("update existing holding", func(t *testing.T) {
		// First, get the existing holding
		existing, err := repo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")
		assert.NoError(t, err)

		// Create a holding with same portfolio and symbol but different values
		holding := &models.Holding{
			PortfolioID:  portfolio.ID,
			Symbol:       "AAPL",
			Quantity:     decimal.NewFromInt(20),
			CostBasis:    decimal.NewFromFloat(3000.00),
			AvgCostPrice: decimal.NewFromFloat(150.00),
		}

		err = repo.Upsert(holding)

		assert.NoError(t, err)
		assert.Equal(t, existing.ID, holding.ID) // Should have same ID

		// Verify it was updated
		found, err := repo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")
		assert.NoError(t, err)
		assert.True(t, decimal.NewFromInt(20).Equal(found.Quantity))
	})
}

func TestHoldingRepository_Delete(t *testing.T) {
	db, _, portfolio := setupHoldingRepoTestDB(t)
	repo := NewHoldingRepository(db)

	holding := &models.Holding{
		PortfolioID:  portfolio.ID,
		Symbol:       "AAPL",
		Quantity:     decimal.NewFromInt(10),
		CostBasis:    decimal.NewFromFloat(1500.00),
		AvgCostPrice: decimal.NewFromFloat(150.00),
	}
	err := repo.Create(holding)
	assert.NoError(t, err)

	t.Run("successful deletion", func(t *testing.T) {
		err := repo.Delete(holding.ID.String())

		assert.NoError(t, err)

		// Verify deletion
		_, err = repo.FindByID(holding.ID.String())
		assert.Error(t, err)
	})

	t.Run("not found error", func(t *testing.T) {
		err := repo.Delete(uuid.New().String())

		assert.Error(t, err)
		assert.Equal(t, models.ErrHoldingNotFound, err)
	})

	t.Run("empty id error", func(t *testing.T) {
		err := repo.Delete("")

		assert.Error(t, err)
	})
}

func TestHoldingRepository_DeleteByPortfolioIDAndSymbol(t *testing.T) {
	db, _, portfolio := setupHoldingRepoTestDB(t)
	repo := NewHoldingRepository(db)

	holding := &models.Holding{
		PortfolioID:  portfolio.ID,
		Symbol:       "AAPL",
		Quantity:     decimal.NewFromInt(10),
		CostBasis:    decimal.NewFromFloat(1500.00),
		AvgCostPrice: decimal.NewFromFloat(150.00),
	}
	err := repo.Create(holding)
	assert.NoError(t, err)

	t.Run("successful deletion", func(t *testing.T) {
		err := repo.DeleteByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")

		assert.NoError(t, err)

		// Verify deletion
		_, err = repo.FindByPortfolioIDAndSymbol(portfolio.ID.String(), "AAPL")
		assert.Error(t, err)
	})

	t.Run("not found error", func(t *testing.T) {
		err := repo.DeleteByPortfolioIDAndSymbol(portfolio.ID.String(), "MSFT")

		assert.Error(t, err)
		assert.Equal(t, models.ErrHoldingNotFound, err)
	})

	t.Run("empty parameters error", func(t *testing.T) {
		err := repo.DeleteByPortfolioIDAndSymbol("", "AAPL")
		assert.Error(t, err)

		err = repo.DeleteByPortfolioIDAndSymbol(portfolio.ID.String(), "")
		assert.Error(t, err)
	})
}
