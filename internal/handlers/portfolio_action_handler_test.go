package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/middleware"
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupActionHandlerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.User{},
		&models.Portfolio{},
		&models.CorporateAction{},
		&models.PortfolioAction{},
	)
	require.NoError(t, err)

	return db
}

func createActionHandlerTestData(t *testing.T, db *gorm.DB) (*models.User, *models.Portfolio, *models.CorporateAction, *models.PortfolioAction) {
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
	}
	require.NoError(t, db.Create(user).Error)

	portfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	require.NoError(t, db.Create(portfolio).Error)

	ratio := decimal.NewFromFloat(2.0)
	corpAction := &models.CorporateAction{
		Symbol:  "AAPL",
		Type:    models.CorporateActionTypeSplit,
		Date:    time.Now().UTC(),
		Ratio:   &ratio,
		Applied: false,
	}
	require.NoError(t, db.Create(corpAction).Error)

	portfolioAction := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		Status:            models.PortfolioActionStatusPending,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC(),
	}
	require.NoError(t, db.Create(portfolioAction).Error)

	return user, portfolio, corpAction, portfolioAction
}

func TestNewPortfolioActionHandler(t *testing.T) {
	db := setupActionHandlerTestDB(t)
	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)

	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.portfolioActionRepo)
	assert.NotNil(t, handler.portfolioRepo)
}

func TestGetPendingActions_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupActionHandlerTestDB(t)
	user, portfolio, _, _ := createActionHandlerTestData(t, db)

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	router := gin.New()
	router.GET("/api/v1/portfolios/:portfolio_id/actions/pending", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, user.ID.String())
		handler.GetPendingActions(c)
	})

	req := httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolio.ID.String()+"/actions/pending", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*dto.PortfolioActionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, "AAPL", response[0].AffectedSymbol)
	assert.Equal(t, "PENDING", response[0].Status)
}

func TestGetPendingActions_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupActionHandlerTestDB(t)
	_, portfolio, _, _ := createActionHandlerTestData(t, db)

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	router := gin.New()
	router.GET("/api/v1/portfolios/:portfolio_id/actions/pending", handler.GetPendingActions)

	req := httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolio.ID.String()+"/actions/pending", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetPendingActions_PortfolioNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupActionHandlerTestDB(t)
	user, _, _, _ := createActionHandlerTestData(t, db)

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	router := gin.New()
	router.GET("/api/v1/portfolios/:portfolio_id/actions/pending", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, user.ID.String())
		handler.GetPendingActions(c)
	})

	req := httptest.NewRequest("GET", "/api/v1/portfolios/00000000-0000-0000-0000-000000000000/actions/pending", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetPendingActions_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupActionHandlerTestDB(t)
	_, portfolio, _, _ := createActionHandlerTestData(t, db)

	// Create another user
	otherUser := &models.User{
		Email:        "other@example.com",
		PasswordHash: "hash",
	}
	require.NoError(t, db.Create(otherUser).Error)

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	router := gin.New()
	router.GET("/api/v1/portfolios/:portfolio_id/actions/pending", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, otherUser.ID.String())
		handler.GetPendingActions(c)
	})

	req := httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolio.ID.String()+"/actions/pending", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetAllActions_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupActionHandlerTestDB(t)
	user, portfolio, _, _ := createActionHandlerTestData(t, db)

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	router := gin.New()
	router.GET("/api/v1/portfolios/:portfolio_id/actions", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, user.ID.String())
		handler.GetAllActions(c)
	})

	req := httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolio.ID.String()+"/actions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*dto.PortfolioActionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
}

func TestGetActionByID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupActionHandlerTestDB(t)
	user, portfolio, _, portfolioAction := createActionHandlerTestData(t, db)

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	router := gin.New()
	router.GET("/api/v1/portfolios/:portfolio_id/actions/:action_id", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, user.ID.String())
		handler.GetActionByID(c)
	})

	req := httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolio.ID.String()+"/actions/"+portfolioAction.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PortfolioActionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, portfolioAction.ID.String(), response.ID)
	assert.Equal(t, "AAPL", response.AffectedSymbol)
}

func TestGetActionByID_ActionNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupActionHandlerTestDB(t)
	user, portfolio, _, _ := createActionHandlerTestData(t, db)

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	router := gin.New()
	router.GET("/api/v1/portfolios/:portfolio_id/actions/:action_id", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, user.ID.String())
		handler.GetActionByID(c)
	})

	req := httptest.NewRequest("GET", "/api/v1/portfolios/"+portfolio.ID.String()+"/actions/00000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetActionByID_WrongPortfolio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupActionHandlerTestDB(t)
	user, _, _, portfolioAction := createActionHandlerTestData(t, db)

	// Create another portfolio
	otherPortfolio := &models.Portfolio{
		UserID:          user.ID,
		Name:            "Other Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}
	require.NoError(t, db.Create(otherPortfolio).Error)

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	router := gin.New()
	router.GET("/api/v1/portfolios/:portfolio_id/actions/:action_id", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, user.ID.String())
		handler.GetActionByID(c)
	})

	req := httptest.NewRequest("GET", "/api/v1/portfolios/"+otherPortfolio.ID.String()+"/actions/"+portfolioAction.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestApproveAction_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupActionHandlerTestDB(t)
	user, portfolio, _, portfolioAction := createActionHandlerTestData(t, db)

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	router := gin.New()
	router.POST("/api/v1/portfolios/:portfolio_id/actions/:action_id/approve", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, user.ID.String())
		handler.ApproveAction(c)
	})

	reqBody := dto.ApproveActionRequest{Notes: "Looks good"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/portfolios/"+portfolio.ID.String()+"/actions/"+portfolioAction.ID.String()+"/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PortfolioActionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "APPROVED", response.Status)
	assert.NotNil(t, response.ReviewedAt)
	assert.Equal(t, "Looks good", response.Notes)
}

