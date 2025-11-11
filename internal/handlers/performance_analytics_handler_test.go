package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/services"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPerformanceAnalyticsService is a mock implementation
type MockPerformanceAnalyticsService struct {
	mock.Mock
}

func (m *MockPerformanceAnalyticsService) CalculateTWR(portfolioID, userID string, startDate, endDate time.Time) (*services.TWRResult, error) {
	args := m.Called(portfolioID, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.TWRResult), args.Error(1)
}

func (m *MockPerformanceAnalyticsService) CalculateMWR(portfolioID, userID string, startDate, endDate time.Time) (*services.MWRResult, error) {
	args := m.Called(portfolioID, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.MWRResult), args.Error(1)
}

func (m *MockPerformanceAnalyticsService) CalculateAnnualizedReturn(portfolioID, userID string, startDate, endDate time.Time) (*services.AnnualizedReturnResult, error) {
	args := m.Called(portfolioID, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AnnualizedReturnResult), args.Error(1)
}

func (m *MockPerformanceAnalyticsService) CompareToBenchmark(portfolioID, userID, benchmarkSymbol string, startDate, endDate time.Time) (*services.BenchmarkComparisonResult, error) {
	args := m.Called(portfolioID, userID, benchmarkSymbol, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.BenchmarkComparisonResult), args.Error(1)
}

func (m *MockPerformanceAnalyticsService) GetPerformanceMetrics(portfolioID, userID string, startDate, endDate time.Time) (*services.PerformanceMetrics, error) {
	args := m.Called(portfolioID, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.PerformanceMetrics), args.Error(1)
}

func TestNewPerformanceAnalyticsHandler(t *testing.T) {
	mockService := new(MockPerformanceAnalyticsService)
	handler := NewPerformanceAnalyticsHandler(mockService)
	assert.NotNil(t, handler)
}

