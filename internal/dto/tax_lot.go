package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// TaxLotResponse represents a tax lot in API responses
type TaxLotResponse struct {
	ID            string          `json:"id"`
	PortfolioID   string          `json:"portfolio_id"`
	Symbol        string          `json:"symbol"`
	PurchaseDate  time.Time       `json:"purchase_date"`
	Quantity      decimal.Decimal `json:"quantity"`
	CostBasis     decimal.Decimal `json:"cost_basis"`
	CostPerShare  decimal.Decimal `json:"cost_per_share"`
	TransactionID string          `json:"transaction_id"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// TaxLossOpportunityResponse represents a tax-loss harvesting opportunity
type TaxLossOpportunityResponse struct {
	Symbol          string          `json:"symbol"`
	CurrentQuantity decimal.Decimal `json:"current_quantity"`
	CostBasis       decimal.Decimal `json:"cost_basis"`
	CurrentValue    decimal.Decimal `json:"current_value"`
	UnrealizedLoss  decimal.Decimal `json:"unrealized_loss"`
	LossPercent     decimal.Decimal `json:"loss_percent"`
}

// TaxReportRequest represents a request to generate a tax report
type TaxReportRequest struct {
	TaxYear int `json:"tax_year" binding:"required,min=2000,max=2100"`
}

// RealizedGainResponse represents a realized gain or loss in API responses
type RealizedGainResponse struct {
	Symbol       string          `json:"symbol"`
	PurchaseDate time.Time       `json:"purchase_date"`
	SaleDate     time.Time       `json:"sale_date"`
	Quantity     decimal.Decimal `json:"quantity"`
	CostBasis    decimal.Decimal `json:"cost_basis"`
	Proceeds     decimal.Decimal `json:"proceeds"`
	Gain         decimal.Decimal `json:"gain"`
	IsLongTerm   bool            `json:"is_long_term"`
}

// TaxReportResponse represents a tax report in API responses
type TaxReportResponse struct {
	Year               int                     `json:"year"`
	ShortTermGains     []*RealizedGainResponse `json:"short_term_gains"`
	LongTermGains      []*RealizedGainResponse `json:"long_term_gains"`
	TotalShortTermGain decimal.Decimal         `json:"total_short_term_gain"`
	TotalLongTermGain  decimal.Decimal         `json:"total_long_term_gain"`
	TotalGain          decimal.Decimal         `json:"total_gain"`
}

// TaxLotAllocationRequest represents a request to allocate a sale to tax lots
type TaxLotAllocationRequest struct {
	Symbol   string          `json:"symbol" binding:"required"`
	Quantity decimal.Decimal `json:"quantity" binding:"required,gt=0"`
	Method   string          `json:"method" binding:"required,oneof=FIFO LIFO SPECIFIC_LOT"`
}

// LotAllocationResponse represents how a sale is allocated to tax lots
type LotAllocationResponse struct {
	TaxLotID     string          `json:"tax_lot_id"`
	Symbol       string          `json:"symbol"`
	PurchaseDate time.Time       `json:"purchase_date"`
	Quantity     decimal.Decimal `json:"quantity"`
	CostBasis    decimal.Decimal `json:"cost_basis"`
	SaleProceeds decimal.Decimal `json:"sale_proceeds,omitempty"`
	Gain         decimal.Decimal `json:"gain,omitempty"`
	IsLongTerm   bool            `json:"is_long_term"`
}
