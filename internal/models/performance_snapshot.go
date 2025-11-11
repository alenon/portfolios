package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PerformanceSnapshot represents a point-in-time snapshot of portfolio performance
type PerformanceSnapshot struct {
	ID             uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	PortfolioID    uuid.UUID        `gorm:"type:uuid;not null;index" json:"portfolio_id" validate:"required"`
	Date           time.Time        `gorm:"not null;index:idx_performance_snapshots_date" json:"date" validate:"required"`
	TotalValue     decimal.Decimal  `gorm:"type:numeric(20,8);not null" json:"total_value" validate:"required"`
	TotalCostBasis decimal.Decimal  `gorm:"type:numeric(20,8);not null" json:"total_cost_basis" validate:"required"`
	TotalReturn    decimal.Decimal  `gorm:"type:numeric(20,8);not null" json:"total_return" validate:"required"`
	TotalReturnPct decimal.Decimal  `gorm:"type:numeric(10,4);not null" json:"total_return_pct" validate:"required"`
	DayChange      *decimal.Decimal `gorm:"type:numeric(20,8)" json:"day_change,omitempty"`
	DayChangePct   *decimal.Decimal `gorm:"type:numeric(10,4)" json:"day_change_pct,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	Portfolio      *Portfolio       `gorm:"foreignKey:PortfolioID" json:"portfolio,omitempty"`
}

// TableName specifies the table name for the PerformanceSnapshot model
func (PerformanceSnapshot) TableName() string {
	return "performance_snapshots"
}

// BeforeCreate hook to generate UUID before creating a new performance snapshot
func (ps *PerformanceSnapshot) BeforeCreate(tx *gorm.DB) error {
	if ps.ID == uuid.Nil {
		ps.ID = uuid.New()
	}
	if ps.CreatedAt.IsZero() {
		ps.CreatedAt = time.Now().UTC()
	}
	return nil
}

// Validate checks if the performance snapshot has valid data
func (ps *PerformanceSnapshot) Validate() error {
	if ps.PortfolioID == uuid.Nil {
		return ErrInvalidPortfolioID
	}
	if ps.Date.IsZero() {
		return ErrInvalidDate
	}
	if ps.TotalValue.LessThan(decimal.Zero) {
		return ErrInvalidValue
	}
	if ps.TotalCostBasis.LessThan(decimal.Zero) {
		return ErrInvalidValue
	}
	return nil
}

// CalculateMetrics calculates the return metrics based on values
func (ps *PerformanceSnapshot) CalculateMetrics() {
	// Calculate total return
	ps.TotalReturn = ps.TotalValue.Sub(ps.TotalCostBasis)

	// Calculate total return percentage
	if !ps.TotalCostBasis.IsZero() {
		ps.TotalReturnPct = ps.TotalReturn.Div(ps.TotalCostBasis).Mul(decimal.NewFromInt(100))
	} else {
		ps.TotalReturnPct = decimal.Zero
	}
}

// CalculateDayChange calculates the day change based on previous snapshot
func (ps *PerformanceSnapshot) CalculateDayChange(previousValue decimal.Decimal) {
	dayChange := ps.TotalValue.Sub(previousValue)
	ps.DayChange = &dayChange

	if !previousValue.IsZero() {
		dayChangePct := dayChange.Div(previousValue).Mul(decimal.NewFromInt(100))
		ps.DayChangePct = &dayChangePct
	} else {
		zero := decimal.Zero
		ps.DayChangePct = &zero
	}
}
