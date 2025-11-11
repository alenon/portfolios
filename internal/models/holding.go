package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Holding represents the current position of a symbol in a portfolio
type Holding struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	PortfolioID  uuid.UUID       `gorm:"type:uuid;not null;index" json:"portfolio_id" validate:"required"`
	Symbol       string          `gorm:"type:varchar(20);not null;index" json:"symbol" validate:"required"`
	Quantity     decimal.Decimal `gorm:"type:numeric(20,8);not null" json:"quantity" validate:"required"`
	CostBasis    decimal.Decimal `gorm:"type:numeric(20,8);not null" json:"cost_basis" validate:"required"`
	AvgCostPrice decimal.Decimal `gorm:"type:numeric(20,8);not null" json:"avg_cost_price" validate:"required"`
	UpdatedAt    time.Time       `json:"updated_at"`
	Portfolio    *Portfolio      `gorm:"foreignKey:PortfolioID" json:"portfolio,omitempty"`
}

// TableName specifies the table name for the Holding model
func (Holding) TableName() string {
	return "holdings"
}

// BeforeCreate hook to generate UUID before creating a new holding
func (h *Holding) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	if h.UpdatedAt.IsZero() {
		h.UpdatedAt = time.Now().UTC()
	}
	return nil
}

// BeforeUpdate hook to update the UpdatedAt timestamp
func (h *Holding) BeforeUpdate(tx *gorm.DB) error {
	h.UpdatedAt = time.Now().UTC()
	return nil
}

// CalculateAvgCostPrice recalculates the average cost price
func (h *Holding) CalculateAvgCostPrice() {
	if h.Quantity.IsZero() {
		h.AvgCostPrice = decimal.Zero
	} else {
		h.AvgCostPrice = h.CostBasis.Div(h.Quantity)
	}
}

// AddShares adds shares to the holding with their cost
func (h *Holding) AddShares(quantity, cost decimal.Decimal) {
	h.Quantity = h.Quantity.Add(quantity)
	h.CostBasis = h.CostBasis.Add(cost)
	h.CalculateAvgCostPrice()
}

// RemoveShares removes shares from the holding using the specified cost basis
func (h *Holding) RemoveShares(quantity, costBasis decimal.Decimal) error {
	if h.Quantity.LessThan(quantity) {
		return ErrInsufficientShares
	}
	h.Quantity = h.Quantity.Sub(quantity)
	h.CostBasis = h.CostBasis.Sub(costBasis)
	h.CalculateAvgCostPrice()
	return nil
}

// CalculateUnrealizedGain calculates the unrealized gain/loss given current market price
func (h *Holding) CalculateUnrealizedGain(marketPrice decimal.Decimal) decimal.Decimal {
	currentValue := marketPrice.Mul(h.Quantity)
	return currentValue.Sub(h.CostBasis)
}

// CalculateUnrealizedGainPercent calculates the unrealized gain/loss percentage
func (h *Holding) CalculateUnrealizedGainPercent(marketPrice decimal.Decimal) decimal.Decimal {
	if h.CostBasis.IsZero() {
		return decimal.Zero
	}
	unrealizedGain := h.CalculateUnrealizedGain(marketPrice)
	return unrealizedGain.Div(h.CostBasis).Mul(decimal.NewFromInt(100))
}
