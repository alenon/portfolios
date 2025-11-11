package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
)

// PortfolioActionHandler handles portfolio action-related HTTP requests
type PortfolioActionHandler struct {
	portfolioActionRepo    repository.PortfolioActionRepository
	portfolioRepo          repository.PortfolioRepository
	corporateActionService CorporateActionService
}

// CorporateActionService interface for applying actions
type CorporateActionService interface {
	ApplyStockSplit(portfolioID, symbol, userID string, ratio decimal.Decimal, date time.Time) error
	ApplyDividend(portfolioID, symbol, userID string, amount decimal.Decimal, date time.Time) error
	ApplyMerger(portfolioID, oldSymbol, newSymbol, userID string, ratio decimal.Decimal, date time.Time) error
}

// NewPortfolioActionHandler creates a new PortfolioActionHandler instance
func NewPortfolioActionHandler(
	portfolioActionRepo repository.PortfolioActionRepository,
	portfolioRepo repository.PortfolioRepository,
	corporateActionService CorporateActionService,
) *PortfolioActionHandler {
	return &PortfolioActionHandler{
		portfolioActionRepo:    portfolioActionRepo,
		portfolioRepo:          portfolioRepo,
		corporateActionService: corporateActionService,
	}
}

// GetPendingActions retrieves all pending corporate actions for a portfolio
// GET /api/v1/portfolios/:portfolio_id/actions/pending
func (h *PortfolioActionHandler) GetPendingActions(c *gin.Context) {
	portfolioID := c.Param("portfolio_id")

	// Get user ID from context
	userID, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// Verify portfolio belongs to user
	portfolio, err := h.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Portfolio not found",
			Code:  "PORTFOLIO_NOT_FOUND",
		})
		return
	}

	if portfolio.UserID.String() != userID.(string) {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "Access denied to this portfolio",
			Code:  "FORBIDDEN",
		})
		return
	}

	// Get pending actions
	actions, err := h.portfolioActionRepo.FindPendingByPortfolioID(portfolioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve pending actions",
			Code:  "RETRIEVAL_FAILED",
		})
		return
	}

	// Convert to response DTOs
	response := make([]*dto.PortfolioActionResponse, len(actions))
	for i, action := range actions {
		response[i] = h.toPortfolioActionResponse(action)
	}

	c.JSON(http.StatusOK, response)
}

// GetAllActions retrieves all corporate actions for a portfolio (all statuses)
// GET /api/v1/portfolios/:portfolio_id/actions
func (h *PortfolioActionHandler) GetAllActions(c *gin.Context) {
	portfolioID := c.Param("portfolio_id")

	// Get user ID from context
	userID, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// Verify portfolio belongs to user
	portfolio, err := h.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Portfolio not found",
			Code:  "PORTFOLIO_NOT_FOUND",
		})
		return
	}

	if portfolio.UserID.String() != userID.(string) {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "Access denied to this portfolio",
			Code:  "FORBIDDEN",
		})
		return
	}

	// Get all actions
	actions, err := h.portfolioActionRepo.FindByPortfolioID(portfolioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve actions",
			Code:  "RETRIEVAL_FAILED",
		})
		return
	}

	// Convert to response DTOs
	response := make([]*dto.PortfolioActionResponse, len(actions))
	for i, action := range actions {
		response[i] = h.toPortfolioActionResponse(action)
	}

	c.JSON(http.StatusOK, response)
}

// GetActionByID retrieves a specific portfolio action
// GET /api/v1/portfolios/:portfolio_id/actions/:action_id
func (h *PortfolioActionHandler) GetActionByID(c *gin.Context) {
	portfolioID := c.Param("portfolio_id")
	actionID := c.Param("action_id")

	// Get user ID from context
	userID, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// Verify portfolio belongs to user
	portfolio, err := h.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Portfolio not found",
			Code:  "PORTFOLIO_NOT_FOUND",
		})
		return
	}

	if portfolio.UserID.String() != userID.(string) {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "Access denied to this portfolio",
			Code:  "FORBIDDEN",
		})
		return
	}

	// Get the action
	action, err := h.portfolioActionRepo.FindByID(actionID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Action not found",
			Code:  "ACTION_NOT_FOUND",
		})
		return
	}

	// Verify action belongs to this portfolio
	if action.PortfolioID.String() != portfolioID {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "Action does not belong to this portfolio",
			Code:  "FORBIDDEN",
		})
		return
	}

	response := h.toPortfolioActionResponse(action)
	c.JSON(http.StatusOK, response)
}

