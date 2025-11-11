package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories for testing
type MockTaxLotRepository struct {
	mock.Mock
}

func (m *MockTaxLotRepository) Create(lot *models.TaxLot) error {
	args := m.Called(lot)
	return args.Error(0)
}

func (m *MockTaxLotRepository) FindByID(id string) (*models.TaxLot, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaxLot), args.Error(1)
}

func (m *MockTaxLotRepository) FindByPortfolioID(portfolioID string) ([]*models.TaxLot, error) {
	args := m.Called(portfolioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TaxLot), args.Error(1)
}

func (m *MockTaxLotRepository) FindByPortfolioIDAndSymbol(portfolioID, symbol string) ([]*models.TaxLot, error) {
	args := m.Called(portfolioID, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TaxLot), args.Error(1)
}

func (m *MockTaxLotRepository) Update(lot *models.TaxLot) error {
	args := m.Called(lot)
	return args.Error(0)
}

func (m *MockTaxLotRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTaxLotRepository) DeleteByPortfolioID(portfolioID string) error {
	args := m.Called(portfolioID)
	return args.Error(0)
}

func (m *MockTaxLotRepository) DeleteByPortfolioIDAndSymbol(portfolioID, symbol string) error {
	args := m.Called(portfolioID, symbol)
	return args.Error(0)
}

func TestApplyStockSplit_Success(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	portfolioID := uuid.New()
	userID := uuid.New()
	symbol := "AAPL"
	ratio := decimal.NewFromFloat(4.0) // 4:1 split
	date := time.Now()

	portfolio := &models.Portfolio{
		ID:           portfolioID,
		UserID:       userID,
		Name:         "Test Portfolio",
		BaseCurrency: "USD",
	}

	holding := &models.Holding{
		ID:           uuid.New(),
		PortfolioID:  portfolioID,
		Symbol:       symbol,
		Quantity:     decimal.NewFromInt(100),
		CostBasis:    decimal.NewFromInt(10000),
		AvgCostPrice: decimal.NewFromInt(100),
	}

	taxLot := &models.TaxLot{
		ID:          uuid.New(),
		PortfolioID: portfolioID,
		Symbol:      symbol,
		Quantity:    decimal.NewFromInt(100),
		CostBasis:   decimal.NewFromInt(10000),
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)
	holdingRepo.On("FindByPortfolioIDAndSymbol", portfolioID.String(), symbol).Return(holding, nil)
	holdingRepo.On("Update", mock.AnythingOfType("*models.Holding")).Return(nil)
	taxLotRepo.On("FindByPortfolioIDAndSymbol", portfolioID.String(), symbol).Return([]*models.TaxLot{taxLot}, nil)
	taxLotRepo.On("Update", mock.AnythingOfType("*models.TaxLot")).Return(nil)
	transactionRepo.On("Create", mock.AnythingOfType("*models.Transaction")).Return(nil)

	err := service.ApplyStockSplit(portfolioID.String(), symbol, userID.String(), ratio, date)

	assert.NoError(t, err)
	portfolioRepo.AssertExpectations(t)
	holdingRepo.AssertExpectations(t)
	taxLotRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

func TestApplyStockSplit_PortfolioNotFound(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()
	ratio := decimal.NewFromFloat(2.0)

	portfolioRepo.On("FindByID", portfolioID).Return(nil, fmt.Errorf("not found"))

	err := service.ApplyStockSplit(portfolioID, "AAPL", userID, ratio, time.Now())

	assert.Error(t, err)
	assert.Equal(t, models.ErrPortfolioNotFound, err)
	portfolioRepo.AssertExpectations(t)
}

func TestApplyStockSplit_UnauthorizedAccess(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	portfolioID := uuid.New()
	ownerID := uuid.New()
	differentUserID := uuid.New()
	ratio := decimal.NewFromFloat(2.0)

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: ownerID,
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)

	err := service.ApplyStockSplit(portfolioID.String(), "AAPL", differentUserID.String(), ratio, time.Now())

	assert.Error(t, err)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)
	portfolioRepo.AssertExpectations(t)
}

func TestApplyStockSplit_InvalidRatio(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	portfolioID := uuid.New()
	userID := uuid.New()
	invalidRatio := decimal.NewFromInt(-1) // Invalid negative ratio

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: userID,
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)

	err := service.ApplyStockSplit(portfolioID.String(), "AAPL", userID.String(), invalidRatio, time.Now())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid split ratio")
	portfolioRepo.AssertExpectations(t)
}

