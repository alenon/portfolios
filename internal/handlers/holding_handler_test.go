package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHoldingService for testing
type MockHoldingService struct {
	mock.Mock
}

func (m *MockHoldingService) GetByPortfolioID(portfolioID, userID string) ([]*models.Holding, error) {
	args := m.Called(portfolioID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Holding), args.Error(1)
}

func (m *MockHoldingService) GetByPortfolioIDAndSymbol(portfolioID, symbol, userID string) (*models.Holding, error) {
	args := m.Called(portfolioID, symbol, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Holding), args.Error(1)
}

func (m *MockHoldingService) GetPortfolioValue(portfolioID, userID string, prices map[string]decimal.Decimal) (decimal.Decimal, error) {
	args := m.Called(portfolioID, userID, prices)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func TestNewHoldingHandler(t *testing.T) {
	mockService := new(MockHoldingService)
	handler := NewHoldingHandler(mockService)
	assert.NotNil(t, handler)
}

func TestHoldingHandler_GetAll(t *testing.T) {
	portfolioID := uuid.New()
	userID := uuid.New()

	holdings := []*models.Holding{
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Symbol:       "AAPL",
			Quantity:     decimal.NewFromInt(100),
			CostBasis:    decimal.NewFromInt(15000),
			AvgCostPrice: decimal.NewFromInt(150),
		},
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Symbol:       "GOOGL",
			Quantity:     decimal.NewFromInt(50),
			CostBasis:    decimal.NewFromInt(10000),
			AvgCostPrice: decimal.NewFromInt(200),
		},
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockService := new(MockHoldingService)
		handler := NewHoldingHandler(mockService)

		mockService.On("GetByPortfolioID", portfolioID.String(), userID.String()).Return(holdings, nil)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/api/v1/portfolios/:id/holdings", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID.String())
			handler.GetAll(c)
		})

		req, _ := http.NewRequest("GET", "/api/v1/portfolios/"+portfolioID.String()+"/holdings", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.HoldingListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Holdings, 2)
		assert.Equal(t, "AAPL", response.Holdings[0].Symbol)

		mockService.AssertExpectations(t)
	})

	t.Run("portfolio not found", func(t *testing.T) {
		mockService := new(MockHoldingService)
		handler := NewHoldingHandler(mockService)

		mockService.On("GetByPortfolioID", portfolioID.String(), userID.String()).Return(nil, models.ErrPortfolioNotFound)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/api/v1/portfolios/:id/holdings", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID.String())
			handler.GetAll(c)
		})

		req, _ := http.NewRequest("GET", "/api/v1/portfolios/"+portfolioID.String()+"/holdings", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		mockService := new(MockHoldingService)
		handler := NewHoldingHandler(mockService)

		mockService.On("GetByPortfolioID", portfolioID.String(), userID.String()).Return(nil, models.ErrUnauthorizedAccess)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/api/v1/portfolios/:id/holdings", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID.String())
			handler.GetAll(c)
		})

		req, _ := http.NewRequest("GET", "/api/v1/portfolios/"+portfolioID.String()+"/holdings", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("missing authentication", func(t *testing.T) {
		mockService := new(MockHoldingService)
		handler := NewHoldingHandler(mockService)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/api/v1/portfolios/:id/holdings", handler.GetAll)

		req, _ := http.NewRequest("GET", "/api/v1/portfolios/"+portfolioID.String()+"/holdings", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHoldingHandler_GetBySymbol(t *testing.T) {
	portfolioID := uuid.New()
	userID := uuid.New()

	holding := &models.Holding{
		ID:           uuid.New(),
		PortfolioID:  portfolioID,
		Symbol:       "AAPL",
		Quantity:     decimal.NewFromInt(100),
		CostBasis:    decimal.NewFromInt(15000),
		AvgCostPrice: decimal.NewFromInt(150),
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockService := new(MockHoldingService)
		handler := NewHoldingHandler(mockService)

		mockService.On("GetByPortfolioIDAndSymbol", portfolioID.String(), "AAPL", userID.String()).Return(holding, nil)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/api/v1/portfolios/:id/holdings/:symbol", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID.String())
			handler.GetBySymbol(c)
		})

		req, _ := http.NewRequest("GET", "/api/v1/portfolios/"+portfolioID.String()+"/holdings/AAPL", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.HoldingResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "AAPL", response.Symbol)

		mockService.AssertExpectations(t)
	})

	t.Run("holding not found", func(t *testing.T) {
		mockService := new(MockHoldingService)
		handler := NewHoldingHandler(mockService)

		mockService.On("GetByPortfolioIDAndSymbol", portfolioID.String(), "TSLA", userID.String()).Return(nil, models.ErrHoldingNotFound)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/api/v1/portfolios/:id/holdings/:symbol", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID.String())
			handler.GetBySymbol(c)
		})

		req, _ := http.NewRequest("GET", "/api/v1/portfolios/"+portfolioID.String()+"/holdings/TSLA", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}