// ApproveAction approves a pending corporate action
// POST /api/v1/portfolios/:portfolio_id/actions/:action_id/approve
func (h *PortfolioActionHandler) ApproveAction(c *gin.Context) {
	portfolioID := c.Param("portfolio_id")
	actionID := c.Param("action_id")

	var req dto.ApproveActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Notes are optional, so binding errors are not critical
		req = dto.ApproveActionRequest{}
	}

	// Get user ID from context
	userID, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// Verify portfolio belongs to user
	portfolio, err := h.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Portfolio not found",
			Code:  "PORTFOLIO_NOT_FOUND",
		})
		return
	}

	if portfolio.UserID.String() != userID.(string) {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "Access denied to this portfolio",
			Code:  "FORBIDDEN",
		})
		return
	}

	// Get the action
	action, err := h.portfolioActionRepo.FindByID(actionID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Action not found",
			Code:  "ACTION_NOT_FOUND",
		})
		return
	}

	// Verify action belongs to this portfolio
	if action.PortfolioID.String() != portfolioID {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "Action does not belong to this portfolio",
			Code:  "FORBIDDEN",
		})
		return
	}

	// Check if action is still pending
	if !action.IsPending() {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Action is not pending",
			Code:  "ACTION_NOT_PENDING",
		})
		return
	}

	// Approve the action
	uid, _ := uuid.Parse(userID.(string))
	action.Approve(uid)
	if req.Notes != "" {
		action.Notes = req.Notes
	}

	if err := h.portfolioActionRepo.Update(action); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to approve action",
			Code:  "APPROVAL_FAILED",
		})
		return
	}

	// Auto-apply the action after approval
	if action.CorporateAction != nil && h.corporateActionService != nil {
		var applyErr error
		switch action.CorporateAction.Type {
		case models.CorporateActionTypeSplit:
			if action.CorporateAction.Ratio != nil {
				applyErr = h.corporateActionService.ApplyStockSplit(
					portfolioID,
					action.AffectedSymbol,
					userID.(string),
					*action.CorporateAction.Ratio,
					action.CorporateAction.Date,
				)
			}
		case models.CorporateActionTypeDividend:
			if action.CorporateAction.Amount != nil {
				applyErr = h.corporateActionService.ApplyDividend(
					portfolioID,
					action.AffectedSymbol,
					userID.(string),
					*action.CorporateAction.Amount,
					action.CorporateAction.Date,
				)
			}
		case models.CorporateActionTypeMerger:
			if action.CorporateAction.Ratio != nil && action.CorporateAction.NewSymbol != nil {
				applyErr = h.corporateActionService.ApplyMerger(
					portfolioID,
					action.AffectedSymbol,
					*action.CorporateAction.NewSymbol,
					userID.(string),
					*action.CorporateAction.Ratio,
					action.CorporateAction.Date,
				)
			}
		}

		// If application was successful, mark the action as applied
		if applyErr == nil {
			now := time.Now().UTC()
			action.AppliedAt = &now
			if err := h.portfolioActionRepo.Update(action); err != nil {
				// Log the error but don't fail the request - the action is already approved
				// In production, this would be logged to an error tracking system
			}
		}
		// If there was an error applying, we don't fail the approval
		// The action remains approved but not applied, allowing manual retry
	}

	response := h.toPortfolioActionResponse(action)
	c.JSON(http.StatusOK, response)
}

// RejectAction rejects a pending corporate action
// POST /api/v1/portfolios/:portfolio_id/actions/:action_id/reject
func (h *PortfolioActionHandler) RejectAction(c *gin.Context) {
	portfolioID := c.Param("portfolio_id")
	actionID := c.Param("action_id")

	var req dto.RejectActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// Verify portfolio belongs to user
	portfolio, err := h.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Portfolio not found",
			Code:  "PORTFOLIO_NOT_FOUND",
		})
		return
	}

	if portfolio.UserID.String() != userID.(string) {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "Access denied to this portfolio",
			Code:  "FORBIDDEN",
		})
		return
	}

	// Get the action
	action, err := h.portfolioActionRepo.FindByID(actionID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Action not found",
			Code:  "ACTION_NOT_FOUND",
		})
		return
	}

	// Verify action belongs to this portfolio
	if action.PortfolioID.String() != portfolioID {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "Action does not belong to this portfolio",
			Code:  "FORBIDDEN",
		})
		return
	}

	// Check if action is still pending
	if !action.IsPending() {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Action is not pending",
			Code:  "ACTION_NOT_PENDING",
		})
		return
	}

	// Reject the action
	uid, _ := uuid.Parse(userID.(string))
	action.Reject(uid, req.Reason)

	if err := h.portfolioActionRepo.Update(action); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to reject action",
			Code:  "REJECTION_FAILED",
		})
		return
	}

	response := h.toPortfolioActionResponse(action)
	c.JSON(http.StatusOK, response)
}

// toPortfolioActionResponse converts a model to a response DTO
func (h *PortfolioActionHandler) toPortfolioActionResponse(action *models.PortfolioAction) *dto.PortfolioActionResponse {
	response := &dto.PortfolioActionResponse{
		ID:             action.ID.String(),
		PortfolioID:    action.PortfolioID.String(),
		Status:         string(action.Status),
		AffectedSymbol: action.AffectedSymbol,
		SharesAffected: action.SharesAffected,
		DetectedAt:     action.DetectedAt,
		ReviewedAt:     action.ReviewedAt,
		AppliedAt:      action.AppliedAt,
		Notes:          action.Notes,
	}

	if action.CorporateAction != nil {
		response.CorporateAction = &dto.CorporateActionResponse{
			ID:          action.CorporateAction.ID.String(),
			Symbol:      action.CorporateAction.Symbol,
			Type:        string(action.CorporateAction.Type),
			Date:        action.CorporateAction.Date,
			Ratio:       action.CorporateAction.Ratio,
			Amount:      action.CorporateAction.Amount,
			NewSymbol:   action.CorporateAction.NewSymbol,
			Currency:    action.CorporateAction.Currency,
			Description: action.CorporateAction.Description,
			Applied:     action.CorporateAction.Applied,
			CreatedAt:   action.CorporateAction.CreatedAt,
		}
	}

	return response
}
