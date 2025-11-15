package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/services"
	"github.com/shopspring/decimal"
)

// TaxLotHandler handles tax lot-related HTTP requests
type TaxLotHandler struct {
	taxLotService services.TaxLotService
}

// NewTaxLotHandler creates a new TaxLotHandler instance
func NewTaxLotHandler(taxLotService services.TaxLotService) *TaxLotHandler {
	return &TaxLotHandler{
		taxLotService: taxLotService,
	}
}

// GetAll retrieves all tax lots for a portfolio
// GET /api/v1/portfolios/:portfolio_id/tax-lots
func (h *TaxLotHandler) GetAll(c *gin.Context) {
	portfolioID := c.Param("id")
	symbol := c.Query("symbol")

	// Get user ID from context
	userID, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	var taxLots []*models.TaxLot
	var err error

	// If symbol is provided, filter by symbol
	if symbol != "" {
		taxLots, err = h.taxLotService.GetByPortfolioIDAndSymbol(portfolioID, symbol, userID.(string))
	} else {
		taxLots, err = h.taxLotService.GetByPortfolioID(portfolioID, userID.(string))
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
				Error: "Access denied to this portfolio",
				Code:  "FORBIDDEN",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve tax lots",
			Code:  "RETRIEVAL_FAILED",
		})
		return
	}

	// Convert to response DTOs
	response := make([]*dto.TaxLotResponse, len(taxLots))
	for i, taxLot := range taxLots {
		response[i] = &dto.TaxLotResponse{
			ID:            taxLot.ID.String(),
			PortfolioID:   taxLot.PortfolioID.String(),
			Symbol:        taxLot.Symbol,
			PurchaseDate:  taxLot.PurchaseDate,
			Quantity:      taxLot.Quantity,
			CostBasis:     taxLot.CostBasis,
			CostPerShare:  taxLot.GetCostPerShare(),
			TransactionID: taxLot.TransactionID.String(),
			CreatedAt:     taxLot.CreatedAt,
			UpdatedAt:     taxLot.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetByID retrieves a specific tax lot by ID
// GET /api/v1/tax-lots/:id
func (h *TaxLotHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	// Get user ID from context
	userID, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	taxLot, err := h.taxLotService.GetByID(id, userID.(string))
	if err != nil {
		if err == models.ErrTaxLotNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Tax lot not found",
				Code:  "TAX_LOT_NOT_FOUND",
			})
			return
		}
		if err == models.ErrUnauthorizedAccess {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error: "Access denied to this tax lot",
				Code:  "FORBIDDEN",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve tax lot",
			Code:  "RETRIEVAL_FAILED",
		})
		return
	}

	response := &dto.TaxLotResponse{
		ID:            taxLot.ID.String(),
		PortfolioID:   taxLot.PortfolioID.String(),
		Symbol:        taxLot.Symbol,
		PurchaseDate:  taxLot.PurchaseDate,
		Quantity:      taxLot.Quantity,
		CostBasis:     taxLot.CostBasis,
		CostPerShare:  taxLot.GetCostPerShare(),
		TransactionID: taxLot.TransactionID.String(),
		CreatedAt:     taxLot.CreatedAt,
		UpdatedAt:     taxLot.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// AllocateSale shows how a sale would be allocated to tax lots
// POST /api/v1/portfolios/:portfolio_id/tax-lots/allocate
func (h *TaxLotHandler) AllocateSale(c *gin.Context) {
	portfolioID := c.Param("id")

	var req dto.TaxLotAllocationRequest
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

	// Convert method string to CostBasisMethod
	var method models.CostBasisMethod
	switch req.Method {
	case "FIFO":
		method = models.CostBasisFIFO
	case "LIFO":
		method = models.CostBasisLIFO
	case "SPECIFIC_LOT":
		method = models.CostBasisSpecificLot
	default:
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid cost basis method",
			Code:  "INVALID_METHOD",
		})
		return
	}

	allocations, err := h.taxLotService.AllocateSale(
		portfolioID,
		req.Symbol,
		userID.(string),
		req.Quantity,
		method,
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
				Error: "Access denied to this portfolio",
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

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to allocate sale",
			Code:  "ALLOCATION_FAILED",
		})
		return
	}

	// Convert to response DTOs
	response := make([]*dto.LotAllocationResponse, len(allocations))
	for i, alloc := range allocations {
		response[i] = &dto.LotAllocationResponse{
			TaxLotID:     alloc.TaxLot.ID.String(),
			Symbol:       alloc.TaxLot.Symbol,
			PurchaseDate: alloc.TaxLot.PurchaseDate,
			Quantity:     alloc.Quantity,
			CostBasis:    alloc.CostBasis,
			IsLongTerm:   alloc.IsLongTerm,
		}
	}

	c.JSON(http.StatusOK, response)
}

