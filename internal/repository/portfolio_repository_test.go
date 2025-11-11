package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/lenon/portfolios/internal/models"
)

func setupPortfolioRepoTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.User{}, &models.Portfolio{})
	assert.NoError(t, err)

	return db
}

func createTestUser(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	err := user.SetPassword("password123")
	assert.NoError(t, err)
	err = db.Create(user).Error
	assert.NoError(t, err)
	return user
}

func TestPortfolioRepository_Create(t *testing.T) {
	db := setupPortfolioRepoTestDB(t)
	repo := NewPortfolioRepository(db)
	user := createTestUser(t, db)

	t.Run("successful creation", func(t *testing.T) {
		portfolio := &models.Portfolio{
			UserID:          user.ID,
			Name:            "Test Portfolio",
			Description:     "Test description",
			BaseCurrency:    "USD",
			CostBasisMethod: models.CostBasisFIFO,
		}

		err := repo.Create(portfolio)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, portfolio.ID)
	})

	t.Run("nil portfolio error", func(t *testing.T) {
		err := repo.Create(nil)

		assert.Error(t, err)
	})
}

func TestPortfolioRepository_FindByID(t *testing.T) {
	db := setupPortfolioRepoTestDB(t)
	repo := NewPortfolioRepository(db)
	user := createTestUser(t, db)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	err := repo.Create(portfolio)
	assert.NoError(t, err)

	t.Run("successful find", func(t *testing.T) {
		found, err := repo.FindByID(portfolio.ID.String())

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, portfolio.ID, found.ID)
		assert.Equal(t, portfolio.Name, found.Name)
	})

	t.Run("not found error", func(t *testing.T) {
		_, err := repo.FindByID(uuid.New().String())

		assert.Error(t, err)
		assert.Equal(t, models.ErrPortfolioNotFound, err)
	})

	t.Run("empty id error", func(t *testing.T) {
		_, err := repo.FindByID("")

		assert.Error(t, err)
	})

	t.Run("invalid uuid error", func(t *testing.T) {
		_, err := repo.FindByID("invalid-uuid")

		assert.Error(t, err)
	})
}

func TestPortfolioRepository_FindByUserID(t *testing.T) {
	db := setupPortfolioRepoTestDB(t)
	repo := NewPortfolioRepository(db)
	user := createTestUser(t, db)

	t.Run("empty list", func(t *testing.T) {
		portfolios, err := repo.FindByUserID(user.ID.String())

		assert.NoError(t, err)
		assert.NotNil(t, portfolios)
		assert.Empty(t, portfolios)
	})

	// Create multiple portfolios
	for i := 1; i <= 3; i++ {
		portfolio := &models.Portfolio{
			UserID:          user.ID,
			Name:            "Portfolio " + string(rune('0'+i)),
			BaseCurrency:    "USD",
			CostBasisMethod: models.CostBasisFIFO,
		}
		err := repo.Create(portfolio)
		assert.NoError(t, err)
	}

	t.Run("find multiple portfolios", func(t *testing.T) {
		portfolios, err := repo.FindByUserID(user.ID.String())

		assert.NoError(t, err)
		assert.Len(t, portfolios, 3)
	})

	t.Run("empty user id error", func(t *testing.T) {
		_, err := repo.FindByUserID("")

		assert.Error(t, err)
	})
}

func TestPortfolioRepository_FindByUserIDAndName(t *testing.T) {
	db := setupPortfolioRepoTestDB(t)
	repo := NewPortfolioRepository(db)
	user := createTestUser(t, db)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Unique Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	err := repo.Create(portfolio)
	assert.NoError(t, err)

	t.Run("successful find", func(t *testing.T) {
		found, err := repo.FindByUserIDAndName(user.ID.String(), "Unique Portfolio")

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, portfolio.ID, found.ID)
	})

	t.Run("not found error", func(t *testing.T) {
		_, err := repo.FindByUserIDAndName(user.ID.String(), "Non-existent")

		assert.Error(t, err)
		assert.Equal(t, models.ErrPortfolioNotFound, err)
	})

	t.Run("empty parameters error", func(t *testing.T) {
		_, err := repo.FindByUserIDAndName("", "Portfolio")
		assert.Error(t, err)

		_, err = repo.FindByUserIDAndName(user.ID.String(), "")
		assert.Error(t, err)
	})
}

func TestPortfolioRepository_Update(t *testing.T) {
	db := setupPortfolioRepoTestDB(t)
	repo := NewPortfolioRepository(db)
	user := createTestUser(t, db)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Original Name",
		Description:     "Original Description",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	err := repo.Create(portfolio)
	assert.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		portfolio.Name = "Updated Name"
		portfolio.Description = "Updated Description"

		err := repo.Update(portfolio)

		assert.NoError(t, err)

		// Verify update
		found, err := repo.FindByID(portfolio.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, "Updated Description", found.Description)
	})

	t.Run("nil portfolio error", func(t *testing.T) {
		err := repo.Update(nil)

		assert.Error(t, err)
	})

	t.Run("not found error", func(t *testing.T) {
		nonExistent := &models.Portfolio{
			ID:              uuid.New(),
			UserID:          user.ID,
			Name:            "Test",
			BaseCurrency:    "USD",
			CostBasisMethod: models.CostBasisFIFO,
		}

		err := repo.Update(nonExistent)

		assert.Error(t, err)
		assert.Equal(t, models.ErrPortfolioNotFound, err)
	})
}

func TestPortfolioRepository_Delete(t *testing.T) {
	db := setupPortfolioRepoTestDB(t)
	repo := NewPortfolioRepository(db)
	user := createTestUser(t, db)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "To Delete",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	err := repo.Create(portfolio)
	assert.NoError(t, err)

	t.Run("successful deletion", func(t *testing.T) {
		err := repo.Delete(portfolio.ID.String())

		assert.NoError(t, err)

		// Verify deletion
		_, err = repo.FindByID(portfolio.ID.String())
		assert.Error(t, err)
	})

	t.Run("not found error", func(t *testing.T) {
		err := repo.Delete(uuid.New().String())

		assert.Error(t, err)
		assert.Equal(t, models.ErrPortfolioNotFound, err)
	})

	t.Run("empty id error", func(t *testing.T) {
		err := repo.Delete("")

		assert.Error(t, err)
	})
}

func TestPortfolioRepository_ExistsByUserIDAndName(t *testing.T) {
	db := setupPortfolioRepoTestDB(t)
	repo := NewPortfolioRepository(db)
	user := createTestUser(t, db)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Existing Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	err := repo.Create(portfolio)
	assert.NoError(t, err)

	t.Run("exists returns true", func(t *testing.T) {
		exists, err := repo.ExistsByUserIDAndName(user.ID.String(), "Existing Portfolio")

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("does not exist returns false", func(t *testing.T) {
		exists, err := repo.ExistsByUserIDAndName(user.ID.String(), "Non-existent")

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("empty parameters error", func(t *testing.T) {
		_, err := repo.ExistsByUserIDAndName("", "Portfolio")
		assert.Error(t, err)

		_, err = repo.ExistsByUserIDAndName(user.ID.String(), "")
		assert.Error(t, err)
	})
}
