package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPortfolioService is a mock implementation of PortfolioService
type MockPortfolioService struct {
	mock.Mock
}

func (m *MockPortfolioService) Create(userID, name, description, baseCurrency string, costBasisMethod models.CostBasisMethod) (*models.Portfolio, error) {
	args := m.Called(userID, name, description, baseCurrency, costBasisMethod)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioService) GetByID(id string, userID string) (*models.Portfolio, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioService) GetAllByUserID(userID string) ([]*models.Portfolio, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioService) Update(id, userID, name, description string) (*models.Portfolio, error) {
	args := m.Called(id, userID, name, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioService) Delete(id, userID string) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestPortfolioHandler_Create(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolio := &models.Portfolio{
			ID:              uuid.New(),
			UserID:          uuid.MustParse(userID),
			Name:            "Test Portfolio",
			Description:     "Test Description",
			BaseCurrency:    "USD",
			CostBasisMethod: models.CostBasisFIFO,
		}

		mockService.On("Create", userID, "Test Portfolio", "Test Description", "USD", models.CostBasisFIFO).
			Return(portfolio, nil)

		router.POST("/portfolios", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Create(c)
		})

		reqBody := dto.CreatePortfolioRequest{
			Name:            "Test Portfolio",
			Description:     "Test Description",
			BaseCurrency:    "USD",
			CostBasisMethod: models.CostBasisFIFO,
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, "/portfolios", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()

		router.POST("/portfolios", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Create(c)
		})

		req, _ := http.NewRequest(http.MethodPost, "/portfolios", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "INVALID_REQUEST", response.Code)
	})

	t.Run("missing authentication", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		router.POST("/portfolios", handler.Create)

		reqBody := dto.CreatePortfolioRequest{
			Name:            "Test Portfolio",
			BaseCurrency:    "USD",
			CostBasisMethod: models.CostBasisFIFO,
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, "/portfolios", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
	})

	t.Run("duplicate portfolio name", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()

		mockService.On("Create", userID, "Test Portfolio", "", "USD", models.CostBasisFIFO).
			Return(nil, models.ErrPortfolioDuplicateName)

		router.POST("/portfolios", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Create(c)
		})

		reqBody := dto.CreatePortfolioRequest{
			Name:            "Test Portfolio",
			BaseCurrency:    "USD",
			CostBasisMethod: models.CostBasisFIFO,
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, "/portfolios", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "DUPLICATE_PORTFOLIO_NAME", response.Code)
		mockService.AssertExpectations(t)
	})
}

func TestPortfolioHandler_GetAll(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolios := []*models.Portfolio{
			{
				ID:              uuid.New(),
				UserID:          uuid.MustParse(userID),
				Name:            "Portfolio 1",
				BaseCurrency:    "USD",
				CostBasisMethod: models.CostBasisFIFO,
			},
			{
				ID:              uuid.New(),
				UserID:          uuid.MustParse(userID),
				Name:            "Portfolio 2",
				BaseCurrency:    "EUR",
				CostBasisMethod: models.CostBasisLIFO,
			},
		}

		mockService.On("GetAllByUserID", userID).Return(portfolios, nil)

		router.GET("/portfolios", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.GetAll(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/portfolios", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var response dto.PortfolioListResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, 2, response.Total)
	})

	t.Run("missing authentication", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		router.GET("/portfolios", handler.GetAll)

		req, _ := http.NewRequest(http.MethodGet, "/portfolios", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
	})
}

