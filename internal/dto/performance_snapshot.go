package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/models"
	"github.com/shopspring/decimal"
)

// PerformanceSnapshotRangeRequest represents request parameters for date range query
type PerformanceSnapshotRangeRequest struct {
	StartDate time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate   time.Time `form:"end_date" time_format:"2006-01-02"`
}

// PerformanceSnapshotResponse represents a performance snapshot
type PerformanceSnapshotResponse struct {
	ID             uuid.UUID        `json:"id"`
	PortfolioID    uuid.UUID        `json:"portfolio_id"`
	Date           time.Time        `json:"date"`
	TotalValue     decimal.Decimal  `json:"total_value"`
	TotalCostBasis decimal.Decimal  `json:"total_cost_basis"`
	TotalReturn    decimal.Decimal  `json:"total_return"`
	TotalReturnPct decimal.Decimal  `json:"total_return_pct"`
	DayChange      *decimal.Decimal `json:"day_change,omitempty"`
	DayChangePct   *decimal.Decimal `json:"day_change_pct,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
}

// PerformanceSnapshotListResponse represents a list of performance snapshots
type PerformanceSnapshotListResponse struct {
	Snapshots []*PerformanceSnapshotResponse `json:"snapshots"`
	Total     int                            `json:"total"`
}

// ToPerformanceSnapshotResponse converts model to DTO
func ToPerformanceSnapshotResponse(snapshot *models.PerformanceSnapshot) *PerformanceSnapshotResponse {
	if snapshot == nil {
		return nil
	}

	return &PerformanceSnapshotResponse{
		ID:             snapshot.ID,
		PortfolioID:    snapshot.PortfolioID,
		Date:           snapshot.Date,
		TotalValue:     snapshot.TotalValue,
		TotalCostBasis: snapshot.TotalCostBasis,
		TotalReturn:    snapshot.TotalReturn,
		TotalReturnPct: snapshot.TotalReturnPct,
		DayChange:      snapshot.DayChange,
		DayChangePct:   snapshot.DayChangePct,
		CreatedAt:      snapshot.CreatedAt,
	}
}

// ToPerformanceSnapshotListResponse converts list of models to DTO
func ToPerformanceSnapshotListResponse(snapshots []*models.PerformanceSnapshot) *PerformanceSnapshotListResponse {
	response := &PerformanceSnapshotListResponse{
		Snapshots: make([]*PerformanceSnapshotResponse, 0, len(snapshots)),
		Total:     len(snapshots),
	}

	for _, snapshot := range snapshots {
		response.Snapshots = append(response.Snapshots, ToPerformanceSnapshotResponse(snapshot))
	}

	return response
}
