package dto

import (
	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/models"
	"github.com/shopspring/decimal"
	"time"
)

// HoldingResponse represents a holding in API responses
type HoldingResponse struct {
	ID           uuid.UUID       `json:"id"`
	PortfolioID  uuid.UUID       `json:"portfolio_id"`
	Symbol       string          `json:"symbol"`
	Quantity     decimal.Decimal `json:"quantity"`
	CostBasis    decimal.Decimal `json:"cost_basis"`
	AvgCostPrice decimal.Decimal `json:"avg_cost_price"`
	UpdatedAt    time.Time       `json:"updated_at"`
	// Optional fields for enriched responses
	MarketPrice          *decimal.Decimal `json:"market_price,omitempty"`
	MarketValue          *decimal.Decimal `json:"market_value,omitempty"`
	UnrealizedGain       *decimal.Decimal `json:"unrealized_gain,omitempty"`
	UnrealizedGainPct    *decimal.Decimal `json:"unrealized_gain_pct,omitempty"`
	DayChange            *decimal.Decimal `json:"day_change,omitempty"`
	DayChangePct         *decimal.Decimal `json:"day_change_pct,omitempty"`
	AllocationPercentage *decimal.Decimal `json:"allocation_percentage,omitempty"`
}

// HoldingListResponse represents a list of holdings with summary
type HoldingListResponse struct {
	Holdings []*HoldingResponse `json:"holdings"`
	Total    int                `json:"total"`
	Summary  *HoldingSummary    `json:"summary,omitempty"`
}

// HoldingSummary provides aggregate statistics for holdings
type HoldingSummary struct {
	TotalMarketValue    decimal.Decimal `json:"total_market_value"`
	TotalCostBasis      decimal.Decimal `json:"total_cost_basis"`
	TotalUnrealizedGain decimal.Decimal `json:"total_unrealized_gain"`
	TotalGainPct        decimal.Decimal `json:"total_gain_pct"`
	TotalPositions      int             `json:"total_positions"`
}

// ToHoldingResponse converts a Holding model to HoldingResponse DTO
func ToHoldingResponse(holding *models.Holding) *HoldingResponse {
	if holding == nil {
		return nil
	}

	return &HoldingResponse{
		ID:           holding.ID,
		PortfolioID:  holding.PortfolioID,
		Symbol:       holding.Symbol,
		Quantity:     holding.Quantity,
		CostBasis:    holding.CostBasis,
		AvgCostPrice: holding.AvgCostPrice,
		UpdatedAt:    holding.UpdatedAt,
	}
}

// ToHoldingResponseWithMarketData converts a Holding model to HoldingResponse with market data
func ToHoldingResponseWithMarketData(holding *models.Holding, marketPrice decimal.Decimal, totalPortfolioValue decimal.Decimal) *HoldingResponse {
	response := ToHoldingResponse(holding)
	if response == nil {
		return nil
	}

	// Calculate market value
	marketValue := holding.Quantity.Mul(marketPrice)
	response.MarketPrice = &marketPrice
	response.MarketValue = &marketValue

	// Calculate unrealized gain
	unrealizedGain := holding.CalculateUnrealizedGain(marketPrice)
	response.UnrealizedGain = &unrealizedGain

	// Calculate unrealized gain percentage
	unrealizedGainPct := holding.CalculateUnrealizedGainPercent(marketPrice)
	response.UnrealizedGainPct = &unrealizedGainPct

	// Calculate allocation percentage
	if !totalPortfolioValue.IsZero() {
		allocationPct := marketValue.Div(totalPortfolioValue).Mul(decimal.NewFromInt(100))
		response.AllocationPercentage = &allocationPct
	}

	return response
}

// ToHoldingListResponse converts a slice of Holdings to HoldingListResponse
func ToHoldingListResponse(holdings []*models.Holding) *HoldingListResponse {
	responses := make([]*HoldingResponse, 0, len(holdings))
	for _, holding := range holdings {
		responses = append(responses, ToHoldingResponse(holding))
	}

	return &HoldingListResponse{
		Holdings: responses,
		Total:    len(holdings),
	}
}