func TestApproveAction_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupActionHandlerTestDB(t)
	_, portfolio, _, portfolioAction := createActionHandlerTestData(t, db)

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	router := gin.New()
	router.POST("/api/v1/portfolios/:portfolio_id/actions/:action_id/approve", handler.ApproveAction)

	req := httptest.NewRequest("POST", "/api/v1/portfolios/"+portfolio.ID.String()+"/actions/"+portfolioAction.ID.String()+"/approve", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestApproveAction_NotPending(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupActionHandlerTestDB(t)
	user, portfolio, _, portfolioAction := createActionHandlerTestData(t, db)

	// Already approve the action
	portfolioAction.Approve(user.ID)
	require.NoError(t, db.Save(portfolioAction).Error)

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	router := gin.New()
	router.POST("/api/v1/portfolios/:portfolio_id/actions/:action_id/approve", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, user.ID.String())
		handler.ApproveAction(c)
	})

	req := httptest.NewRequest("POST", "/api/v1/portfolios/"+portfolio.ID.String()+"/actions/"+portfolioAction.ID.String()+"/approve", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRejectAction_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupActionHandlerTestDB(t)
	user, portfolio, _, portfolioAction := createActionHandlerTestData(t, db)

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	router := gin.New()
	router.POST("/api/v1/portfolios/:portfolio_id/actions/:action_id/reject", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, user.ID.String())
		handler.RejectAction(c)
	})

	reqBody := dto.RejectActionRequest{Reason: "Not needed"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/portfolios/"+portfolio.ID.String()+"/actions/"+portfolioAction.ID.String()+"/reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PortfolioActionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "REJECTED", response.Status)
	assert.NotNil(t, response.ReviewedAt)
	assert.Equal(t, "Not needed", response.Notes)
}

func TestRejectAction_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupActionHandlerTestDB(t)
	user, portfolio, _, portfolioAction := createActionHandlerTestData(t, db)

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	router := gin.New()
	router.POST("/api/v1/portfolios/:portfolio_id/actions/:action_id/reject", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, user.ID.String())
		handler.RejectAction(c)
	})

	// Missing required "reason" field
	reqBody := map[string]string{}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/portfolios/"+portfolio.ID.String()+"/actions/"+portfolioAction.ID.String()+"/reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRejectAction_NotPending(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupActionHandlerTestDB(t)
	user, portfolio, _, portfolioAction := createActionHandlerTestData(t, db)

	// Already approve the action
	portfolioAction.Approve(user.ID)
	require.NoError(t, db.Save(portfolioAction).Error)

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	router := gin.New()
	router.POST("/api/v1/portfolios/:portfolio_id/actions/:action_id/reject", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, user.ID.String())
		handler.RejectAction(c)
	})

	reqBody := dto.RejectActionRequest{Reason: "Changed my mind"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/portfolios/"+portfolio.ID.String()+"/actions/"+portfolioAction.ID.String()+"/reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestToPortfolioActionResponse(t *testing.T) {
	db := setupActionHandlerTestDB(t)
	_, _, corpAction, portfolioAction := createActionHandlerTestData(t, db)

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	// Load the action with corporate action
	loadedAction, err := portfolioActionRepo.FindByID(portfolioAction.ID.String())
	require.NoError(t, err)

	response := handler.toPortfolioActionResponse(loadedAction)

	assert.Equal(t, portfolioAction.ID.String(), response.ID)
	assert.Equal(t, "AAPL", response.AffectedSymbol)
	assert.Equal(t, int64(100), response.SharesAffected)
	assert.Equal(t, "PENDING", response.Status)
	assert.NotNil(t, response.CorporateAction)
	assert.Equal(t, corpAction.ID.String(), response.CorporateAction.ID)
	assert.Equal(t, "AAPL", response.CorporateAction.Symbol)
	assert.Equal(t, "SPLIT", response.CorporateAction.Type)
}

func TestToPortfolioActionResponse_NoCorporateAction(t *testing.T) {
	db := setupActionHandlerTestDB(t)
	user, portfolio, corpAction, _ := createActionHandlerTestData(t, db)

	// Create action without loading corporate action relationship
	portfolioAction := &models.PortfolioAction{
		PortfolioID:       portfolio.ID,
		CorporateActionID: corpAction.ID,
		Status:            models.PortfolioActionStatusPending,
		AffectedSymbol:    "AAPL",
		SharesAffected:    100,
		DetectedAt:        time.Now().UTC(),
	}

	portfolioActionRepo := repository.NewPortfolioActionRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	handler := NewPortfolioActionHandler(portfolioActionRepo, portfolioRepo)

	// Create but don't load relationships
	require.NoError(t, db.Create(portfolioAction).Error)

	// Load without preload to ensure CorporateAction is nil
	var actionWithoutRelation models.PortfolioAction
	require.NoError(t, db.Where("id = ?", portfolioAction.ID).First(&actionWithoutRelation).Error)

	response := handler.toPortfolioActionResponse(&actionWithoutRelation)

	assert.Equal(t, portfolioAction.ID.String(), response.ID)
	assert.Nil(t, response.CorporateAction)

	// Cleanup
	_ = user
}
