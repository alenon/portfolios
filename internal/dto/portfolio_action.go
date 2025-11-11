package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// PortfolioActionResponse represents a pending corporate action for a portfolio
type PortfolioActionResponse struct {
	ID              string                   `json:"id"`
	PortfolioID     string                   `json:"portfolio_id"`
	Status          string                   `json:"status"`
	AffectedSymbol  string                   `json:"affected_symbol"`
	SharesAffected  int64                    `json:"shares_affected"`
	DetectedAt      time.Time                `json:"detected_at"`
	ReviewedAt      *time.Time               `json:"reviewed_at,omitempty"`
	AppliedAt       *time.Time               `json:"applied_at,omitempty"`
	Notes           string                   `json:"notes,omitempty"`
	CorporateAction *CorporateActionResponse `json:"corporate_action"`
}

// CorporateActionResponse represents a corporate action in API responses
type CorporateActionResponse struct {
	ID          string           `json:"id"`
	Symbol      string           `json:"symbol"`
	Type        string           `json:"type"`
	Date        time.Time        `json:"date"`
	Ratio       *decimal.Decimal `json:"ratio,omitempty"`
	Amount      *decimal.Decimal `json:"amount,omitempty"`
	NewSymbol   *string          `json:"new_symbol,omitempty"`
	Currency    *string          `json:"currency,omitempty"`
	Description string           `json:"description,omitempty"`
	Applied     bool             `json:"applied"`
	CreatedAt   time.Time        `json:"created_at"`
}

// ApproveActionRequest represents a request to approve a pending action
type ApproveActionRequest struct {
	Notes string `json:"notes,omitempty"`
}

// RejectActionRequest represents a request to reject a pending action
type RejectActionRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// PortfolioActionSummary provides a summary of pending actions for a portfolio
type PortfolioActionSummary struct {
	PortfolioID   string `json:"portfolio_id"`
	PendingCount  int    `json:"pending_count"`
	ApprovedCount int    `json:"approved_count"`
	RejectedCount int    `json:"rejected_count"`
	AppliedCount  int    `json:"applied_count"`
}