func TestPortfolioHandler_GetByID(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()
		portfolio := &models.Portfolio{
			ID:              uuid.MustParse(portfolioID),
			UserID:          uuid.MustParse(userID),
			Name:            "Test Portfolio",
			BaseCurrency:    "USD",
			CostBasisMethod: models.CostBasisFIFO,
		}

		mockService.On("GetByID", portfolioID, userID).Return(portfolio, nil)

		router.GET("/portfolios/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.GetByID(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/portfolios/"+portfolioID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("portfolio not found", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()

		mockService.On("GetByID", portfolioID, userID).Return(nil, models.ErrPortfolioNotFound)

		router.GET("/portfolios/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.GetByID(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/portfolios/"+portfolioID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "PORTFOLIO_NOT_FOUND", response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()

		mockService.On("GetByID", portfolioID, userID).Return(nil, models.ErrUnauthorizedAccess)

		router.GET("/portfolios/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.GetByID(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/portfolios/"+portfolioID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "FORBIDDEN", response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("missing authentication", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		portfolioID := uuid.New().String()

		router.GET("/portfolios/:id", handler.GetByID)

		req, _ := http.NewRequest(http.MethodGet, "/portfolios/"+portfolioID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
	})
}

func TestPortfolioHandler_Update(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()
		portfolio := &models.Portfolio{
			ID:              uuid.MustParse(portfolioID),
			UserID:          uuid.MustParse(userID),
			Name:            "Updated Portfolio",
			Description:     "Updated Description",
			BaseCurrency:    "USD",
			CostBasisMethod: models.CostBasisFIFO,
		}

		mockService.On("Update", portfolioID, userID, "Updated Portfolio", "Updated Description").
			Return(portfolio, nil)

		router.PUT("/portfolios/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Update(c)
		})

		reqBody := dto.UpdatePortfolioRequest{
			Name:        "Updated Portfolio",
			Description: "Updated Description",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPut, "/portfolios/"+portfolioID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()

		router.PUT("/portfolios/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Update(c)
		})

		req, _ := http.NewRequest(http.MethodPut, "/portfolios/"+portfolioID, bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "INVALID_REQUEST", response.Code)
	})

	t.Run("portfolio not found", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()

		mockService.On("Update", portfolioID, userID, "Updated Portfolio", "").
			Return(nil, models.ErrPortfolioNotFound)

		router.PUT("/portfolios/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Update(c)
		})

		reqBody := dto.UpdatePortfolioRequest{
			Name: "Updated Portfolio",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPut, "/portfolios/"+portfolioID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "PORTFOLIO_NOT_FOUND", response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()

		mockService.On("Update", portfolioID, userID, "Updated Portfolio", "").
			Return(nil, models.ErrUnauthorizedAccess)

		router.PUT("/portfolios/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Update(c)
		})

		reqBody := dto.UpdatePortfolioRequest{
			Name: "Updated Portfolio",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPut, "/portfolios/"+portfolioID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "FORBIDDEN", response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("duplicate portfolio name", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()

		mockService.On("Update", portfolioID, userID, "Duplicate Portfolio", "").
			Return(nil, models.ErrPortfolioDuplicateName)

		router.PUT("/portfolios/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Update(c)
		})

		reqBody := dto.UpdatePortfolioRequest{
			Name: "Duplicate Portfolio",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPut, "/portfolios/"+portfolioID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "DUPLICATE_PORTFOLIO_NAME", response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("missing authentication", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		portfolioID := uuid.New().String()

		router.PUT("/portfolios/:id", handler.Update)

		reqBody := dto.UpdatePortfolioRequest{
			Name: "Updated Portfolio",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPut, "/portfolios/"+portfolioID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
	})
}

func TestPortfolioHandler_Delete(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()

		mockService.On("Delete", portfolioID, userID).Return(nil)

		router.DELETE("/portfolios/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Delete(c)
		})

		req, _ := http.NewRequest(http.MethodDelete, "/portfolios/"+portfolioID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var response dto.MessageResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Portfolio deleted successfully", response.Message)
	})

	t.Run("portfolio not found", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()

		mockService.On("Delete", portfolioID, userID).Return(models.ErrPortfolioNotFound)

		router.DELETE("/portfolios/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Delete(c)
		})

		req, _ := http.NewRequest(http.MethodDelete, "/portfolios/"+portfolioID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "PORTFOLIO_NOT_FOUND", response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()

		mockService.On("Delete", portfolioID, userID).Return(models.ErrUnauthorizedAccess)

		router.DELETE("/portfolios/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Delete(c)
		})

		req, _ := http.NewRequest(http.MethodDelete, "/portfolios/"+portfolioID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "FORBIDDEN", response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("missing authentication", func(t *testing.T) {
		mockService := new(MockPortfolioService)
		handler := NewPortfolioHandler(mockService)
		router := setupTestRouter()

		portfolioID := uuid.New().String()

		router.DELETE("/portfolios/:id", handler.Delete)

		req, _ := http.NewRequest(http.MethodDelete, "/portfolios/"+portfolioID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
	})
}