func TestApplyDividend_Success(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	portfolioID := uuid.New()
	userID := uuid.New()
	symbol := "AAPL"
	amount := decimal.NewFromFloat(250.50)
	date := time.Now()

	portfolio := &models.Portfolio{
		ID:           portfolioID,
		UserID:       userID,
		Name:         "Test Portfolio",
		BaseCurrency: "USD",
	}

	holding := &models.Holding{
		ID:          uuid.New(),
		PortfolioID: portfolioID,
		Symbol:      symbol,
		Quantity:    decimal.NewFromInt(100),
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)
	holdingRepo.On("FindByPortfolioIDAndSymbol", portfolioID.String(), symbol).Return(holding, nil)
	transactionRepo.On("Create", mock.AnythingOfType("*models.Transaction")).Return(nil)

	err := service.ApplyDividend(portfolioID.String(), symbol, userID.String(), amount, date)

	assert.NoError(t, err)
	portfolioRepo.AssertExpectations(t)
	holdingRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

func TestApplyDividend_InvalidAmount(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	portfolioID := uuid.New()
	userID := uuid.New()
	invalidAmount := decimal.NewFromInt(-100) // Invalid negative amount

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: userID,
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)

	err := service.ApplyDividend(portfolioID.String(), "AAPL", userID.String(), invalidAmount, time.Now())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid dividend amount")
	portfolioRepo.AssertExpectations(t)
}

func Skip_TestApplyMerger_Success(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	portfolioID := uuid.New()
	userID := uuid.New()
	oldSymbol := "FB"
	newSymbol := "META"
	ratio := decimal.NewFromFloat(1.0) // 1:1 conversion
	date := time.Now()

	portfolio := &models.Portfolio{
		ID:           portfolioID,
		UserID:       userID,
		Name:         "Test Portfolio",
		BaseCurrency: "USD",
	}

	oldHolding := &models.Holding{
		ID:          uuid.New(),
		PortfolioID: portfolioID,
		Symbol:      oldSymbol,
		Quantity:    decimal.NewFromInt(100),
		CostBasis:   decimal.NewFromInt(10000),
	}

	// New holding doesn't exist yet
	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)
	holdingRepo.On("FindByPortfolioIDAndSymbol", portfolioID.String(), oldSymbol).Return(oldHolding, nil)
	holdingRepo.On("FindByPortfolioIDAndSymbol", portfolioID.String(), newSymbol).Return(nil, fmt.Errorf("not found"))
	holdingRepo.On("Create", mock.AnythingOfType("*models.Holding")).Return(nil)
	holdingRepo.On("DeleteByPortfolioIDAndSymbol", portfolioID.String(), oldSymbol).Return(nil)
	taxLotRepo.On("FindByPortfolioIDAndSymbol", portfolioID.String(), oldSymbol).Return([]*models.TaxLot{}, nil).Once()
	taxLotRepo.On("DeleteByPortfolioIDAndSymbol", portfolioID.String(), oldSymbol).Return(nil)
	transactionRepo.On("Create", mock.AnythingOfType("*models.Transaction")).Return(nil).Times(2) // SELL and BUY

	err := service.ApplyMerger(portfolioID.String(), oldSymbol, newSymbol, userID.String(), ratio, date)

	assert.NoError(t, err)
	portfolioRepo.AssertExpectations(t)
	holdingRepo.AssertExpectations(t)
	taxLotRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

func TestApplyMerger_InvalidRatio(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	portfolioID := uuid.New()
	userID := uuid.New()
	invalidRatio := decimal.NewFromInt(0) // Invalid zero ratio

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: userID,
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)

	err := service.ApplyMerger(portfolioID.String(), "FB", "META", userID.String(), invalidRatio, time.Now())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid merger ratio")
	portfolioRepo.AssertExpectations(t)
}

// MockCorporateActionRepository for testing
type MockCorporateActionRepository struct {
	mock.Mock
}

func (m *MockCorporateActionRepository) Create(action *models.CorporateAction) error {
	args := m.Called(action)
	return args.Error(0)
}

func (m *MockCorporateActionRepository) FindByID(id string) (*models.CorporateAction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CorporateAction), args.Error(1)
}

func (m *MockCorporateActionRepository) FindBySymbol(symbol string) ([]*models.CorporateAction, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CorporateAction), args.Error(1)
}

func (m *MockCorporateActionRepository) FindBySymbolAndDate(symbol string, date time.Time) (*models.CorporateAction, error) {
	args := m.Called(symbol, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CorporateAction), args.Error(1)
}

