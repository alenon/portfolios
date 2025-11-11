package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPerformanceSnapshotTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.User{}, &models.Portfolio{}, &models.PerformanceSnapshot{})
	assert.NoError(t, err)

	return db
}

func TestPerformanceSnapshotRepository_Create(t *testing.T) {
	db := setupPerformanceSnapshotTestDB(t)
	repo := NewPerformanceSnapshotRepository(db)

	// Create a user and portfolio
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
	}
	assert.NoError(t, db.Create(user).Error)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	assert.NoError(t, db.Create(portfolio).Error)

	snapshot := &models.PerformanceSnapshot{
		PortfolioID:    portfolio.ID,
		Date:           time.Now().UTC(),
		TotalValue:     decimal.NewFromInt(10000),
		TotalCostBasis: decimal.NewFromInt(8000),
		TotalReturn:    decimal.NewFromInt(2000),
		TotalReturnPct: decimal.NewFromFloat(25.0),
	}

	err := repo.Create(snapshot)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, snapshot.ID)
}

func TestPerformanceSnapshotRepository_Create_Nil(t *testing.T) {
	db := setupPerformanceSnapshotTestDB(t)
	repo := NewPerformanceSnapshotRepository(db)

	err := repo.Create(nil)
	assert.Error(t, err)
}

func TestPerformanceSnapshotRepository_FindByID(t *testing.T) {
	db := setupPerformanceSnapshotTestDB(t)
	repo := NewPerformanceSnapshotRepository(db)

	// Create a user and portfolio
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
	}
	assert.NoError(t, db.Create(user).Error)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	assert.NoError(t, db.Create(portfolio).Error)

	snapshot := &models.PerformanceSnapshot{
		PortfolioID:    portfolio.ID,
		Date:           time.Now().UTC(),
		TotalValue:     decimal.NewFromInt(10000),
		TotalCostBasis: decimal.NewFromInt(8000),
		TotalReturn:    decimal.NewFromInt(2000),
		TotalReturnPct: decimal.NewFromFloat(25.0),
	}
	assert.NoError(t, repo.Create(snapshot))

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByID(snapshot.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, snapshot.ID, found.ID)
		assert.True(t, snapshot.TotalValue.Equal(found.TotalValue))
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.FindByID(uuid.New().String())
		assert.Equal(t, models.ErrPerformanceSnapshotNotFound, err)
	})

	t.Run("empty ID", func(t *testing.T) {
		_, err := repo.FindByID("")
		assert.Error(t, err)
	})
}

func TestPerformanceSnapshotRepository_FindByPortfolioID(t *testing.T) {
	db := setupPerformanceSnapshotTestDB(t)
	repo := NewPerformanceSnapshotRepository(db)

	// Create a user and portfolio
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
	}
	assert.NoError(t, db.Create(user).Error)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	assert.NoError(t, db.Create(portfolio).Error)

	// Create snapshots
	now := time.Now().UTC()
	for i := 0; i < 5; i++ {
		snapshot := &models.PerformanceSnapshot{
			PortfolioID:    portfolio.ID,
			Date:           now.Add(time.Duration(-i) * 24 * time.Hour),
			TotalValue:     decimal.NewFromInt(10000 + int64(i*100)),
			TotalCostBasis: decimal.NewFromInt(8000),
			TotalReturn:    decimal.NewFromInt(2000 + int64(i*100)),
			TotalReturnPct: decimal.NewFromFloat(25.0),
		}
		assert.NoError(t, repo.Create(snapshot))
	}

	t.Run("find all", func(t *testing.T) {
		snapshots, err := repo.FindByPortfolioID(portfolio.ID.String(), 0, 0)
		assert.NoError(t, err)
		assert.Len(t, snapshots, 5)
		// Should be ordered by date DESC
		assert.True(t, snapshots[0].Date.After(snapshots[1].Date))
	})

	t.Run("with limit", func(t *testing.T) {
		snapshots, err := repo.FindByPortfolioID(portfolio.ID.String(), 2, 0)
		assert.NoError(t, err)
		assert.Len(t, snapshots, 2)
	})

	t.Run("with offset", func(t *testing.T) {
		snapshots, err := repo.FindByPortfolioID(portfolio.ID.String(), 2, 2)
		assert.NoError(t, err)
		assert.Len(t, snapshots, 2)
	})

	t.Run("empty portfolio ID", func(t *testing.T) {
		_, err := repo.FindByPortfolioID("", 0, 0)
		assert.Error(t, err)
	})
}

