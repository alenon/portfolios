package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPortfolioAction_TableName(t *testing.T) {
	action := &PortfolioAction{}
	assert.Equal(t, "portfolio_actions", action.TableName())
}

func TestPortfolioAction_BeforeCreate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}, &Portfolio{}, &CorporateAction{}, &PortfolioAction{}))

	// Create test user and portfolio
	user := &User{Email: "test@example.com", PasswordHash: "hash"}
	require.NoError(t, db.Create(user).Error)

	portfolio := &Portfolio{
		UserID:          user.ID,
		Name:            "Test",
		BaseCurrency:    "USD",
		CostBasisMethod: CostBasisFIFO,
	}
	require.NoError(t, db.Create(portfolio).Error)

	corpAction := &CorporateAction{
		Symbol: "AAPL",
		Type:   CorporateActionTypeSplit,
		Date:   time.Now().UTC(),
	}
	require.NoError(t, db.Create(corpAction).Error)

	portfolioAction := &PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
	}

	err = db.Create(portfolioAction).Error
	require.NoError(t, err)

	// Check defaults
	assert.NotEqual(t, uuid.Nil, portfolioAction.ID)
	assert.Equal(t, PortfolioActionStatusPending, portfolioAction.Status)
	assert.False(t, portfolioAction.DetectedAt.IsZero())
	assert.False(t, portfolioAction.CreatedAt.IsZero())
	assert.False(t, portfolioAction.UpdatedAt.IsZero())
}

func TestPortfolioAction_BeforeUpdate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}, &Portfolio{}, &CorporateAction{}, &PortfolioAction{}))

	user := &User{Email: "test@example.com", PasswordHash: "hash"}
	require.NoError(t, db.Create(user).Error)

	portfolio := &Portfolio{
		UserID:          user.ID,
		Name:            "Test",
		BaseCurrency:    "USD",
		CostBasisMethod: CostBasisFIFO,
	}
	require.NoError(t, db.Create(portfolio).Error)

	corpAction := &CorporateAction{
		Symbol: "AAPL",
		Type:   CorporateActionTypeSplit,
		Date:   time.Now().UTC(),
	}
	require.NoError(t, db.Create(corpAction).Error)

	portfolioAction := &PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
	}
	require.NoError(t, db.Create(portfolioAction).Error)

	oldUpdatedAt := portfolioAction.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	// Update
	portfolioAction.Notes = "Updated"
	require.NoError(t, db.Save(portfolioAction).Error)

	assert.True(t, portfolioAction.UpdatedAt.After(oldUpdatedAt))
}

func TestPortfolioAction_IsPending(t *testing.T) {
	action := &PortfolioAction{Status: PortfolioActionStatusPending}
	assert.True(t, action.IsPending())

	action.Status = PortfolioActionStatusApproved
	assert.False(t, action.IsPending())
}

func TestPortfolioAction_IsApplied(t *testing.T) {
	action := &PortfolioAction{Status: PortfolioActionStatusApplied}
	assert.True(t, action.IsApplied())

	action.Status = PortfolioActionStatusPending
	assert.False(t, action.IsApplied())
}

func TestPortfolioAction_Approve(t *testing.T) {
	userID := uuid.New()
	action := &PortfolioAction{Status: PortfolioActionStatusPending}

	action.Approve(userID)

	assert.Equal(t, PortfolioActionStatusApproved, action.Status)
	assert.NotNil(t, action.ReviewedAt)
	assert.NotNil(t, action.ReviewedByUserID)
	assert.Equal(t, userID, *action.ReviewedByUserID)
}

func TestPortfolioAction_Reject(t *testing.T) {
	userID := uuid.New()
	action := &PortfolioAction{Status: PortfolioActionStatusPending}
	reason := "Not needed"

	action.Reject(userID, reason)

	assert.Equal(t, PortfolioActionStatusRejected, action.Status)
	assert.NotNil(t, action.ReviewedAt)
	assert.NotNil(t, action.ReviewedByUserID)
	assert.Equal(t, userID, *action.ReviewedByUserID)
	assert.Equal(t, reason, action.Notes)
}

func TestPortfolioAction_Reject_EmptyReason(t *testing.T) {
	userID := uuid.New()
	action := &PortfolioAction{
		Status: PortfolioActionStatusPending,
		Notes:  "Original notes",
	}

	action.Reject(userID, "")

	assert.Equal(t, PortfolioActionStatusRejected, action.Status)
	assert.Equal(t, "Original notes", action.Notes) // Notes unchanged
}

func TestPortfolioAction_MarkApplied(t *testing.T) {
	action := &PortfolioAction{Status: PortfolioActionStatusApproved}

	action.MarkApplied()

	assert.Equal(t, PortfolioActionStatusApplied, action.Status)
	assert.NotNil(t, action.AppliedAt)
}
