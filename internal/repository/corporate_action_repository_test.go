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

func setupCorporateActionTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.CorporateAction{})
	require.NoError(t, err)

	return db
}

func TestCorporateActionRepository_Create(t *testing.T) {
	db := setupCorporateActionTestDB(t)
	repo := NewCorporateActionRepository(db)

	ratio := decimal.NewFromFloat(2.0)
	description := "2-for-1 stock split"
	action := &models.CorporateAction{
		Symbol:      "AAPL",
		Type:        models.CorporateActionTypeSplit,
		Date:        time.Now().UTC(),
		Ratio:       &ratio,
		Description: description,
		Applied:     false,
	}

	err := repo.Create(action)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, action.ID)
}

func TestCorporateActionRepository_Create_Nil(t *testing.T) {
	db := setupCorporateActionTestDB(t)
	repo := NewCorporateActionRepository(db)

	err := repo.Create(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestCorporateActionRepository_FindByID(t *testing.T) {
	db := setupCorporateActionTestDB(t)
	repo := NewCorporateActionRepository(db)

	ratio := decimal.NewFromFloat(2.0)
	action := &models.CorporateAction{
		Symbol:      "AAPL",
		Type:        models.CorporateActionTypeSplit,
		Date:        time.Now().UTC(),
		Ratio:       &ratio,
		Description: "Split",
		Applied:     false,
	}
	err := repo.Create(action)
	require.NoError(t, err)

	found, err := repo.FindByID(action.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, action.ID, found.ID)
	assert.Equal(t, action.Symbol, found.Symbol)
	assert.Equal(t, action.Type, found.Type)
}

func TestCorporateActionRepository_FindByID_NotFound(t *testing.T) {
	db := setupCorporateActionTestDB(t)
	repo := NewCorporateActionRepository(db)

	_, err := repo.FindByID(uuid.New().String())
	assert.Error(t, err)
	assert.Equal(t, models.ErrCorporateActionNotFound, err)
}

func TestCorporateActionRepository_FindByID_InvalidID(t *testing.T) {
	db := setupCorporateActionTestDB(t)
	repo := NewCorporateActionRepository(db)

	_, err := repo.FindByID("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestCorporateActionRepository_FindBySymbol(t *testing.T) {
	db := setupCorporateActionTestDB(t)
	repo := NewCorporateActionRepository(db)

	// Create multiple actions for different symbols
	ratio := decimal.NewFromFloat(2.0)
	action1 := &models.CorporateAction{
		Symbol:  "AAPL",
		Type:    models.CorporateActionTypeSplit,
		Date:    time.Now().UTC().AddDate(0, 0, -2),
		Ratio:   &ratio,
		Applied: false,
	}
	action2 := &models.CorporateAction{
		Symbol:  "AAPL",
		Type:    models.CorporateActionTypeDividend,
		Date:    time.Now().UTC().AddDate(0, 0, -1),
		Amount:  &ratio,
		Applied: false,
	}
	action3 := &models.CorporateAction{
		Symbol:  "MSFT",
		Type:    models.CorporateActionTypeSplit,
		Date:    time.Now().UTC(),
		Ratio:   &ratio,
		Applied: false,
	}

	require.NoError(t, repo.Create(action1))
	require.NoError(t, repo.Create(action2))
	require.NoError(t, repo.Create(action3))

	actions, err := repo.FindBySymbol("AAPL")
	assert.NoError(t, err)
	assert.Len(t, actions, 2)
	// Should be ordered by date DESC
	assert.True(t, actions[0].Date.After(actions[1].Date) || actions[0].Date.Equal(actions[1].Date))
}

func TestCorporateActionRepository_FindBySymbolAndDateRange(t *testing.T) {
	db := setupCorporateActionTestDB(t)
	repo := NewCorporateActionRepository(db)

	now := time.Now().UTC()
	ratio := decimal.NewFromFloat(2.0)

	action1 := &models.CorporateAction{
		Symbol:  "AAPL",
		Type:    models.CorporateActionTypeSplit,
		Date:    now.AddDate(0, 0, -10),
		Ratio:   &ratio,
		Applied: false,
	}
	action2 := &models.CorporateAction{
		Symbol:  "AAPL",
		Type:    models.CorporateActionTypeSplit,
		Date:    now.AddDate(0, 0, -5),
		Ratio:   &ratio,
		Applied: false,
	}
	action3 := &models.CorporateAction{
		Symbol:  "AAPL",
		Type:    models.CorporateActionTypeSplit,
		Date:    now,
		Ratio:   &ratio,
		Applied: false,
	}

	require.NoError(t, repo.Create(action1))
	require.NoError(t, repo.Create(action2))
	require.NoError(t, repo.Create(action3))

	startDate := now.AddDate(0, 0, -7)
	endDate := now.AddDate(0, 0, 1)

	actions, err := repo.FindBySymbolAndDateRange("AAPL", startDate, endDate)
	assert.NoError(t, err)
	assert.Len(t, actions, 2) // Only action2 and action3 should be in range
}

func TestCorporateActionRepository_FindUnapplied(t *testing.T) {
	db := setupCorporateActionTestDB(t)
	repo := NewCorporateActionRepository(db)

	ratio := decimal.NewFromFloat(2.0)
	action1 := &models.CorporateAction{
		Symbol:  "AAPL",
		Type:    models.CorporateActionTypeSplit,
		Date:    time.Now().UTC(),
		Ratio:   &ratio,
		Applied: false,
	}
	action2 := &models.CorporateAction{
		Symbol:  "MSFT",
		Type:    models.CorporateActionTypeSplit,
		Date:    time.Now().UTC(),
		Ratio:   &ratio,
		Applied: true, // Already applied
	}
	action3 := &models.CorporateAction{
		Symbol:  "GOOGL",
		Type:    models.CorporateActionTypeSplit,
		Date:    time.Now().UTC(),
		Ratio:   &ratio,
		Applied: false,
	}

	require.NoError(t, repo.Create(action1))
	require.NoError(t, repo.Create(action2))
	require.NoError(t, repo.Create(action3))

	actions, err := repo.FindUnapplied()
	assert.NoError(t, err)
	assert.Len(t, actions, 2) // Only action1 and action3
	for _, action := range actions {
		assert.False(t, action.Applied)
	}
}

func TestCorporateActionRepository_Update(t *testing.T) {
	db := setupCorporateActionTestDB(t)
	repo := NewCorporateActionRepository(db)

	ratio := decimal.NewFromFloat(2.0)
	action := &models.CorporateAction{
		Symbol:  "AAPL",
		Type:    models.CorporateActionTypeSplit,
		Date:    time.Now().UTC(),
		Ratio:   &ratio,
		Applied: false,
	}
	err := repo.Create(action)
	require.NoError(t, err)

	// Update
	action.Applied = true
	newDesc := "Updated description"
	action.Description = newDesc

	err = repo.Update(action)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.FindByID(action.ID.String())
	assert.NoError(t, err)
	assert.True(t, found.Applied)
	assert.Equal(t, newDesc, found.Description)
}

func TestCorporateActionRepository_Update_Nil(t *testing.T) {
	db := setupCorporateActionTestDB(t)
	repo := NewCorporateActionRepository(db)

	err := repo.Update(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestCorporateActionRepository_Delete(t *testing.T) {
	db := setupCorporateActionTestDB(t)
	repo := NewCorporateActionRepository(db)

	ratio := decimal.NewFromFloat(2.0)
	action := &models.CorporateAction{
		Symbol:  "AAPL",
		Type:    models.CorporateActionTypeSplit,
		Date:    time.Now().UTC(),
		Ratio:   &ratio,
		Applied: false,
	}
	err := repo.Create(action)
	require.NoError(t, err)

	err = repo.Delete(action.ID.String())
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.FindByID(action.ID.String())
	assert.Error(t, err)
	assert.Equal(t, models.ErrCorporateActionNotFound, err)
}

func TestCorporateActionRepository_Delete_NotFound(t *testing.T) {
	db := setupCorporateActionTestDB(t)
	repo := NewCorporateActionRepository(db)

	err := repo.Delete(uuid.New().String())
	assert.Error(t, err)
	assert.Equal(t, models.ErrCorporateActionNotFound, err)
}

func TestCorporateActionRepository_Delete_InvalidID(t *testing.T) {
	db := setupCorporateActionTestDB(t)
	repo := NewCorporateActionRepository(db)

	err := repo.Delete("invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}