// IdentifyTaxLossOpportunities identifies potential tax-loss harvesting opportunities
// GET /api/v1/portfolios/:portfolio_id/tax-lots/harvest
func (h *TaxLotHandler) IdentifyTaxLossOpportunities(c *gin.Context) {
	portfolioID := c.Param("id")

	// Get threshold from query params (default to -3%)
	thresholdStr := c.DefaultQuery("threshold", "-3")
	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid threshold value",
			Code:  "INVALID_THRESHOLD",
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

	opportunities, err := h.taxLotService.IdentifyTaxLossOpportunities(
		portfolioID,
		userID.(string),
		decimal.NewFromFloat(threshold),
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
				Error: "Access denied to this portfolio",
				Code:  "FORBIDDEN",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to identify tax loss opportunities",
			Code:  "IDENTIFICATION_FAILED",
		})
		return
	}

	// Convert to response DTOs
	response := make([]*dto.TaxLossOpportunityResponse, len(opportunities))
	for i, opp := range opportunities {
		response[i] = &dto.TaxLossOpportunityResponse{
			Symbol:          opp.Symbol,
			CurrentQuantity: opp.CurrentQuantity,
			CostBasis:       opp.CostBasis,
			CurrentValue:    opp.CurrentValue,
			UnrealizedLoss:  opp.UnrealizedLoss,
			LossPercent:     opp.LossPercent,
		}
	}

	c.JSON(http.StatusOK, response)
}

// GenerateTaxReport generates a tax report for a given year
// POST /api/v1/portfolios/:portfolio_id/tax-lots/report
func (h *TaxLotHandler) GenerateTaxReport(c *gin.Context) {
	portfolioID := c.Param("id")

	var req dto.TaxReportRequest
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

	report, err := h.taxLotService.GenerateTaxReport(
		portfolioID,
		userID.(string),
		req.TaxYear,
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
				Error: "Access denied to this portfolio",
				Code:  "FORBIDDEN",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to generate tax report",
			Code:  "REPORT_GENERATION_FAILED",
		})
		return
	}

	// Convert to response DTO
	shortTermGains := make([]*dto.RealizedGainResponse, len(report.ShortTermGains))
	for i, gain := range report.ShortTermGains {
		shortTermGains[i] = &dto.RealizedGainResponse{
			Symbol:       gain.Symbol,
			PurchaseDate: gain.PurchaseDate,
			SaleDate:     gain.SaleDate,
			Quantity:     gain.Quantity,
			CostBasis:    gain.CostBasis,
			Proceeds:     gain.Proceeds,
			Gain:         gain.Gain,
			IsLongTerm:   gain.IsLongTerm,
		}
	}

	longTermGains := make([]*dto.RealizedGainResponse, len(report.LongTermGains))
	for i, gain := range report.LongTermGains {
		longTermGains[i] = &dto.RealizedGainResponse{
			Symbol:       gain.Symbol,
			PurchaseDate: gain.PurchaseDate,
			SaleDate:     gain.SaleDate,
			Quantity:     gain.Quantity,
			CostBasis:    gain.CostBasis,
			Proceeds:     gain.Proceeds,
			Gain:         gain.Gain,
			IsLongTerm:   gain.IsLongTerm,
		}
	}

	response := &dto.TaxReportResponse{
		Year:               report.Year,
		ShortTermGains:     shortTermGains,
		LongTermGains:      longTermGains,
		TotalShortTermGain: report.TotalShortTermGain,
		TotalLongTermGain:  report.TotalLongTermGain,
		TotalGain:          report.TotalGain,
	}

	c.JSON(http.StatusOK, response)
}
