package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/services"
)

// PerformanceAnalyticsHandler handles performance analytics HTTP requests
type PerformanceAnalyticsHandler struct {
	analyticsService services.PerformanceAnalyticsService
}

// NewPerformanceAnalyticsHandler creates a new PerformanceAnalyticsHandler instance
func NewPerformanceAnalyticsHandler(analyticsService services.PerformanceAnalyticsService) *PerformanceAnalyticsHandler {
	return &PerformanceAnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetPerformanceMetrics retrieves comprehensive performance metrics
// GET /api/v1/portfolios/:id/performance/metrics
func (h *PerformanceAnalyticsHandler) GetPerformanceMetrics(c *gin.Context) {
	portfolioID := c.Param("id")

	// Get user ID from context
	userID, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// Parse query parameters for date range
	var req dto.PerformanceMetricsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid query parameters: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Set default date range if not provided (last year)
	startDate := req.StartDate
	endDate := req.EndDate
	if startDate.IsZero() {
		startDate = time.Now().AddDate(-1, 0, 0)
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	// Get performance metrics
	metrics, err := h.analyticsService.GetPerformanceMetrics(
		portfolioID,
		userID.(string),
		startDate,
		endDate,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToPerformanceMetricsResponse(metrics))
}

// GetTWR calculates Time-Weighted Return
// GET /api/v1/portfolios/:id/performance/twr
func (h *PerformanceAnalyticsHandler) GetTWR(c *gin.Context) {
	portfolioID := c.Param("id")

	// Get user ID from context
	userID, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// Parse query parameters
	var req dto.PerformanceMetricsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid query parameters: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Set default date range if not provided
	startDate := req.StartDate
	endDate := req.EndDate
	if startDate.IsZero() {
		startDate = time.Now().AddDate(-1, 0, 0)
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	// Calculate TWR
	twr, err := h.analyticsService.CalculateTWR(
		portfolioID,
		userID.(string),
		startDate,
		endDate,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToTWRResponse(twr))
}

// GetMWR calculates Money-Weighted Return (IRR)
// GET /api/v1/portfolios/:id/performance/mwr
func (h *PerformanceAnalyticsHandler) GetMWR(c *gin.Context) {
	portfolioID := c.Param("id")

	// Get user ID from context
	userID, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// Parse query parameters
	var req dto.PerformanceMetricsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid query parameters: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Set default date range if not provided
	startDate := req.StartDate
	endDate := req.EndDate
	if startDate.IsZero() {
		startDate = time.Now().AddDate(-1, 0, 0)
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	// Calculate MWR
	mwr, err := h.analyticsService.CalculateMWR(
		portfolioID,
		userID.(string),
		startDate,
		endDate,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToMWRResponse(mwr))
}

// GetBenchmarkComparison compares portfolio to a benchmark
// GET /api/v1/portfolios/:id/performance/benchmark
func (h *PerformanceAnalyticsHandler) GetBenchmarkComparison(c *gin.Context) {
	portfolioID := c.Param("id")

	// Get user ID from context
	userID, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// Parse query parameters
	var req dto.BenchmarkComparisonRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid query parameters: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Validate benchmark symbol
	if req.BenchmarkSymbol == "" {
		req.BenchmarkSymbol = "SPY" // Default to S&P 500
	}

	// Set default date range if not provided
	startDate := req.StartDate
	endDate := req.EndDate
	if startDate.IsZero() {
		startDate = time.Now().AddDate(-1, 0, 0)
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	// Compare to benchmark
	comparison, err := h.analyticsService.CompareToBenchmark(
		portfolioID,
		userID.(string),
		req.BenchmarkSymbol,
		startDate,
		endDate,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToBenchmarkComparisonResponse(comparison))
}

// GetAnnualizedReturn calculates annualized return
// GET /api/v1/portfolios/:id/performance/annualized
func (h *PerformanceAnalyticsHandler) GetAnnualizedReturn(c *gin.Context) {
	portfolioID := c.Param("id")

	// Get user ID from context
	userID, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "User not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// Parse query parameters
	var req dto.PerformanceMetricsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid query parameters: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Set default date range if not provided
	startDate := req.StartDate
	endDate := req.EndDate
	if startDate.IsZero() {
		startDate = time.Now().AddDate(-1, 0, 0)
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	// Calculate annualized return
	annualizedReturn, err := h.analyticsService.CalculateAnnualizedReturn(
		portfolioID,
		userID.(string),
		startDate,
		endDate,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToAnnualizedReturnResponse(annualizedReturn))
}

// handleError handles errors and returns appropriate HTTP responses
func (h *PerformanceAnalyticsHandler) handleError(c *gin.Context, err error) {
	switch err {
	case models.ErrPortfolioNotFound:
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Portfolio not found",
			Code:  "PORTFOLIO_NOT_FOUND",
		})
	case models.ErrUnauthorizedAccess:
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error: "You don't have access to this portfolio",
			Code:  "UNAUTHORIZED_ACCESS",
		})
	case models.ErrPerformanceSnapshotNotFound:
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Performance data not found for this period",
			Code:  "SNAPSHOT_NOT_FOUND",
		})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INTERNAL_ERROR",
		})
	}
}
