package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/services"
)

// PerformanceSnapshotHandler handles performance snapshot HTTP requests
type PerformanceSnapshotHandler struct {
	snapshotService services.PerformanceSnapshotService
}

// NewPerformanceSnapshotHandler creates a new PerformanceSnapshotHandler instance
func NewPerformanceSnapshotHandler(snapshotService services.PerformanceSnapshotService) *PerformanceSnapshotHandler {
	return &PerformanceSnapshotHandler{
		snapshotService: snapshotService,
	}
}

// GetSnapshots retrieves performance snapshots for a portfolio
// GET /api/v1/portfolios/:id/snapshots
func (h *PerformanceSnapshotHandler) GetSnapshots(c *gin.Context) {
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
	limitStr := c.DefaultQuery("limit", "30")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		limit = 30
	}
	if limit > 365 {
		limit = 365 // Max 1 year of daily snapshots
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	snapshots, err := h.snapshotService.GetByPortfolioID(portfolioID, userID.(string), limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToPerformanceSnapshotListResponse(snapshots))
}

// GetSnapshotsByDateRange retrieves performance snapshots within a date range
// GET /api/v1/portfolios/:id/snapshots/range
func (h *PerformanceSnapshotHandler) GetSnapshotsByDateRange(c *gin.Context) {
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
	var req dto.PerformanceSnapshotRangeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid query parameters: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Set default date range if not provided (last 30 days)
	startDate := req.StartDate
	endDate := req.EndDate
	if startDate.IsZero() {
		startDate = time.Now().AddDate(0, 0, -30)
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	// Validate date range
	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "End date must be after start date",
			Code:  "INVALID_DATE_RANGE",
		})
		return
	}

	snapshots, err := h.snapshotService.GetByDateRange(portfolioID, userID.(string), startDate, endDate)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToPerformanceSnapshotListResponse(snapshots))
}

// GetLatestSnapshot retrieves the most recent snapshot for a portfolio
// GET /api/v1/portfolios/:id/snapshots/latest
func (h *PerformanceSnapshotHandler) GetLatestSnapshot(c *gin.Context) {
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

	snapshot, err := h.snapshotService.GetLatest(portfolioID, userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToPerformanceSnapshotResponse(snapshot))
}

// handleError handles errors and returns appropriate HTTP responses
func (h *PerformanceSnapshotHandler) handleError(c *gin.Context, err error) {
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
			Error: "No performance snapshots found",
			Code:  "SNAPSHOT_NOT_FOUND",
		})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
			Code:  "INTERNAL_ERROR",
		})
	}
}
