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

func setupPortfolioActionTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.User{},
		&models.Portfolio{},
		&models.CorporateAction{},
		&models.PortfolioAction{},
	)
	require.NoError(t, err)

	return db
}

func createTestDataForPortfolioAction(t *testing.T, db *gorm.DB) (*models.Portfolio, *models.CorporateAction) {
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}
	require.NoError(t, db.Create(user).Error)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	require.NoError(t, db.Create(portfolio).Error)

	ratio := decimal.NewFromFloat(2.0)
	action := &models.CorporateAction{
		Symbol:  "AAPL",
		Type:    models.CorporateActionTypeSplit,
		Date:    time.Now().UTC(),
		Ratio:   &ratio,
		Applied: false,
	}
	require.NoError(t, db.Create(action).Error)

	return portfolio, action
}

func TestPortfolioActionRepository_Create(t *testing.T) {
	db := setupPortfolioActionTestDB(t)
	repo := NewPortfolioActionRepository(db)

	portfolio, corpAction := createTestDataForPortfolioAction(t, db)

	portfolioAction := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		Status:            models.PortfolioActionStatusPending,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC(),
	}

	err := repo.Create(portfolioAction)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, portfolioAction.ID)
}

func TestPortfolioActionRepository_Create_Nil(t *testing.T) {
	db := setupPortfolioActionTestDB(t)
	repo := NewPortfolioActionRepository(db)

	err := repo.Create(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestPortfolioActionRepository_FindByID(t *testing.T) {
	db := setupPortfolioActionTestDB(t)
	repo := NewPortfolioActionRepository(db)

	portfolio, corpAction := createTestDataForPortfolioAction(t, db)

	portfolioAction := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		Status:            models.PortfolioActionStatusPending,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC(),
	}
	require.NoError(t, repo.Create(portfolioAction))

	found, err := repo.FindByID(portfolioAction.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, portfolioAction.ID, found.ID)
	assert.Equal(t, portfolioAction.Status, found.Status)
	assert.NotNil(t, found.Portfolio)
	assert.NotNil(t, found.CorporateAction)
}

func TestPortfolioActionRepository_FindByID_NotFound(t *testing.T) {
	db := setupPortfolioActionTestDB(t)
	repo := NewPortfolioActionRepository(db)

	_, err := repo.FindByID(uuid.New().String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPortfolioActionRepository_FindByPortfolioID(t *testing.T) {
	db := setupPortfolioActionTestDB(t)
	repo := NewPortfolioActionRepository(db)

	portfolio, corpAction := createTestDataForPortfolioAction(t, db)

	// Create multiple actions
	action1 := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		Status:            models.PortfolioActionStatusPending,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC().AddDate(0, 0, -2),
	}
	action2 := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		Status:            models.PortfolioActionStatusApproved,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC().AddDate(0, 0, -1),
	}

	require.NoError(t, repo.Create(action1))
	require.NoError(t, repo.Create(action2))

	actions, err := repo.FindByPortfolioID(portfolio.ID.String())
	assert.NoError(t, err)
	assert.Len(t, actions, 2)
	// Should be ordered by detected_at DESC
	assert.True(t, actions[0].DetectedAt.After(actions[1].DetectedAt) || actions[0].DetectedAt.Equal(actions[1].DetectedAt))
}

func TestPortfolioActionRepository_FindPendingByPortfolioID(t *testing.T) {
	db := setupPortfolioActionTestDB(t)
	repo := NewPortfolioActionRepository(db)

	portfolio, corpAction := createTestDataForPortfolioAction(t, db)

	// Create actions with different statuses
	pending := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		Status:            models.PortfolioActionStatusPending,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC(),
	}
	approved := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		Status:            models.PortfolioActionStatusApproved,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC(),
	}

	require.NoError(t, repo.Create(pending))
	require.NoError(t, repo.Create(approved))

	actions, err := repo.FindPendingByPortfolioID(portfolio.ID.String())
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
	assert.Equal(t, models.PortfolioActionStatusPending, actions[0].Status)
}

