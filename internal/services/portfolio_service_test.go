package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupPortfolioTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate schemas
	err = db.AutoMigrate(&models.User{}, &models.Portfolio{})
	assert.NoError(t, err)

	return db
}

func TestPortfolioService_Create(t *testing.T) {
	db := setupPortfolioTestDB(t)
	portfolioRepo := repository.NewPortfolioRepository(db)
	userRepo := repository.NewUserRepository(db)
	service := NewPortfolioService(portfolioRepo, userRepo)

	// Create a test user
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	err := user.SetPassword("password123")
	assert.NoError(t, err)
	err = userRepo.Create(user)
	assert.NoError(t, err)

	t.Run("successful creation", func(t *testing.T) {
		portfolio, err := service.Create(
			user.ID.String(),
			"Test Portfolio",
			"A test portfolio",
			"USD",
			models.CostBasisFIFO,
		)

		assert.NoError(t, err)
		assert.NotNil(t, portfolio)
		assert.Equal(t, "Test Portfolio", portfolio.Name)
		assert.Equal(t, "A test portfolio", portfolio.Description)
		assert.Equal(t, "USD", portfolio.BaseCurrency)
		assert.Equal(t, models.CostBasisFIFO, portfolio.CostBasisMethod)
		assert.Equal(t, user.ID, portfolio.UserID)
	})

	t.Run("duplicate name error", func(t *testing.T) {
		_, err := service.Create(
			user.ID.String(),
			"Test Portfolio",
			"Another portfolio",
			"USD",
			models.CostBasisFIFO,
		)

		assert.Error(t, err)
		assert.Equal(t, models.ErrPortfolioDuplicateName, err)
	})

	t.Run("empty name error", func(t *testing.T) {
		_, err := service.Create(
			user.ID.String(),
			"",
			"Description",
			"USD",
			models.CostBasisFIFO,
		)

		assert.Error(t, err)
		assert.Equal(t, models.ErrPortfolioNameRequired, err)
	})

	t.Run("invalid user error", func(t *testing.T) {
		_, err := service.Create(
			uuid.New().String(),
			"New Portfolio",
			"Description",
			"USD",
			models.CostBasisFIFO,
		)

		assert.Error(t, err)
	})

	t.Run("default values", func(t *testing.T) {
		portfolio, err := service.Create(
			user.ID.String(),
			"Default Portfolio",
			"",
			"",
			"",
		)

		assert.NoError(t, err)
		assert.Equal(t, "USD", portfolio.BaseCurrency)
		assert.Equal(t, models.CostBasisFIFO, portfolio.CostBasisMethod)
	})
}

func TestPortfolioService_GetByID(t *testing.T) {
	db := setupPortfolioTestDB(t)
	portfolioRepo := repository.NewPortfolioRepository(db)
	userRepo := repository.NewUserRepository(db)
	service := NewPortfolioService(portfolioRepo, userRepo)

	// Create test user
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	err := user.SetPassword("password123")
	assert.NoError(t, err)
	err = userRepo.Create(user)
	assert.NoError(t, err)

	// Create test portfolio
	portfolio, err := service.Create(
		user.ID.String(),
		"Test Portfolio",
		"Description",
		"USD",
		models.CostBasisFIFO,
	)
	assert.NoError(t, err)

	t.Run("successful retrieval", func(t *testing.T) {
		retrieved, err := service.GetByID(portfolio.ID.String(), user.ID.String())

		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, portfolio.ID, retrieved.ID)
		assert.Equal(t, portfolio.Name, retrieved.Name)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		otherUserID := uuid.New().String()
		_, err := service.GetByID(portfolio.ID.String(), otherUserID)

		assert.Error(t, err)
		assert.Equal(t, models.ErrUnauthorizedAccess, err)
	})

	t.Run("not found error", func(t *testing.T) {
		_, err := service.GetByID(uuid.New().String(), user.ID.String())

		assert.Error(t, err)
	})
}

