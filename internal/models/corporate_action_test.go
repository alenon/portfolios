package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCorporateAction_TableName(t *testing.T) {
	action := &CorporateAction{}
	assert.Equal(t, "corporate_actions", action.TableName())
}

func TestCorporateAction_BeforeCreate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&CorporateAction{}))

	ratio := decimal.NewFromFloat(2.0)
	action := &CorporateAction{
		Symbol: "AAPL",
		Type:   CorporateActionTypeSplit,
		Date:   time.Now().UTC(),
		Ratio:  &ratio,
	}

	err = db.Create(action).Error
	require.NoError(t, err)

	// Check UUID was generated
	assert.NotEqual(t, uuid.Nil, action.ID)
	assert.False(t, action.CreatedAt.IsZero())
}

func TestCorporateAction_Validate_Split(t *testing.T) {
	ratio := decimal.NewFromFloat(2.0)
	action := &CorporateAction{
		Symbol: "AAPL",
		Type:   CorporateActionTypeSplit,
		Date:   time.Now().UTC(),
		Ratio:  &ratio,
	}

	err := action.Validate()
	assert.NoError(t, err)
}

func TestCorporateAction_Validate_SplitMissingRatio(t *testing.T) {
	action := &CorporateAction{
		Symbol: "AAPL",
		Type:   CorporateActionTypeSplit,
		Date:   time.Now().UTC(),
	}

	err := action.Validate()
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCorporateActionType, err)
}

func TestCorporateAction_Validate_SplitNegativeRatio(t *testing.T) {
	ratio := decimal.NewFromFloat(-2.0)
	action := &CorporateAction{
		Symbol: "AAPL",
		Type:   CorporateActionTypeSplit,
		Date:   time.Now().UTC(),
		Ratio:  &ratio,
	}

	err := action.Validate()
	assert.Error(t, err)
}

func TestCorporateAction_Validate_Dividend(t *testing.T) {
	amount := decimal.NewFromFloat(0.25)
	action := &CorporateAction{
		Symbol: "AAPL",
		Type:   CorporateActionTypeDividend,
		Date:   time.Now().UTC(),
		Amount: &amount,
	}

	err := action.Validate()
	assert.NoError(t, err)
}

func TestCorporateAction_Validate_DividendMissingAmount(t *testing.T) {
	action := &CorporateAction{
		Symbol: "AAPL",
		Type:   CorporateActionTypeDividend,
		Date:   time.Now().UTC(),
	}

	err := action.Validate()
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCorporateActionType, err)
}

func TestCorporateAction_Validate_Merger(t *testing.T) {
	ratio := decimal.NewFromFloat(1.5)
	newSymbol := "ABC"
	action := &CorporateAction{
		Symbol:    "XYZ",
		Type:      CorporateActionTypeMerger,
		Date:      time.Now().UTC(),
		Ratio:     &ratio,
		NewSymbol: &newSymbol,
	}

	err := action.Validate()
	assert.NoError(t, err)
}

func TestCorporateAction_Validate_MergerMissingNewSymbol(t *testing.T) {
	ratio := decimal.NewFromFloat(1.5)
	action := &CorporateAction{
		Symbol: "XYZ",
		Type:   CorporateActionTypeMerger,
		Date:   time.Now().UTC(),
		Ratio:  &ratio,
	}

	err := action.Validate()
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCorporateActionType, err)
}

func TestCorporateAction_Validate_Spinoff(t *testing.T) {
	ratio := decimal.NewFromFloat(0.5)
	newSymbol := "SPIN"
	action := &CorporateAction{
		Symbol:    "PARENT",
		Type:      CorporateActionTypeSpinoff,
		Date:      time.Now().UTC(),
		Ratio:     &ratio,
		NewSymbol: &newSymbol,
	}

	err := action.Validate()
	assert.NoError(t, err)
}

func TestCorporateAction_Validate_SpinoffMissingNewSymbol(t *testing.T) {
	ratio := decimal.NewFromFloat(0.5)
	action := &CorporateAction{
		Symbol: "PARENT",
		Type:   CorporateActionTypeSpinoff,
		Date:   time.Now().UTC(),
		Ratio:  &ratio,
	}

	err := action.Validate()
	assert.Error(t, err)
}

func TestCorporateAction_Validate_TickerChange(t *testing.T) {
	newSymbol := "NEWT"
	action := &CorporateAction{
		Symbol:    "OLDT",
		Type:      CorporateActionTypeTickerChange,
		Date:      time.Now().UTC(),
		NewSymbol: &newSymbol,
	}

	err := action.Validate()
	assert.NoError(t, err)
}

func TestCorporateAction_Validate_TickerChangeMissingNewSymbol(t *testing.T) {
	action := &CorporateAction{
		Symbol: "OLDT",
		Type:   CorporateActionTypeTickerChange,
		Date:   time.Now().UTC(),
	}

	err := action.Validate()
	assert.Error(t, err)
}

func TestCorporateAction_Validate_EmptySymbol(t *testing.T) {
	ratio := decimal.NewFromFloat(2.0)
	action := &CorporateAction{
		Symbol: "",
		Type:   CorporateActionTypeSplit,
		Date:   time.Now().UTC(),
		Ratio:  &ratio,
	}

	err := action.Validate()
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidSymbol, err)
}

func TestCorporateAction_Validate_InvalidType(t *testing.T) {
	action := &CorporateAction{
		Symbol: "AAPL",
		Type:   "INVALID",
		Date:   time.Now().UTC(),
	}

	err := action.Validate()
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCorporateActionType, err)
}
