package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTransactionService is a mock implementation of TransactionService
type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) Create(portfolioID, userID string, transactionType models.TransactionType, symbol string, date time.Time, quantity, price decimal.Decimal, commission decimal.Decimal, currency, notes string) (*models.Transaction, error) {
	args := m.Called(portfolioID, userID, transactionType, symbol, date, quantity, price, commission, currency, notes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetByID(id, userID string) (*models.Transaction, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetByPortfolioID(portfolioID, userID string) ([]*models.Transaction, error) {
	args := m.Called(portfolioID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetByPortfolioIDAndSymbol(portfolioID, symbol, userID string) ([]*models.Transaction, error) {
	args := m.Called(portfolioID, symbol, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

func (m *MockTransactionService) Update(id, userID string, transactionType models.TransactionType, symbol string, date time.Time, quantity, price decimal.Decimal, commission decimal.Decimal, currency, notes string) (*models.Transaction, error) {
	args := m.Called(id, userID, transactionType, symbol, date, quantity, price, commission, currency, notes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionService) Delete(id, userID string) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func TestTransactionHandler_Create(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()
		price := decimal.NewFromFloat(150.00)
		date := time.Now()

		transaction := &models.Transaction{
			ID:          uuid.New(),
			PortfolioID: uuid.MustParse(portfolioID),
			Type:        models.TransactionTypeBuy,
			Symbol:      "AAPL",
			Date:        date,
			Quantity:    decimal.NewFromInt(10),
			Price:       &price,
			Commission:  decimal.NewFromFloat(1.00),
			Currency:    "USD",
		}

		mockService.On("Create", portfolioID, userID, models.TransactionTypeBuy, "AAPL",
			mock.AnythingOfType("time.Time"), mock.Anything, mock.Anything,
			mock.Anything, "USD", "").
			Return(transaction, nil)

		router.POST("/portfolios/:portfolio_id/transactions", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Create(c)
		})

		reqBody := dto.CreateTransactionRequest{
			Type:       models.TransactionTypeBuy,
			Symbol:     "AAPL",
			Date:       date,
			Quantity:   decimal.NewFromInt(10),
			Price:      &price,
			Commission: decimal.NewFromFloat(1.00),
			Currency:   "USD",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, "/portfolios/"+portfolioID+"/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()

		router.POST("/portfolios/:portfolio_id/transactions", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Create(c)
		})

		req, _ := http.NewRequest(http.MethodPost, "/portfolios/"+portfolioID+"/transactions", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "INVALID_REQUEST", response.Code)
	})

	t.Run("missing authentication", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		portfolioID := uuid.New().String()

		router.POST("/portfolios/:portfolio_id/transactions", handler.Create)

		price := decimal.NewFromFloat(150.00)
		reqBody := dto.CreateTransactionRequest{
			Type:     models.TransactionTypeBuy,
			Symbol:   "AAPL",
			Date:     time.Now(),
			Quantity: decimal.NewFromInt(10),
			Price:    &price,
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, "/portfolios/"+portfolioID+"/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
	})

	t.Run("portfolio not found", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()
		price := decimal.NewFromFloat(150.00)

		mockService.On("Create", portfolioID, userID, models.TransactionTypeBuy, "AAPL",
			mock.AnythingOfType("time.Time"), mock.Anything, mock.Anything,
			mock.Anything, "USD", "").
			Return(nil, models.ErrPortfolioNotFound)

		router.POST("/portfolios/:portfolio_id/transactions", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Create(c)
		})

		reqBody := dto.CreateTransactionRequest{
			Type:     models.TransactionTypeBuy,
			Symbol:   "AAPL",
			Date:     time.Now(),
			Quantity: decimal.NewFromInt(10),
			Price:    &price,
			Currency: "USD",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, "/portfolios/"+portfolioID+"/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "PORTFOLIO_NOT_FOUND", response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("insufficient shares", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()
		price := decimal.NewFromFloat(150.00)

		mockService.On("Create", portfolioID, userID, models.TransactionTypeSell, "AAPL",
			mock.AnythingOfType("time.Time"), mock.Anything, mock.Anything,
			mock.Anything, "USD", "").
			Return(nil, models.ErrInsufficientShares)

		router.POST("/portfolios/:portfolio_id/transactions", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Create(c)
		})

		reqBody := dto.CreateTransactionRequest{
			Type:     models.TransactionTypeSell,
			Symbol:   "AAPL",
			Date:     time.Now(),
			Quantity: decimal.NewFromInt(10),
			Price:    &price,
			Currency: "USD",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, "/portfolios/"+portfolioID+"/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "INSUFFICIENT_SHARES", response.Code)
		mockService.AssertExpectations(t)
	})
}

func TestTransactionHandler_GetAll(t *testing.T) {
	t.Run("successful retrieval without filter", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()
		price := decimal.NewFromFloat(150.00)

		transactions := []*models.Transaction{
			{
				ID:          uuid.New(),
				PortfolioID: uuid.MustParse(portfolioID),
				Type:        models.TransactionTypeBuy,
				Symbol:      "AAPL",
				Date:        time.Now(),
				Quantity:    decimal.NewFromInt(10),
				Price:       &price,
			},
		}

		mockService.On("GetByPortfolioID", portfolioID, userID).Return(transactions, nil)

		router.GET("/portfolios/:portfolio_id/transactions", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.GetAll(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/portfolios/"+portfolioID+"/transactions", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("successful retrieval with symbol filter", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()
		price := decimal.NewFromFloat(150.00)

		transactions := []*models.Transaction{
			{
				ID:          uuid.New(),
				PortfolioID: uuid.MustParse(portfolioID),
				Type:        models.TransactionTypeBuy,
				Symbol:      "AAPL",
				Date:        time.Now(),
				Quantity:    decimal.NewFromInt(10),
				Price:       &price,
			},
		}

		mockService.On("GetByPortfolioIDAndSymbol", portfolioID, "AAPL", userID).Return(transactions, nil)

		router.GET("/portfolios/:portfolio_id/transactions", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.GetAll(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/portfolios/"+portfolioID+"/transactions?symbol=AAPL", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("portfolio not found", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		portfolioID := uuid.New().String()

		mockService.On("GetByPortfolioID", portfolioID, userID).Return(nil, models.ErrPortfolioNotFound)

		router.GET("/portfolios/:portfolio_id/transactions", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.GetAll(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/portfolios/"+portfolioID+"/transactions", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "PORTFOLIO_NOT_FOUND", response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("missing authentication", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		portfolioID := uuid.New().String()

		router.GET("/portfolios/:portfolio_id/transactions", handler.GetAll)

		req, _ := http.NewRequest(http.MethodGet, "/portfolios/"+portfolioID+"/transactions", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
	})
}

func TestTransactionHandler_GetByID(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		transactionID := uuid.New().String()
		price := decimal.NewFromFloat(150.00)

		transaction := &models.Transaction{
			ID:          uuid.MustParse(transactionID),
			PortfolioID: uuid.New(),
			Type:        models.TransactionTypeBuy,
			Symbol:      "AAPL",
			Date:        time.Now(),
			Quantity:    decimal.NewFromInt(10),
			Price:       &price,
		}

		mockService.On("GetByID", transactionID, userID).Return(transaction, nil)

		router.GET("/transactions/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.GetByID(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/transactions/"+transactionID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("transaction not found", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		transactionID := uuid.New().String()

		mockService.On("GetByID", transactionID, userID).Return(nil, models.ErrTransactionNotFound)

		router.GET("/transactions/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.GetByID(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/transactions/"+transactionID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "TRANSACTION_NOT_FOUND", response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		transactionID := uuid.New().String()

		mockService.On("GetByID", transactionID, userID).Return(nil, models.ErrUnauthorizedAccess)

		router.GET("/transactions/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.GetByID(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/transactions/"+transactionID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "FORBIDDEN", response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("missing authentication", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		transactionID := uuid.New().String()

		router.GET("/transactions/:id", handler.GetByID)

		req, _ := http.NewRequest(http.MethodGet, "/transactions/"+transactionID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
	})
}

func TestTransactionHandler_Update(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		transactionID := uuid.New().String()
		price := decimal.NewFromFloat(160.00)
		date := time.Now()

		transaction := &models.Transaction{
			ID:          uuid.MustParse(transactionID),
			PortfolioID: uuid.New(),
			Type:        models.TransactionTypeBuy,
			Symbol:      "AAPL",
			Date:        date,
			Quantity:    decimal.NewFromInt(15),
			Price:       &price,
			Commission:  decimal.NewFromFloat(2.00),
			Currency:    "USD",
		}

		mockService.On("Update", transactionID, userID, models.TransactionTypeBuy, "AAPL",
			mock.AnythingOfType("time.Time"), mock.Anything, mock.Anything,
			mock.Anything, "USD", "Updated").
			Return(transaction, nil)

		router.PUT("/transactions/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Update(c)
		})

		reqBody := dto.UpdateTransactionRequest{
			Type:       models.TransactionTypeBuy,
			Symbol:     "AAPL",
			Date:       date,
			Quantity:   decimal.NewFromInt(15),
			Price:      &price,
			Commission: decimal.NewFromFloat(2.00),
			Currency:   "USD",
			Notes:      "Updated",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPut, "/transactions/"+transactionID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		transactionID := uuid.New().String()

		router.PUT("/transactions/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Update(c)
		})

		req, _ := http.NewRequest(http.MethodPut, "/transactions/"+transactionID, bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "INVALID_REQUEST", response.Code)
	})

	t.Run("transaction not found", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		transactionID := uuid.New().String()
		price := decimal.NewFromFloat(160.00)

		mockService.On("Update", transactionID, userID, models.TransactionTypeBuy, "AAPL",
			mock.AnythingOfType("time.Time"), mock.Anything, mock.Anything,
			mock.Anything, "USD", "").
			Return(nil, models.ErrTransactionNotFound)

		router.PUT("/transactions/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Update(c)
		})

		reqBody := dto.UpdateTransactionRequest{
			Type:     models.TransactionTypeBuy,
			Symbol:   "AAPL",
			Date:     time.Now(),
			Quantity: decimal.NewFromInt(15),
			Price:    &price,
			Currency: "USD",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPut, "/transactions/"+transactionID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "TRANSACTION_NOT_FOUND", response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("missing authentication", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		transactionID := uuid.New().String()

		router.PUT("/transactions/:id", handler.Update)

		price := decimal.NewFromFloat(160.00)
		reqBody := dto.UpdateTransactionRequest{
			Type:     models.TransactionTypeBuy,
			Symbol:   "AAPL",
			Date:     time.Now(),
			Quantity: decimal.NewFromInt(15),
			Price:    &price,
			Currency: "USD",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPut, "/transactions/"+transactionID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
	})
}

func TestTransactionHandler_Delete(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		transactionID := uuid.New().String()

		mockService.On("Delete", transactionID, userID).Return(nil)

		router.DELETE("/transactions/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Delete(c)
		})

		req, _ := http.NewRequest(http.MethodDelete, "/transactions/"+transactionID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var response dto.MessageResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Transaction deleted successfully", response.Message)
	})

	t.Run("transaction not found", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		transactionID := uuid.New().String()

		mockService.On("Delete", transactionID, userID).Return(models.ErrTransactionNotFound)

		router.DELETE("/transactions/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Delete(c)
		})

		req, _ := http.NewRequest(http.MethodDelete, "/transactions/"+transactionID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "TRANSACTION_NOT_FOUND", response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		userID := uuid.New().String()
		transactionID := uuid.New().String()

		mockService.On("Delete", transactionID, userID).Return(models.ErrUnauthorizedAccess)

		router.DELETE("/transactions/:id", func(c *gin.Context) {
			c.Set(middleware.UserIDContextKey, userID)
			handler.Delete(c)
		})

		req, _ := http.NewRequest(http.MethodDelete, "/transactions/"+transactionID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "FORBIDDEN", response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("missing authentication", func(t *testing.T) {
		mockService := new(MockTransactionService)
		handler := NewTransactionHandler(mockService)
		router := setupTestRouter()

		transactionID := uuid.New().String()

		router.DELETE("/transactions/:id", handler.Delete)

		req, _ := http.NewRequest(http.MethodDelete, "/transactions/"+transactionID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dto.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "UNAUTHORIZED", response.Code)
	})
}