func TestPortfolioService_GetAllByUserID(t *testing.T) {
	db := setupPortfolioTestDB(t)
	portfolioRepo := repository.NewPortfolioRepository(db)
	userRepo := repository.NewUserRepository(db)
	service := NewPortfolioService(portfolioRepo, userRepo)

	// Create test user
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	err := user.SetPassword("password123")
	assert.NoError(t, err)
	err = userRepo.Create(user)
	assert.NoError(t, err)

	t.Run("empty list", func(t *testing.T) {
		portfolios, err := service.GetAllByUserID(user.ID.String())

		assert.NoError(t, err)
		assert.NotNil(t, portfolios)
		assert.Empty(t, portfolios)
	})

	// Create test portfolios
	_, err = service.Create(user.ID.String(), "Portfolio 1", "Desc 1", "USD", models.CostBasisFIFO)
	assert.NoError(t, err)
	_, err = service.Create(user.ID.String(), "Portfolio 2", "Desc 2", "USD", models.CostBasisLIFO)
	assert.NoError(t, err)

	t.Run("multiple portfolios", func(t *testing.T) {
		portfolios, err := service.GetAllByUserID(user.ID.String())

		assert.NoError(t, err)
		assert.Len(t, portfolios, 2)
	})
}

func TestPortfolioService_Update(t *testing.T) {
	db := setupPortfolioTestDB(t)
	portfolioRepo := repository.NewPortfolioRepository(db)
	userRepo := repository.NewUserRepository(db)
	service := NewPortfolioService(portfolioRepo, userRepo)

	// Create test user
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	err := user.SetPassword("password123")
	assert.NoError(t, err)
	err = userRepo.Create(user)
	assert.NoError(t, err)

	// Create test portfolio
	portfolio, err := service.Create(
		user.ID.String(),
		"Original Name",
		"Original Description",
		"USD",
		models.CostBasisFIFO,
	)
	assert.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		updated, err := service.Update(
			portfolio.ID.String(),
			user.ID.String(),
			"Updated Name",
			"Updated Description",
		)

		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "Updated Description", updated.Description)
	})

	t.Run("unauthorized update", func(t *testing.T) {
		otherUserID := uuid.New().String()
		_, err := service.Update(
			portfolio.ID.String(),
			otherUserID,
			"Hacked Name",
			"Hacked Description",
		)

		assert.Error(t, err)
		assert.Equal(t, models.ErrUnauthorizedAccess, err)
	})

	t.Run("duplicate name error", func(t *testing.T) {
		// Create another portfolio
		_, err := service.Create(
			user.ID.String(),
			"Another Portfolio",
			"Description",
			"USD",
			models.CostBasisFIFO,
		)
		assert.NoError(t, err)

		// Try to update first portfolio with the second portfolio's name
		_, err = service.Update(
			portfolio.ID.String(),
			user.ID.String(),
			"Another Portfolio",
			"Description",
		)

		assert.Error(t, err)
		assert.Equal(t, models.ErrPortfolioDuplicateName, err)
	})
}

func TestPortfolioService_Delete(t *testing.T) {
	db := setupPortfolioTestDB(t)
	portfolioRepo := repository.NewPortfolioRepository(db)
	userRepo := repository.NewUserRepository(db)
	service := NewPortfolioService(portfolioRepo, userRepo)

	// Create test user
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	err := user.SetPassword("password123")
	assert.NoError(t, err)
	err = userRepo.Create(user)
	assert.NoError(t, err)

	// Create test portfolio
	portfolio, err := service.Create(
		user.ID.String(),
		"To Delete",
		"Description",
		"USD",
		models.CostBasisFIFO,
	)
	assert.NoError(t, err)

	t.Run("successful deletion", func(t *testing.T) {
		err := service.Delete(portfolio.ID.String(), user.ID.String())

		assert.NoError(t, err)

		// Verify it's deleted
		_, err = service.GetByID(portfolio.ID.String(), user.ID.String())
		assert.Error(t, err)
	})

	t.Run("unauthorized deletion", func(t *testing.T) {
		// Create another portfolio
		portfolio2, err := service.Create(
			user.ID.String(),
			"Another Portfolio",
			"Description",
			"USD",
			models.CostBasisFIFO,
		)
		assert.NoError(t, err)

		otherUserID := uuid.New().String()
		err = service.Delete(portfolio2.ID.String(), otherUserID)

		assert.Error(t, err)
		assert.Equal(t, models.ErrUnauthorizedAccess, err)
	})
}
