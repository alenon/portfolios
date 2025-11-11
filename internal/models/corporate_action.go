package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CorporateActionType represents the type of corporate action
type CorporateActionType string

const (
	CorporateActionTypeSplit        CorporateActionType = "SPLIT"
	CorporateActionTypeDividend     CorporateActionType = "DIVIDEND"
	CorporateActionTypeMerger       CorporateActionType = "MERGER"
	CorporateActionTypeSpinoff      CorporateActionType = "SPINOFF"
	CorporateActionTypeTickerChange CorporateActionType = "TICKER_CHANGE"
)

// CorporateAction represents a corporate action that affects securities
type CorporateAction struct {
	ID          uuid.UUID           `gorm:"type:uuid;primaryKey" json:"id"`
	Symbol      string              `gorm:"type:varchar(20);not null;index" json:"symbol" validate:"required"`
	Type        CorporateActionType `gorm:"type:varchar(20);not null;index" json:"type" validate:"required"`
	Date        time.Time           `gorm:"not null;index" json:"date" validate:"required"`
	Ratio       *decimal.Decimal    `gorm:"type:numeric(20,8)" json:"ratio,omitempty"`
	Amount      *decimal.Decimal    `gorm:"type:numeric(20,8)" json:"amount,omitempty"`
	NewSymbol   *string             `gorm:"type:varchar(20)" json:"new_symbol,omitempty"`
	Currency    *string             `gorm:"type:varchar(3)" json:"currency,omitempty"`
	Description string              `gorm:"type:text" json:"description,omitempty"`
	Applied     bool                `gorm:"not null;default:false;index" json:"applied"`
	CreatedAt   time.Time           `json:"created_at"`
}

// TableName specifies the table name for the CorporateAction model
func (CorporateAction) TableName() string {
	return "corporate_actions"
}

// BeforeCreate hook to generate UUID before creating a new corporate action
func (ca *CorporateAction) BeforeCreate(tx *gorm.DB) error {
	if ca.ID == uuid.Nil {
		ca.ID = uuid.New()
	}
	if ca.CreatedAt.IsZero() {
		ca.CreatedAt = time.Now().UTC()
	}
	return nil
}

// Validate checks if the corporate action has valid data
func (ca *CorporateAction) Validate() error {
	if ca.Symbol == "" {
		return ErrInvalidSymbol
	}
	if !ca.isValidType() {
		return ErrInvalidCorporateActionType
	}

	// Type-specific validation
	switch ca.Type {
	case CorporateActionTypeSplit:
		if ca.Ratio == nil || ca.Ratio.IsZero() || ca.Ratio.IsNegative() {
			return ErrInvalidCorporateActionType
		}
	case CorporateActionTypeDividend:
		if ca.Amount == nil || ca.Amount.IsZero() || ca.Amount.IsNegative() {
			return ErrInvalidCorporateActionType
		}
	case CorporateActionTypeMerger, CorporateActionTypeTickerChange:
		if ca.NewSymbol == nil || *ca.NewSymbol == "" {
			return ErrInvalidCorporateActionType
		}
	case CorporateActionTypeSpinoff:
		if ca.NewSymbol == nil || *ca.NewSymbol == "" || ca.Ratio == nil {
			return ErrInvalidCorporateActionType
		}
	}

	return nil
}

// isValidType checks if the corporate action type is valid
func (ca *CorporateAction) isValidType() bool {
	switch ca.Type {
	case CorporateActionTypeSplit, CorporateActionTypeDividend,
		CorporateActionTypeMerger, CorporateActionTypeSpinoff,
		CorporateActionTypeTickerChange:
		return true
	default:
		return false
	}
}
