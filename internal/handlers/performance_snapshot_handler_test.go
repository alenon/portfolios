package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPerformanceSnapshotService is a mock implementation
type MockPerformanceSnapshotService struct {
	mock.Mock
}

func (m *MockPerformanceSnapshotService) CreateSnapshot(portfolioID, userID string, prices map[string]decimal.Decimal) (*models.PerformanceSnapshot, error) {
	args := m.Called(portfolioID, userID, prices)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PerformanceSnapshot), args.Error(1)
}

func (m *MockPerformanceSnapshotService) GetByPortfolioID(portfolioID, userID string, limit, offset int) ([]*models.PerformanceSnapshot, error) {
	args := m.Called(portfolioID, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PerformanceSnapshot), args.Error(1)
}

func (m *MockPerformanceSnapshotService) GetByDateRange(portfolioID, userID string, startDate, endDate time.Time) ([]*models.PerformanceSnapshot, error) {
	args := m.Called(portfolioID, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PerformanceSnapshot), args.Error(1)
}

func (m *MockPerformanceSnapshotService) GetLatest(portfolioID, userID string) (*models.PerformanceSnapshot, error) {
	args := m.Called(portfolioID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PerformanceSnapshot), args.Error(1)
}

func TestNewPerformanceSnapshotHandler(t *testing.T) {
	mockService := new(MockPerformanceSnapshotService)
	handler := NewPerformanceSnapshotHandler(mockService)
	assert.NotNil(t, handler)
}

func TestGetSnapshots_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockPerformanceSnapshotService)
	handler := NewPerformanceSnapshotHandler(mockService)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()

	expectedSnapshots := []*models.PerformanceSnapshot{
		{
			ID:             uuid.New(),
			PortfolioID:    uuid.MustParse(portfolioID),
			Date:           time.Now(),
			TotalValue:     decimal.NewFromInt(12000),
			TotalCostBasis: decimal.NewFromInt(10000),
			TotalReturn:    decimal.NewFromInt(2000),
			TotalReturnPct: decimal.NewFromInt(20),
		},
	}

	mockService.On("GetByPortfolioID", portfolioID, userID, 30, 0).Return(expectedSnapshots, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: portfolioID}}
	c.Set(middleware.UserIDContextKey, userID)
	c.Request = httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolioID+"/snapshots", nil)

	handler.GetSnapshots(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetSnapshots_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockPerformanceSnapshotService)
	handler := NewPerformanceSnapshotHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "test-portfolio-id"}}
	c.Request = httptest.NewRequest("GET", "/api/v1/portfolios/test-portfolio-id/snapshots", nil)
	// No user ID in context

	handler.GetSnapshots(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetSnapshots_PortfolioNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockPerformanceSnapshotService)
	handler := NewPerformanceSnapshotHandler(mockService)

	portfolioID := "test-portfolio-id"
	userID := "test-user-id"

	mockService.On("GetByPortfolioID", portfolioID, userID, 30, 0).Return(nil, models.ErrPortfolioNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: portfolioID}}
	c.Set(middleware.UserIDContextKey, userID)
	c.Request = httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolioID+"/snapshots", nil)

	handler.GetSnapshots(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetSnapshotsByDateRange_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockPerformanceSnapshotService)
	handler := NewPerformanceSnapshotHandler(mockService)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()

	expectedSnapshots := []*models.PerformanceSnapshot{
		{
			ID:             uuid.New(),
			PortfolioID:    uuid.MustParse(portfolioID),
			Date:           time.Now(),
			TotalValue:     decimal.NewFromInt(12000),
			TotalCostBasis: decimal.NewFromInt(10000),
			TotalReturn:    decimal.NewFromInt(2000),
			TotalReturnPct: decimal.NewFromInt(20),
		},
	}

	mockService.On("GetByDateRange", portfolioID, userID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(expectedSnapshots, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: portfolioID}}
	c.Set(middleware.UserIDContextKey, userID)
	c.Request = httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolioID+"/snapshots/range", nil)

	handler.GetSnapshotsByDateRange(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetSnapshotsByDateRange_InvalidDateRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockPerformanceSnapshotService)
	handler := NewPerformanceSnapshotHandler(mockService)

	portfolioID := "test-portfolio-id"
	userID := "test-user-id"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: portfolioID}}
	c.Set(middleware.UserIDContextKey, userID)
	// End date before start date
	c.Request = httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolioID+"/snapshots/range?start_date=2024-12-31&end_date=2024-01-01", nil)

	handler.GetSnapshotsByDateRange(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetLatestSnapshot_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockPerformanceSnapshotService)
	handler := NewPerformanceSnapshotHandler(mockService)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()

	expectedSnapshot := &models.PerformanceSnapshot{
		ID:             uuid.New(),
		PortfolioID:    uuid.MustParse(portfolioID),
		Date:           time.Now(),
		TotalValue:     decimal.NewFromInt(12000),
		TotalCostBasis: decimal.NewFromInt(10000),
		TotalReturn:    decimal.NewFromInt(2000),
		TotalReturnPct: decimal.NewFromInt(20),
	}

	mockService.On("GetLatest", portfolioID, userID).Return(expectedSnapshot, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: portfolioID}}
	c.Set(middleware.UserIDContextKey, userID)
	c.Request = httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolioID+"/snapshots/latest", nil)

	handler.GetLatestSnapshot(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetLatestSnapshot_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockPerformanceSnapshotService)
	handler := NewPerformanceSnapshotHandler(mockService)

	portfolioID := "test-portfolio-id"
	userID := "test-user-id"

	mockService.On("GetLatest", portfolioID, userID).Return(nil, models.ErrPerformanceSnapshotNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: portfolioID}}
	c.Set(middleware.UserIDContextKey, userID)
	c.Request = httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolioID+"/snapshots/latest", nil)

	handler.GetLatestSnapshot(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}