func (m *MockCorporateActionRepository) FindBySymbolAndDateRange(symbol string, startDate, endDate time.Time) ([]*models.CorporateAction, error) {
	args := m.Called(symbol, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CorporateAction), args.Error(1)
}

func (m *MockCorporateActionRepository) FindUnapplied() ([]*models.CorporateAction, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CorporateAction), args.Error(1)
}

func (m *MockCorporateActionRepository) Update(action *models.CorporateAction) error {
	args := m.Called(action)
	return args.Error(0)
}

func (m *MockCorporateActionRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTaxLotRepository) FindByTransactionID(transactionID string) ([]*models.TaxLot, error) {
	args := m.Called(transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TaxLot), args.Error(1)
}

// CRUD Tests

func TestCorporateActionService_Create(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	symbol := "AAPL"
	actionType := models.CorporateActionTypeSplit
	date := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	ratio := decimal.NewFromFloat(4.0)
	description := "4-for-1 stock split"

	caRepo.On("Create", mock.MatchedBy(func(action *models.CorporateAction) bool {
		return action.Symbol == symbol &&
			action.Type == actionType &&
			action.Date.Equal(date) &&
			action.Ratio.Equal(ratio) &&
			action.Description == description
	})).Return(nil)

	result, err := service.Create(symbol, actionType, date, &ratio, nil, nil, nil, &description)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, symbol, result.Symbol)
	assert.Equal(t, actionType, result.Type)
	caRepo.AssertExpectations(t)
}

func TestCorporateActionService_Create_ValidationError(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	// Empty symbol should cause validation error
	symbol := ""
	actionType := models.CorporateActionTypeSplit
	date := time.Now()
	description := "test"

	result, err := service.Create(symbol, actionType, date, nil, nil, nil, nil, &description)

	assert.Error(t, err)
	assert.Nil(t, result)
	// No expectations on caRepo since Create shouldn't be called
}

func TestCorporateActionService_GetByID(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	actionID := uuid.New().String()
	action := &models.CorporateAction{
		ID:     uuid.MustParse(actionID),
		Symbol: "AAPL",
		Type:   models.CorporateActionTypeSplit,
	}

	caRepo.On("FindByID", actionID).Return(action, nil)

	result, err := service.GetByID(actionID)

	assert.NoError(t, err)
	assert.Equal(t, action, result)
	caRepo.AssertExpectations(t)
}

func TestCorporateActionService_GetByID_NotFound(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	actionID := uuid.New().String()
	expectedErr := fmt.Errorf("not found")
	caRepo.On("FindByID", actionID).Return(nil, expectedErr)

	result, err := service.GetByID(actionID)

	assert.Error(t, err)
	assert.Nil(t, result)
	caRepo.AssertExpectations(t)
}

func TestCorporateActionService_GetBySymbol(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	symbol := "AAPL"
	actions := []*models.CorporateAction{
		{
			ID:     uuid.New(),
			Symbol: symbol,
			Type:   models.CorporateActionTypeSplit,
		},
		{
			ID:     uuid.New(),
			Symbol: symbol,
			Type:   models.CorporateActionTypeDividend,
		},
	}

	caRepo.On("FindBySymbol", symbol).Return(actions, nil)

	result, err := service.GetBySymbol(symbol)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, actions, result)
	caRepo.AssertExpectations(t)
}

func TestCorporateActionService_GetBySymbolAndDateRange(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	symbol := "AAPL"
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	actions := []*models.CorporateAction{
		{
			ID:     uuid.New(),
			Symbol: symbol,
			Type:   models.CorporateActionTypeSplit,
			Date:   time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	caRepo.On("FindBySymbolAndDateRange", symbol, startDate, endDate).Return(actions, nil)

	result, err := service.GetBySymbolAndDateRange(symbol, startDate, endDate)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, actions, result)
	caRepo.AssertExpectations(t)
}

func TestCorporateActionService_GetUnapplied(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	actions := []*models.CorporateAction{
		{
			ID:      uuid.New(),
			Symbol:  "AAPL",
			Type:    models.CorporateActionTypeSplit,
			Applied: false,
		},
		{
			ID:      uuid.New(),
			Symbol:  "GOOGL",
			Type:    models.CorporateActionTypeDividend,
			Applied: false,
		},
	}

	caRepo.On("FindUnapplied").Return(actions, nil)

	result, err := service.GetUnapplied()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, actions, result)
	caRepo.AssertExpectations(t)
}

