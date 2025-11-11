package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/services"
)

// TransactionHandler handles transaction-related HTTP requests
type TransactionHandler struct {
	transactionService services.TransactionService
}

// NewTransactionHandler creates a new TransactionHandler instance
func NewTransactionHandler(transactionService services.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

// Create handles transaction creation
// POST /api/v1/portfolios/:portfolio_id/transactions
func (h *TransactionHandler) Create(c *gin.Context) {
	portfolioID := c.Param("portfolio_id")
	if portfolioID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Portfolio ID is required",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	var req dto.CreateTransactionRequest
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

	// Extract price or use zero
	var price decimal.Decimal
	if req.Price != nil {
		price = *req.Price
	}

	// Create transaction
	transaction, err := h.transactionService.Create(
		portfolioID,
		userID.(string),
		req.Type,
		req.Symbol,
		req.Date,
		req.Quantity,
		price,
		req.Commission,
		req.Currency,
		req.Notes,
	)
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

		if err == models.ErrInsufficientShares {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: "Insufficient shares for sale",
				Code:  "INSUFFICIENT_SHARES",
			})
			return
		}

		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: err.Error(),
				Code:  "VALIDATION_ERROR",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to create transaction: " + err.Error(),
			Code:  "CREATION_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.ToTransactionResponse(transaction))
}

// GetAll retrieves all transactions for a portfolio
// GET /api/v1/portfolios/:portfolio_id/transactions
func (h *TransactionHandler) GetAll(c *gin.Context) {
	portfolioID := c.Param("portfolio_id")
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

	// Get optional symbol filter
	symbol := c.Query("symbol")

	var transactions []*models.Transaction
	var err error

	if symbol != "" {
		transactions, err = h.transactionService.GetByPortfolioIDAndSymbol(portfolioID, symbol, userID.(string))
	} else {
		transactions, err = h.transactionService.GetByPortfolioID(portfolioID, userID.(string))
	}

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
			Error: "Failed to retrieve transactions",
			Code:  "RETRIEVAL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToTransactionListResponse(transactions))
}

// GetByID retrieves a specific transaction
// GET /api/v1/transactions/:id
func (h *TransactionHandler) GetByID(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Transaction ID is required",
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

	// Get transaction
	transaction, err := h.transactionService.GetByID(transactionID, userID.(string))
	if err != nil {
		if err == models.ErrTransactionNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Transaction not found",
				Code:  "TRANSACTION_NOT_FOUND",
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
			Error: "Failed to retrieve transaction",
			Code:  "RETRIEVAL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToTransactionResponse(transaction))
}

// Update updates a transaction
// PUT /api/v1/transactions/:id
func (h *TransactionHandler) Update(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Transaction ID is required",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	var req dto.UpdateTransactionRequest
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

	// Extract price or use zero
	var price decimal.Decimal
	if req.Price != nil {
		price = *req.Price
	}

	// Update transaction
	transaction, err := h.transactionService.Update(
		transactionID,
		userID.(string),
		req.Type,
		req.Symbol,
		req.Date,
		req.Quantity,
		price,
		req.Commission,
		req.Currency,
		req.Notes,
	)
	if err != nil {
		if err == models.ErrTransactionNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Transaction not found",
				Code:  "TRANSACTION_NOT_FOUND",
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
			Error: "Failed to update transaction",
			Code:  "UPDATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToTransactionResponse(transaction))
}

// Delete deletes a transaction
// DELETE /api/v1/transactions/:id
func (h *TransactionHandler) Delete(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Transaction ID is required",
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

	// Delete transaction
	err := h.transactionService.Delete(transactionID, userID.(string))
	if err != nil {
		if err == models.ErrTransactionNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Transaction not found",
				Code:  "TRANSACTION_NOT_FOUND",
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
			Error: "Failed to delete transaction",
			Code:  "DELETE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Transaction deleted successfully",
	})
}
