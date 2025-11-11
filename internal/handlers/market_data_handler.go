package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/services"
)

// MarketDataHandler handles market data HTTP requests
type MarketDataHandler struct {
	marketDataService services.MarketDataService
}

// NewMarketDataHandler creates a new MarketDataHandler instance
func NewMarketDataHandler(marketDataService services.MarketDataService) *MarketDataHandler {
	return &MarketDataHandler{
		marketDataService: marketDataService,
	}
}

// GetQuote retrieves a quote for a single symbol
// GET /api/v1/market/quote/:symbol
func (h *MarketDataHandler) GetQuote(c *gin.Context) {
	symbol := c.Param("symbol")

	if symbol == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Symbol is required",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	quote, err := h.marketDataService.GetQuote(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve quote: " + err.Error(),
			Code:  "QUOTE_FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, quote)
}

// GetQuotes retrieves quotes for multiple symbols
// POST /api/v1/market/quotes
func (h *MarketDataHandler) GetQuotes(c *gin.Context) {
	var req dto.GetQuotesRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if len(req.Symbols) == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "At least one symbol is required",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Limit to reasonable number of symbols
	if len(req.Symbols) > 100 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Maximum 100 symbols allowed per request",
			Code:  "TOO_MANY_SYMBOLS",
		})
		return
	}

	quotes, err := h.marketDataService.GetQuotes(req.Symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve quotes: " + err.Error(),
			Code:  "QUOTES_FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToQuotesResponse(quotes))
}

// GetHistoricalPrices retrieves historical price data for a symbol
// GET /api/v1/market/history/:symbol
func (h *MarketDataHandler) GetHistoricalPrices(c *gin.Context) {
	symbol := c.Param("symbol")

	if symbol == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Symbol is required",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	var req dto.HistoricalPricesRequest
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

	prices, err := h.marketDataService.GetHistoricalPrices(symbol, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve historical prices: " + err.Error(),
			Code:  "HISTORICAL_DATA_FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToHistoricalPricesResponse(prices))
}

// GetExchangeRate retrieves exchange rate between two currencies
// GET /api/v1/market/exchange
func (h *MarketDataHandler) GetExchangeRate(c *gin.Context) {
	var req dto.ExchangeRateRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid query parameters: " + err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if req.From == "" || req.To == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Both 'from' and 'to' currency codes are required",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	rate, err := h.marketDataService.GetExchangeRate(req.From, req.To)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve exchange rate: " + err.Error(),
			Code:  "EXCHANGE_RATE_FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToExchangeRateResponse(req.From, req.To, rate))
}

// ClearCache clears the market data cache
// POST /api/v1/market/cache/clear
func (h *MarketDataHandler) ClearCache(c *gin.Context) {
	h.marketDataService.ClearCache()
	c.JSON(http.StatusOK, gin.H{
		"message": "Market data cache cleared successfully",
	})
}
