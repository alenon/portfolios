package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lenon/portfolios/internal/services"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMarketDataService is a mock implementation
type MockMarketDataService struct {
	mock.Mock
}

func (m *MockMarketDataService) GetQuote(symbol string) (*services.Quote, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.Quote), args.Error(1)
}

func (m *MockMarketDataService) GetQuotes(symbols []string) (map[string]*services.Quote, error) {
	args := m.Called(symbols)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*services.Quote), args.Error(1)
}

func (m *MockMarketDataService) GetHistoricalPrices(symbol string, startDate, endDate time.Time) ([]*services.HistoricalPrice, error) {
	args := m.Called(symbol, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*services.HistoricalPrice), args.Error(1)
}

func (m *MockMarketDataService) GetExchangeRate(fromCurrency, toCurrency string) (decimal.Decimal, error) {
	args := m.Called(fromCurrency, toCurrency)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockMarketDataService) RefreshCache(symbol string) error {
	args := m.Called(symbol)
	return args.Error(0)
}

func (m *MockMarketDataService) ClearCache() {
	m.Called()
}

func TestNewMarketDataHandler(t *testing.T) {
	mockService := new(MockMarketDataService)
	handler := NewMarketDataHandler(mockService)
	assert.NotNil(t, handler)
}

func TestGetQuote_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMarketDataService)
	handler := NewMarketDataHandler(mockService)

	symbol := "AAPL"
	expectedQuote := &services.Quote{
		Symbol:        symbol,
		Price:         decimal.NewFromFloat(150.25),
		Open:          decimal.NewFromFloat(149.50),
		High:          decimal.NewFromFloat(151.00),
		Low:           decimal.NewFromFloat(149.00),
		Volume:        1000000,
		PreviousClose: decimal.NewFromFloat(149.00),
		Change:        decimal.NewFromFloat(1.25),
		ChangePercent: decimal.NewFromFloat(0.84),
		LastUpdated:   time.Now(),
	}

	mockService.On("GetQuote", symbol).Return(expectedQuote, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "symbol", Value: symbol}}

	handler.GetQuote(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetQuote_EmptySymbol(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMarketDataService)
	handler := NewMarketDataHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "symbol", Value: ""}}

	handler.GetQuote(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetQuotes_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMarketDataService)
	handler := NewMarketDataHandler(mockService)

	symbols := []string{"AAPL", "GOOGL", "MSFT"}
	expectedQuotes := map[string]*services.Quote{
		"AAPL": {
			Symbol: "AAPL",
			Price:  decimal.NewFromFloat(150.25),
		},
		"GOOGL": {
			Symbol: "GOOGL",
			Price:  decimal.NewFromFloat(2800.50),
		},
		"MSFT": {
			Symbol: "MSFT",
			Price:  decimal.NewFromFloat(300.75),
		},
	}

	mockService.On("GetQuotes", symbols).Return(expectedQuotes, nil)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"symbols": symbols,
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/market/quotes", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.GetQuotes(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetQuotes_EmptySymbols(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMarketDataService)
	handler := NewMarketDataHandler(mockService)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"symbols": []string{},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/market/quotes", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.GetQuotes(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetQuotes_TooManySymbols(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMarketDataService)
	handler := NewMarketDataHandler(mockService)

	symbols := make([]string, 101)
	for i := 0; i < 101; i++ {
		symbols[i] = "SYM" + string(rune(i))
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"symbols": symbols,
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/market/quotes", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.GetQuotes(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetHistoricalPrices_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMarketDataService)
	handler := NewMarketDataHandler(mockService)

	symbol := "AAPL"
	expectedPrices := []*services.HistoricalPrice{
		{
			Date:   time.Now().AddDate(0, 0, -1),
			Open:   decimal.NewFromFloat(149.50),
			High:   decimal.NewFromFloat(151.00),
			Low:    decimal.NewFromFloat(149.00),
			Close:  decimal.NewFromFloat(150.25),
			Volume: 1000000,
		},
	}

	mockService.On("GetHistoricalPrices", symbol, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(expectedPrices, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "symbol", Value: symbol}}
	c.Request = httptest.NewRequest("GET", "/api/v1/market/history/"+symbol, nil)

	handler.GetHistoricalPrices(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetHistoricalPrices_EmptySymbol(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMarketDataService)
	handler := NewMarketDataHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "symbol", Value: ""}}
	c.Request = httptest.NewRequest("GET", "/api/v1/market/history/", nil)

	handler.GetHistoricalPrices(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetExchangeRate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMarketDataService)
	handler := NewMarketDataHandler(mockService)

	from := "USD"
	to := "EUR"
	expectedRate := decimal.NewFromFloat(0.85)

	mockService.On("GetExchangeRate", from, to).Return(expectedRate, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/market/exchange?from="+from+"&to="+to, nil)

	handler.GetExchangeRate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetExchangeRate_MissingParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMarketDataService)
	handler := NewMarketDataHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/market/exchange", nil)

	handler.GetExchangeRate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestClearCache_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMarketDataService)
	handler := NewMarketDataHandler(mockService)

	mockService.On("ClearCache").Return()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/market/cache/clear", nil)

	handler.ClearCache(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
