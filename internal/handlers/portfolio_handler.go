package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/services"
)

// PortfolioHandler handles portfolio-related HTTP requests
type PortfolioHandler struct {
	portfolioService services.PortfolioService
}

// NewPortfolioHandler creates a new PortfolioHandler instance
func NewPortfolioHandler(portfolioService services.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{
		portfolioService: portfolioService,
	}
}

// Create handles portfolio creation
// POST /api/v1/portfolios
func (h *PortfolioHandler) Create(c *gin.Context) {
	var req dto.CreatePortfolioRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// Create portfolio
	portfolio, err := h.portfolioService.Create(
		userID.(string),
		req.Name,
		req.Description,
		req.BaseCurrency,
		req.CostBasisMethod,
	)
	if err != nil {
		// Check for duplicate name error
		if err == models.ErrPortfolioDuplicateName {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: "A portfolio with this name already exists",
				Code:  "DUPLICATE_PORTFOLIO_NAME",
			})
			return
		}

		// Check for validation errors
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: err.Error(),
				Code:  "VALIDATION_ERROR",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to create portfolio",
			Code:  "CREATION_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.ToPortfolioResponse(portfolio))
}

// GetAll retrieves all portfolios for the authenticated user
// GET /api/v1/portfolios
func (h *PortfolioHandler) GetAll(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// Get all portfolios
	portfolios, err := h.portfolioService.GetAllByUserID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve portfolios",
			Code:  "RETRIEVAL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToPortfolioListResponse(portfolios))
}

// GetByID retrieves a specific portfolio
// GET /api/v1/portfolios/:id
func (h *PortfolioHandler) GetByID(c *gin.Context) {
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

	// Get portfolio
	portfolio, err := h.portfolioService.GetByID(portfolioID, userID.(string))
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
				Error: "Access denied",
				Code:  "FORBIDDEN",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve portfolio",
			Code:  "RETRIEVAL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToPortfolioResponse(portfolio))
}

// Update updates a portfolio
// PUT /api/v1/portfolios/:id
func (h *PortfolioHandler) Update(c *gin.Context) {
	portfolioID := c.Param("id")
	if portfolioID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Portfolio ID is required",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	var req dto.UpdatePortfolioRequest
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

	// Update portfolio
	portfolio, err := h.portfolioService.Update(portfolioID, userID.(string), req.Name, req.Description)
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
				Error: "Access denied",
				Code:  "FORBIDDEN",
			})
			return
		}

		if err == models.ErrPortfolioDuplicateName {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: "A portfolio with this name already exists",
				Code:  "DUPLICATE_PORTFOLIO_NAME",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to update portfolio",
			Code:  "UPDATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToPortfolioResponse(portfolio))
}

// Delete deletes a portfolio
// DELETE /api/v1/portfolios/:id
func (h *PortfolioHandler) Delete(c *gin.Context) {
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

	// Delete portfolio
	err := h.portfolioService.Delete(portfolioID, userID.(string))
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
				Error: "Access denied",
				Code:  "FORBIDDEN",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to delete portfolio",
			Code:  "DELETE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Portfolio deleted successfully",
	})
}