func TestPortfolioActionRepository_FindByPortfolioIDAndStatus(t *testing.T) {
	db := setupPortfolioActionTestDB(t)
	repo := NewPortfolioActionRepository(db)

	portfolio, corpAction := createTestDataForPortfolioAction(t, db)

	pending := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		Status:            models.PortfolioActionStatusPending,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC(),
	}
	approved := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		Status:            models.PortfolioActionStatusApproved,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC(),
	}

	require.NoError(t, repo.Create(pending))
	require.NoError(t, repo.Create(approved))

	actions, err := repo.FindByPortfolioIDAndStatus(portfolio.ID.String(), models.PortfolioActionStatusApproved)
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
	assert.Equal(t, models.PortfolioActionStatusApproved, actions[0].Status)
}

func TestPortfolioActionRepository_FindPendingByCorporateActionID(t *testing.T) {
	db := setupPortfolioActionTestDB(t)
	repo := NewPortfolioActionRepository(db)

	portfolio, corpAction := createTestDataForPortfolioAction(t, db)

	pending := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		Status:            models.PortfolioActionStatusPending,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC(),
	}
	approved := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		Status:            models.PortfolioActionStatusApproved,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC(),
	}

	require.NoError(t, repo.Create(pending))
	require.NoError(t, repo.Create(approved))

	actions, err := repo.FindPendingByCorporateActionID(corpAction.ID.String())
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
	assert.Equal(t, models.PortfolioActionStatusPending, actions[0].Status)
	assert.NotNil(t, actions[0].Portfolio)
}

func TestPortfolioActionRepository_Update(t *testing.T) {
	db := setupPortfolioActionTestDB(t)
	repo := NewPortfolioActionRepository(db)

	portfolio, corpAction := createTestDataForPortfolioAction(t, db)

	portfolioAction := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		Status:            models.PortfolioActionStatusPending,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC(),
	}
	require.NoError(t, repo.Create(portfolioAction))

	// Update status
	portfolioAction.Status = models.PortfolioActionStatusApproved
	now := time.Now().UTC()
	portfolioAction.ReviewedAt = &now

	err := repo.Update(portfolioAction)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.FindByID(portfolioAction.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, models.PortfolioActionStatusApproved, found.Status)
	assert.NotNil(t, found.ReviewedAt)
}

func TestPortfolioActionRepository_Delete(t *testing.T) {
	db := setupPortfolioActionTestDB(t)
	repo := NewPortfolioActionRepository(db)

	portfolio, corpAction := createTestDataForPortfolioAction(t, db)

	portfolioAction := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		Status:            models.PortfolioActionStatusPending,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC(),
	}
	require.NoError(t, repo.Create(portfolioAction))

	err := repo.Delete(portfolioAction.ID.String())
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.FindByID(portfolioAction.ID.String())
	assert.Error(t, err)
}

func TestPortfolioActionRepository_Delete_NotFound(t *testing.T) {
	db := setupPortfolioActionTestDB(t)
	repo := NewPortfolioActionRepository(db)

	err := repo.Delete(uuid.New().String())
	assert.Error(t, err)
}

func TestPortfolioActionRepository_ExistsPendingForPortfolioAndAction(t *testing.T) {
	db := setupPortfolioActionTestDB(t)
	repo := NewPortfolioActionRepository(db)

	portfolio, corpAction := createTestDataForPortfolioAction(t, db)

	// No action exists yet
	exists, err := repo.ExistsPendingForPortfolioAndAction(portfolio.ID.String(), corpAction.ID.String())
	assert.NoError(t, err)
	assert.False(t, exists)

	// Create pending action
	portfolioAction := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		Status:            models.PortfolioActionStatusPending,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC(),
	}
	require.NoError(t, repo.Create(portfolioAction))

	// Now it should exist
	exists, err = repo.ExistsPendingForPortfolioAndAction(portfolio.ID.String(), corpAction.ID.String())
	assert.NoError(t, err)
	assert.True(t, exists)

	// Change status to approved
	portfolioAction.Status = models.PortfolioActionStatusApproved
	require.NoError(t, repo.Update(portfolioAction))

	// Should not exist as pending anymore
	exists, err = repo.ExistsPendingForPortfolioAndAction(portfolio.ID.String(), corpAction.ID.String())
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestPortfolioActionRepository_ExistsPendingForPortfolioAndAction_InvalidID(t *testing.T) {
	db := setupPortfolioActionTestDB(t)
	repo := NewPortfolioActionRepository(db)

	_, err := repo.ExistsPendingForPortfolioAndAction("invalid", uuid.New().String())
	assert.Error(t, err)

	_, err = repo.ExistsPendingForPortfolioAndAction(uuid.New().String(), "invalid")
	assert.Error(t, err)
}
