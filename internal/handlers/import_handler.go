package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/services"
)

// ImportHandler handles CSV import operations
type ImportHandler struct {
	importService services.CSVImportService
}

// NewImportHandler creates a new ImportHandler instance
func NewImportHandler(importService services.CSVImportService) *ImportHandler {
	return &ImportHandler{
		importService: importService,
	}
}

// ImportCSV handles CSV file import
// @Summary Import transactions from CSV
// @Description Import transactions from a CSV file in various broker formats
// @Tags imports
// @Accept json
// @Produce json
// @Param id path string true "Portfolio ID"
// @Param request body dto.CSVImportRequest true "CSV import request"
// @Success 200 {object} dto.ImportResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/portfolios/{id}/transactions/import/csv [post]
func (h *ImportHandler) ImportCSV(c *gin.Context) {
	portfolioID := c.Param("id")
	userID := c.GetString("user_id")

	var req dto.CSVImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	result, err := h.importService.ImportFromCSV(portfolioID, userID, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "portfolio not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	statusCode := http.StatusOK
	if !result.Success {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, result)
}

// ImportBulk handles bulk transaction import
// @Summary Bulk import transactions
// @Description Import multiple pre-parsed transactions in bulk
// @Tags imports
// @Accept json
// @Produce json
// @Param id path string true "Portfolio ID"
// @Param request body dto.BulkImportRequest true "Bulk import request"
// @Success 200 {object} dto.ImportResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/portfolios/{id}/transactions/import/bulk [post]
func (h *ImportHandler) ImportBulk(c *gin.Context) {
	portfolioID := c.Param("id")
	userID := c.GetString("user_id")

	var req dto.BulkImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	result, err := h.importService.ImportBulk(portfolioID, userID, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "portfolio not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	statusCode := http.StatusOK
	if !result.Success {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, result)
}

// GetImportBatches retrieves all import batches for a portfolio
// @Summary Get import batches
// @Description Retrieve all import batches for a portfolio
// @Tags imports
// @Produce json
// @Param id path string true "Portfolio ID"
// @Success 200 {object} dto.ImportBatchListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/portfolios/{id}/imports/batches [get]
func (h *ImportHandler) GetImportBatches(c *gin.Context) {
	portfolioID := c.Param("id")
	userID := c.GetString("user_id")

	result, err := h.importService.GetImportBatches(portfolioID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "portfolio not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteImportBatch deletes all transactions from a specific import batch
// @Summary Delete import batch
// @Description Delete all transactions from a specific import batch
// @Tags imports
// @Produce json
// @Param id path string true "Portfolio ID"
// @Param batch_id path string true "Import Batch ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security Bearer
// @Router /api/v1/portfolios/{id}/imports/batches/{batch_id} [delete]
func (h *ImportHandler) DeleteImportBatch(c *gin.Context) {
	portfolioID := c.Param("id")
	batchIDStr := c.Param("batch_id")
	userID := c.GetString("user_id")

	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid batch ID"})
		return
	}

	err = h.importService.DeleteImportBatch(portfolioID, userID, batchID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "portfolio not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