func TestPerformanceSnapshotRepository_FindByPortfolioIDAndDateRange(t *testing.T) {
	db := setupPerformanceSnapshotTestDB(t)
	repo := NewPerformanceSnapshotRepository(db)

	// Create a user and portfolio
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
	}
	assert.NoError(t, db.Create(user).Error)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	assert.NoError(t, db.Create(portfolio).Error)

	// Create snapshots over 10 days
	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		snapshot := &models.PerformanceSnapshot{
			PortfolioID:    portfolio.ID,
			Date:           now.Add(time.Duration(-i) * 24 * time.Hour),
			TotalValue:     decimal.NewFromInt(10000 + int64(i*100)),
			TotalCostBasis: decimal.NewFromInt(8000),
			TotalReturn:    decimal.NewFromInt(2000 + int64(i*100)),
			TotalReturnPct: decimal.NewFromFloat(25.0),
		}
		assert.NoError(t, repo.Create(snapshot))
	}

	startDate := now.Add(-5 * 24 * time.Hour)
	endDate := now

	snapshots, err := repo.FindByPortfolioIDAndDateRange(portfolio.ID.String(), startDate, endDate)
	assert.NoError(t, err)
	assert.LessOrEqual(t, len(snapshots), 6) // Should have at most 6 snapshots (days 0-5)
	// Should be ordered by date ASC
	if len(snapshots) > 1 {
		assert.True(t, snapshots[0].Date.Before(snapshots[1].Date) || snapshots[0].Date.Equal(snapshots[1].Date))
	}
}

func TestPerformanceSnapshotRepository_FindLatestByPortfolioID(t *testing.T) {
	db := setupPerformanceSnapshotTestDB(t)
	repo := NewPerformanceSnapshotRepository(db)

	// Create a user and portfolio
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
	}
	assert.NoError(t, db.Create(user).Error)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	assert.NoError(t, db.Create(portfolio).Error)

	// Create snapshots
	now := time.Now().UTC()
	var latestSnapshot *models.PerformanceSnapshot
	for i := 0; i < 5; i++ {
		snapshot := &models.PerformanceSnapshot{
			PortfolioID:    portfolio.ID,
			Date:           now.Add(time.Duration(-i) * 24 * time.Hour),
			TotalValue:     decimal.NewFromInt(10000 + int64(i*100)),
			TotalCostBasis: decimal.NewFromInt(8000),
			TotalReturn:    decimal.NewFromInt(2000 + int64(i*100)),
			TotalReturnPct: decimal.NewFromFloat(25.0),
		}
		assert.NoError(t, repo.Create(snapshot))
		if i == 0 {
			latestSnapshot = snapshot
		}
	}

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindLatestByPortfolioID(portfolio.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, latestSnapshot.ID, found.ID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.FindLatestByPortfolioID(uuid.New().String())
		assert.Equal(t, models.ErrPerformanceSnapshotNotFound, err)
	})

	t.Run("empty portfolio ID", func(t *testing.T) {
		_, err := repo.FindLatestByPortfolioID("")
		assert.Error(t, err)
	})
}

func TestPerformanceSnapshotRepository_Delete(t *testing.T) {
	db := setupPerformanceSnapshotTestDB(t)
	repo := NewPerformanceSnapshotRepository(db)

	// Create a user and portfolio
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
	}
	assert.NoError(t, db.Create(user).Error)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	assert.NoError(t, db.Create(portfolio).Error)

	snapshot := &models.PerformanceSnapshot{
		PortfolioID:    portfolio.ID,
		Date:           time.Now().UTC(),
		TotalValue:     decimal.NewFromInt(10000),
		TotalCostBasis: decimal.NewFromInt(8000),
		TotalReturn:    decimal.NewFromInt(2000),
		TotalReturnPct: decimal.NewFromFloat(25.0),
	}
	assert.NoError(t, repo.Create(snapshot))

	t.Run("successful delete", func(t *testing.T) {
		err := repo.Delete(snapshot.ID.String())
		assert.NoError(t, err)

		// Verify it's deleted
		_, err = repo.FindByID(snapshot.ID.String())
		assert.Equal(t, models.ErrPerformanceSnapshotNotFound, err)
	})

	t.Run("not found", func(t *testing.T) {
		err := repo.Delete(uuid.New().String())
		assert.Equal(t, models.ErrPerformanceSnapshotNotFound, err)
	})

	t.Run("empty ID", func(t *testing.T) {
		err := repo.Delete("")
		assert.Error(t, err)
	})
}

func TestPerformanceSnapshotRepository_DeleteByPortfolioID(t *testing.T) {
	db := setupPerformanceSnapshotTestDB(t)
	repo := NewPerformanceSnapshotRepository(db)

	// Create a user and portfolio
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
	}
	assert.NoError(t, db.Create(user).Error)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	assert.NoError(t, db.Create(portfolio).Error)

	// Create multiple snapshots
	for i := 0; i < 5; i++ {
		snapshot := &models.PerformanceSnapshot{
			PortfolioID:    portfolio.ID,
			Date:           time.Now().UTC().Add(time.Duration(-i) * 24 * time.Hour),
			TotalValue:     decimal.NewFromInt(10000 + int64(i*100)),
			TotalCostBasis: decimal.NewFromInt(8000),
			TotalReturn:    decimal.NewFromInt(2000 + int64(i*100)),
			TotalReturnPct: decimal.NewFromFloat(25.0),
		}
		assert.NoError(t, repo.Create(snapshot))
	}

	err := repo.DeleteByPortfolioID(portfolio.ID.String())
	assert.NoError(t, err)

	// Verify all deleted
	snapshots, err := repo.FindByPortfolioID(portfolio.ID.String(), 0, 0)
	assert.NoError(t, err)
	assert.Len(t, snapshots, 0)
}
