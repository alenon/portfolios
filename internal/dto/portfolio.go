package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/lenon/portfolios/internal/models"
)

// CreatePortfolioRequest represents the request to create a new portfolio
type CreatePortfolioRequest struct {
	Name            string                 `json:"name" binding:"required,min=1,max=255"`
	Description     string                 `json:"description,omitempty"`
	BaseCurrency    string                 `json:"base_currency" binding:"required,len=3"`
	CostBasisMethod models.CostBasisMethod `json:"cost_basis_method" binding:"required,oneof=FIFO LIFO SPECIFIC_LOT"`
}

// UpdatePortfolioRequest represents the request to update a portfolio
type UpdatePortfolioRequest struct {
	Name        string `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description string `json:"description,omitempty"`
}

// PortfolioResponse represents a portfolio in API responses
type PortfolioResponse struct {
	ID              uuid.UUID              `json:"id"`
	UserID          uuid.UUID              `json:"user_id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	BaseCurrency    string                 `json:"base_currency"`
	CostBasisMethod models.CostBasisMethod `json:"cost_basis_method"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// PortfolioListResponse represents a list of portfolios
type PortfolioListResponse struct {
	Portfolios []*PortfolioResponse `json:"portfolios"`
	Total      int                  `json:"total"`
}

// ToPortfolioResponse converts a Portfolio model to PortfolioResponse DTO
func ToPortfolioResponse(portfolio *models.Portfolio) *PortfolioResponse {
	if portfolio == nil {
		return nil
	}

	return &PortfolioResponse{
		ID:              portfolio.ID,
		UserID:          portfolio.UserID,
		Name:            portfolio.Name,
		Description:     portfolio.Description,
		BaseCurrency:    portfolio.BaseCurrency,
		CostBasisMethod: portfolio.CostBasisMethod,
		CreatedAt:       portfolio.CreatedAt,
		UpdatedAt:       portfolio.UpdatedAt,
	}
}

// ToPortfolioListResponse converts a list of Portfolio models to PortfolioListResponse DTO
func ToPortfolioListResponse(portfolios []*models.Portfolio) *PortfolioListResponse {
	response := &PortfolioListResponse{
		Portfolios: make([]*PortfolioResponse, 0, len(portfolios)),
		Total:      len(portfolios),
	}

	for _, portfolio := range portfolios {
		response.Portfolios = append(response.Portfolios, ToPortfolioResponse(portfolio))
	}

	return response
}