func TestCorporateActionService_MarkAsApplied(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	actionID := uuid.New().String()
	action := &models.CorporateAction{
		ID:      uuid.MustParse(actionID),
		Symbol:  "AAPL",
		Type:    models.CorporateActionTypeSplit,
		Applied: false,
	}

	caRepo.On("FindByID", actionID).Return(action, nil)
	caRepo.On("Update", mock.MatchedBy(func(a *models.CorporateAction) bool {
		return a.ID.String() == actionID && a.Applied == true
	})).Return(nil)

	err := service.MarkAsApplied(actionID)

	assert.NoError(t, err)
	caRepo.AssertExpectations(t)
}

func TestCorporateActionService_MarkAsApplied_NotFound(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	actionID := uuid.New().String()
	expectedErr := fmt.Errorf("not found")
	caRepo.On("FindByID", actionID).Return(nil, expectedErr)

	err := service.MarkAsApplied(actionID)

	assert.Error(t, err)
	caRepo.AssertExpectations(t)
}

func TestCorporateActionService_Delete(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	actionID := uuid.New().String()
	caRepo.On("Delete", actionID).Return(nil)

	err := service.Delete(actionID)

	assert.NoError(t, err)
	caRepo.AssertExpectations(t)
}

func TestCorporateActionService_Delete_Error(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	actionID := uuid.New().String()
	expectedErr := fmt.Errorf("delete error")
	caRepo.On("Delete", actionID).Return(expectedErr)

	err := service.Delete(actionID)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	caRepo.AssertExpectations(t)
}

func TestApplySpinoff_Success(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	portfolioID := uuid.New().String()
	userID := uuid.New()
	oldSymbol := "AAPL"
	newSymbol := "APPL-SPIN"
	ratio := decimal.NewFromFloat(0.5)
	date := time.Now()

	portfolio := &models.Portfolio{
		ID:           uuid.MustParse(portfolioID),
		UserID:       userID,
		Name:         "Test Portfolio",
		BaseCurrency: "USD",
	}

	parentHolding := &models.Holding{
		ID:           uuid.New(),
		PortfolioID:  uuid.MustParse(portfolioID),
		Symbol:       oldSymbol,
		Quantity:     decimal.NewFromInt(100),
		CostBasis:    decimal.NewFromInt(10000),
		AvgCostPrice: decimal.NewFromInt(100),
	}

	taxLot := &models.TaxLot{
		ID:           uuid.New(),
		PortfolioID:  uuid.MustParse(portfolioID),
		Symbol:       oldSymbol,
		PurchaseDate: time.Now().AddDate(0, -6, 0),
		Quantity:     decimal.NewFromInt(100),
		CostBasis:    decimal.NewFromInt(10000),
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)
	holdingRepo.On("FindByPortfolioIDAndSymbol", portfolioID, oldSymbol).Return(parentHolding, nil)
	holdingRepo.On("FindByPortfolioIDAndSymbol", portfolioID, newSymbol).Return(nil, fmt.Errorf("not found"))
	holdingRepo.On("Create", mock.AnythingOfType("*models.Holding")).Return(nil)
	holdingRepo.On("Update", mock.AnythingOfType("*models.Holding")).Return(nil)
	taxLotRepo.On("FindByPortfolioIDAndSymbol", portfolioID, oldSymbol).Return([]*models.TaxLot{taxLot}, nil)
	taxLotRepo.On("Create", mock.AnythingOfType("*models.TaxLot")).Return(nil)
	taxLotRepo.On("Update", mock.AnythingOfType("*models.TaxLot")).Return(nil)
	transactionRepo.On("Create", mock.AnythingOfType("*models.Transaction")).Return(nil)

	err := service.ApplySpinoff(portfolioID, oldSymbol, newSymbol, userID.String(), ratio, date)

	assert.NoError(t, err)
	portfolioRepo.AssertExpectations(t)
	holdingRepo.AssertExpectations(t)
	taxLotRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

func TestApplySpinoff_UnauthorizedAccess(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	portfolioID := uuid.New().String()
	userID := uuid.New()
	differentUserID := uuid.New()

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: differentUserID, // Different user
		Name:   "Test Portfolio",
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)

	err := service.ApplySpinoff(portfolioID, "AAPL", "AAPL-SPIN", userID.String(), decimal.NewFromFloat(0.5), time.Now())

	assert.Error(t, err)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)
	portfolioRepo.AssertExpectations(t)
}

