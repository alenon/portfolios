package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// TaxLot represents an individual purchase lot for tax tracking
type TaxLot struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	PortfolioID   uuid.UUID       `gorm:"type:uuid;not null;index" json:"portfolio_id" validate:"required"`
	Symbol        string          `gorm:"type:varchar(20);not null;index" json:"symbol" validate:"required"`
	PurchaseDate  time.Time       `gorm:"not null" json:"purchase_date" validate:"required"`
	Quantity      decimal.Decimal `gorm:"type:numeric(20,8);not null" json:"quantity" validate:"required"`
	CostBasis     decimal.Decimal `gorm:"type:numeric(20,8);not null" json:"cost_basis" validate:"required"`
	TransactionID uuid.UUID       `gorm:"type:uuid;not null" json:"transaction_id" validate:"required"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Portfolio     *Portfolio      `gorm:"foreignKey:PortfolioID" json:"portfolio,omitempty"`
	Transaction   *Transaction    `gorm:"foreignKey:TransactionID" json:"transaction,omitempty"`
}

// TableName specifies the table name for the TaxLot model
func (TaxLot) TableName() string {
	return "tax_lots"
}

// BeforeCreate hook to generate UUID before creating a new tax lot
func (tl *TaxLot) BeforeCreate(tx *gorm.DB) error {
	if tl.ID == uuid.Nil {
		tl.ID = uuid.New()
	}
	if tl.CreatedAt.IsZero() {
		tl.CreatedAt = time.Now().UTC()
	}
	if tl.UpdatedAt.IsZero() {
		tl.UpdatedAt = time.Now().UTC()
	}
	return nil
}

// BeforeUpdate hook to update the UpdatedAt timestamp
func (tl *TaxLot) BeforeUpdate(tx *gorm.DB) error {
	tl.UpdatedAt = time.Now().UTC()
	return nil
}

// GetCostPerShare calculates the cost per share for this lot
func (tl *TaxLot) GetCostPerShare() decimal.Decimal {
	if tl.Quantity.IsZero() {
		return decimal.Zero
	}
	return tl.CostBasis.Div(tl.Quantity)
}

// CalculateGain calculates the realized gain/loss when selling from this lot
func (tl *TaxLot) CalculateGain(salePrice, quantity decimal.Decimal) decimal.Decimal {
	costBasis := tl.GetCostPerShare().Mul(quantity)
	proceeds := salePrice.Mul(quantity)
	return proceeds.Sub(costBasis)
}

// IsLongTerm determines if the holding is long-term (held > 1 year) as of the given date
func (tl *TaxLot) IsLongTerm(asOfDate time.Time) bool {
	oneYearLater := tl.PurchaseDate.AddDate(1, 0, 0)
	return asOfDate.After(oneYearLater) || asOfDate.Equal(oneYearLater)
}

// ReduceQuantity reduces the quantity of this lot when shares are sold
func (tl *TaxLot) ReduceQuantity(quantitySold decimal.Decimal) error {
	if tl.Quantity.LessThan(quantitySold) {
		return ErrInsufficientShares
	}

	costPerShare := tl.GetCostPerShare()
	tl.Quantity = tl.Quantity.Sub(quantitySold)
	tl.CostBasis = tl.Quantity.Mul(costPerShare)

	return nil
}
