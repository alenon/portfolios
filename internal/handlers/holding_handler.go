package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/services"
)

// HoldingHandler handles holding-related HTTP requests
type HoldingHandler struct {
	holdingService services.HoldingService
}

// NewHoldingHandler creates a new HoldingHandler instance
func NewHoldingHandler(holdingService services.HoldingService) *HoldingHandler {
	return &HoldingHandler{
		holdingService: holdingService,
	}
}

// GetAll retrieves all holdings for a portfolio
// GET /api/v1/portfolios/:id/holdings
func (h *HoldingHandler) GetAll(c *gin.Context) {
	// Get portfolio ID from URL parameter
	portfolioID := c.Param("id")
	if portfolioID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Portfolio ID is required",
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

	// Get all holdings for the portfolio
	holdings, err := h.holdingService.GetByPortfolioID(portfolioID, userID.(string))
	if err != nil {
		if err == models.ErrPortfolioNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Portfolio not found",
				Code:  "PORTFOLIO_NOT_FOUND",
			})
			return
		}
		if err == models.ErrUnauthorizedAccess {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error: "You don't have permission to access this portfolio",
				Code:  "FORBIDDEN",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve holdings",
			Code:  "RETRIEVAL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToHoldingListResponse(holdings))
}

// GetBySymbol retrieves a specific holding by symbol
// GET /api/v1/portfolios/:id/holdings/:symbol
func (h *HoldingHandler) GetBySymbol(c *gin.Context) {
	// Get portfolio ID and symbol from URL parameters
	portfolioID := c.Param("id")
	symbol := c.Param("symbol")

	if portfolioID == "" || symbol == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Portfolio ID and symbol are required",
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

	// Get the holding
	holding, err := h.holdingService.GetByPortfolioIDAndSymbol(portfolioID, symbol, userID.(string))
	if err != nil {
		if err == models.ErrPortfolioNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Portfolio not found",
				Code:  "PORTFOLIO_NOT_FOUND",
			})
			return
		}
		if err == models.ErrUnauthorizedAccess {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error: "You don't have permission to access this portfolio",
				Code:  "FORBIDDEN",
			})
			return
		}
		if err == models.ErrHoldingNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Holding not found",
				Code:  "HOLDING_NOT_FOUND",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve holding",
			Code:  "RETRIEVAL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToHoldingResponse(holding))
}