func TestApplySpinoff_InvalidRatio(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	portfolioID := uuid.New().String()
	userID := uuid.New()

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: userID,
		Name:   "Test Portfolio",
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)

	// Test with zero ratio
	err := service.ApplySpinoff(portfolioID, "AAPL", "AAPL-SPIN", userID.String(), decimal.Zero, time.Now())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid spinoff ratio")
	portfolioRepo.AssertExpectations(t)
}

func TestApplyTickerChange_Success(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	portfolioID := uuid.New().String()
	userID := uuid.New()
	oldSymbol := "FB"
	newSymbol := "META"
	date := time.Now()

	portfolio := &models.Portfolio{
		ID:           uuid.MustParse(portfolioID),
		UserID:       userID,
		Name:         "Test Portfolio",
		BaseCurrency: "USD",
	}

	oldHolding := &models.Holding{
		ID:           uuid.New(),
		PortfolioID:  uuid.MustParse(portfolioID),
		Symbol:       oldSymbol,
		Quantity:     decimal.NewFromInt(100),
		CostBasis:    decimal.NewFromInt(10000),
		AvgCostPrice: decimal.NewFromInt(100),
	}

	taxLot := &models.TaxLot{
		ID:           uuid.New(),
		PortfolioID:  uuid.MustParse(portfolioID),
		Symbol:       oldSymbol,
		PurchaseDate: time.Now().AddDate(0, -6, 0),
		Quantity:     decimal.NewFromInt(100),
		CostBasis:    decimal.NewFromInt(10000),
	}

	txn := &models.Transaction{
		ID:          uuid.New(),
		PortfolioID: uuid.MustParse(portfolioID),
		Symbol:      oldSymbol,
		Type:        models.TransactionTypeBuy,
		Quantity:    decimal.NewFromInt(100),
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)
	holdingRepo.On("FindByPortfolioIDAndSymbol", portfolioID, oldSymbol).Return(oldHolding, nil)
	holdingRepo.On("FindByPortfolioIDAndSymbol", portfolioID, newSymbol).Return(nil, fmt.Errorf("not found"))
	holdingRepo.On("Create", mock.AnythingOfType("*models.Holding")).Return(nil)
	holdingRepo.On("DeleteByPortfolioIDAndSymbol", portfolioID, oldSymbol).Return(nil)
	taxLotRepo.On("FindByPortfolioIDAndSymbol", portfolioID, oldSymbol).Return([]*models.TaxLot{taxLot}, nil)
	taxLotRepo.On("Create", mock.AnythingOfType("*models.TaxLot")).Return(nil)
	taxLotRepo.On("DeleteByPortfolioIDAndSymbol", portfolioID, oldSymbol).Return(nil)
	transactionRepo.On("FindByPortfolioIDAndSymbol", portfolioID, oldSymbol).Return([]*models.Transaction{txn}, nil)
	transactionRepo.On("Update", mock.AnythingOfType("*models.Transaction")).Return(nil)
	transactionRepo.On("Create", mock.AnythingOfType("*models.Transaction")).Return(nil)

	err := service.ApplyTickerChange(portfolioID, oldSymbol, newSymbol, userID.String(), date)

	assert.NoError(t, err)
	portfolioRepo.AssertExpectations(t)
	holdingRepo.AssertExpectations(t)
	taxLotRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

func TestApplyTickerChange_UnauthorizedAccess(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	portfolioID := uuid.New().String()
	userID := uuid.New()
	differentUserID := uuid.New()

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: differentUserID, // Different user
		Name:   "Test Portfolio",
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)

	err := service.ApplyTickerChange(portfolioID, "FB", "META", userID.String(), time.Now())

	assert.Error(t, err)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)
	portfolioRepo.AssertExpectations(t)
}

func TestApplyTickerChange_NoHolding(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	holdingRepo := new(MockHoldingRepository)
	taxLotRepo := new(MockTaxLotRepository)
	transactionRepo := new(MockTransactionRepository)
	caRepo := new(MockCorporateActionRepository)

	service := NewCorporateActionService(caRepo, portfolioRepo, transactionRepo, holdingRepo, taxLotRepo)

	portfolioID := uuid.New().String()
	userID := uuid.New()

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: userID,
		Name:   "Test Portfolio",
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)
	holdingRepo.On("FindByPortfolioIDAndSymbol", portfolioID, "FB").Return(nil, fmt.Errorf("not found"))

	err := service.ApplyTickerChange(portfolioID, "FB", "META", userID.String(), time.Now())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no holding found")
	portfolioRepo.AssertExpectations(t)
	holdingRepo.AssertExpectations(t)
}