func TestGetPerformanceMetrics_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockPerformanceAnalyticsService)
	handler := NewPerformanceAnalyticsHandler(mockService)

	portfolioID := "test-portfolio-id"
	userID := "test-user-id"
	startDate := time.Now().AddDate(-1, 0, 0)
	endDate := time.Now()

	expectedMetrics := &services.PerformanceMetrics{
		StartDate:           startDate,
		EndDate:             endDate,
		StartingValue:       decimal.NewFromInt(10000),
		EndingValue:         decimal.NewFromInt(12000),
		TotalReturn:         decimal.NewFromInt(2000),
		TotalReturnPct:      decimal.NewFromInt(20),
		TimeWeightedReturn:  decimal.NewFromInt(20),
		MoneyWeightedReturn: decimal.NewFromInt(18),
		AnnualizedReturn:    decimal.NewFromInt(19),
		TotalDeposits:       decimal.NewFromInt(10000),
		TotalWithdrawals:    decimal.Zero,
		NetCashFlow:         decimal.NewFromInt(10000),
		Years:               1.0,
	}

	mockService.On("GetPerformanceMetrics", portfolioID, userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(expectedMetrics, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: portfolioID}}
	c.Set(middleware.UserIDContextKey, userID)
	c.Request = httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolioID+"/performance/metrics", nil)

	handler.GetPerformanceMetrics(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetPerformanceMetrics_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockPerformanceAnalyticsService)
	handler := NewPerformanceAnalyticsHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "test-portfolio-id"}}
	// No user ID in context

	handler.GetPerformanceMetrics(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetPerformanceMetrics_PortfolioNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockPerformanceAnalyticsService)
	handler := NewPerformanceAnalyticsHandler(mockService)

	portfolioID := "test-portfolio-id"
	userID := "test-user-id"

	mockService.On("GetPerformanceMetrics", portfolioID, userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(nil, models.ErrPortfolioNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: portfolioID}}
	c.Set(middleware.UserIDContextKey, userID)
	c.Request = httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolioID+"/performance/metrics", nil)

	handler.GetPerformanceMetrics(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetTWR_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockPerformanceAnalyticsService)
	handler := NewPerformanceAnalyticsHandler(mockService)

	portfolioID := "test-portfolio-id"
	userID := "test-user-id"

	expectedResult := &services.TWRResult{
		StartDate:     time.Now().AddDate(-1, 0, 0),
		EndDate:       time.Now(),
		TWR:           decimal.NewFromFloat(0.20),
		TWRPercent:    decimal.NewFromInt(20),
		AnnualizedTWR: decimal.NewFromInt(20),
		NumPeriods:    365,
		StartingValue: decimal.NewFromInt(10000),
		EndingValue:   decimal.NewFromInt(12000),
	}

	mockService.On("CalculateTWR", portfolioID, userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(expectedResult, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: portfolioID}}
	c.Set(middleware.UserIDContextKey, userID)
	c.Request = httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolioID+"/performance/twr", nil)

	handler.GetTWR(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetMWR_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockPerformanceAnalyticsService)
	handler := NewPerformanceAnalyticsHandler(mockService)

	portfolioID := "test-portfolio-id"
	userID := "test-user-id"

	expectedResult := &services.MWRResult{
		StartDate:     time.Now().AddDate(-1, 0, 0),
		EndDate:       time.Now(),
		MWR:           decimal.NewFromFloat(0.18),
		MWRPercent:    decimal.NewFromInt(18),
		AnnualizedMWR: decimal.NewFromInt(18),
		TotalCashFlow: decimal.NewFromInt(1000),
		StartingValue: decimal.NewFromInt(10000),
		EndingValue:   decimal.NewFromInt(12000),
	}

	mockService.On("CalculateMWR", portfolioID, userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(expectedResult, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: portfolioID}}
	c.Set(middleware.UserIDContextKey, userID)
	c.Request = httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolioID+"/performance/mwr", nil)

	handler.GetMWR(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetBenchmarkComparison_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockPerformanceAnalyticsService)
	handler := NewPerformanceAnalyticsHandler(mockService)

	portfolioID := "test-portfolio-id"
	userID := "test-user-id"

	expectedResult := &services.BenchmarkComparisonResult{
		StartDate:           time.Now().AddDate(-1, 0, 0),
		EndDate:             time.Now(),
		BenchmarkSymbol:     "SPY",
		PortfolioReturn:     decimal.NewFromInt(20),
		BenchmarkReturn:     decimal.NewFromInt(15),
		Alpha:               decimal.NewFromInt(5),
		PortfolioAnnualized: decimal.NewFromInt(20),
		BenchmarkAnnualized: decimal.NewFromInt(15),
		Outperformance:      decimal.NewFromInt(5),
	}

	mockService.On("CompareToBenchmark", portfolioID, userID, "SPY", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(expectedResult, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: portfolioID}}
	c.Set(middleware.UserIDContextKey, userID)
	c.Request = httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolioID+"/performance/benchmark?benchmark_symbol=SPY", nil)

	handler.GetBenchmarkComparison(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetAnnualizedReturn_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockPerformanceAnalyticsService)
	handler := NewPerformanceAnalyticsHandler(mockService)

	portfolioID := "test-portfolio-id"
	userID := "test-user-id"

	expectedResult := &services.AnnualizedReturnResult{
		StartDate:        time.Now().AddDate(-1, 0, 0),
		EndDate:          time.Now(),
		TotalReturn:      decimal.NewFromInt(2000),
		TotalReturnPct:   decimal.NewFromInt(20),
		AnnualizedReturn: decimal.NewFromInt(20),
		Years:            1.0,
	}

	mockService.On("CalculateAnnualizedReturn", portfolioID, userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(expectedResult, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: portfolioID}}
	c.Set(middleware.UserIDContextKey, userID)
	c.Request = httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolioID+"/performance/annualized", nil)

	handler.GetAnnualizedReturn(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	mockService.AssertExpectations(t)
}
