package services

import (
	"github.com/lenon/portfolios/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"time"
)

// MockHoldingRepository for testing
type MockHoldingRepository struct {
	mock.Mock
}

func (m *MockHoldingRepository) Create(holding *models.Holding) error {
	args := m.Called(holding)
	return args.Error(0)
}

func (m *MockHoldingRepository) FindByID(id string) (*models.Holding, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Holding), args.Error(1)
}

func (m *MockHoldingRepository) FindByPortfolioID(portfolioID string) ([]*models.Holding, error) {
	args := m.Called(portfolioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Holding), args.Error(1)
}

func (m *MockHoldingRepository) FindByPortfolioIDAndSymbol(portfolioID, symbol string) (*models.Holding, error) {
	args := m.Called(portfolioID, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Holding), args.Error(1)
}

func (m *MockHoldingRepository) FindBySymbol(symbol string) ([]*models.Holding, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Holding), args.Error(1)
}

func (m *MockHoldingRepository) Update(holding *models.Holding) error {
	args := m.Called(holding)
	return args.Error(0)
}

func (m *MockHoldingRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockHoldingRepository) DeleteByPortfolioID(portfolioID string) error {
	args := m.Called(portfolioID)
	return args.Error(0)
}

func (m *MockHoldingRepository) DeleteByPortfolioIDAndSymbol(portfolioID, symbol string) error {
	args := m.Called(portfolioID, symbol)
	return args.Error(0)
}

func (m *MockHoldingRepository) Upsert(holding *models.Holding) error {
	args := m.Called(holding)
	return args.Error(0)
}

// MockPortfolioRepository for testing
type MockPortfolioRepository struct {
	mock.Mock
}

func (m *MockPortfolioRepository) Create(portfolio *models.Portfolio) error {
	args := m.Called(portfolio)
	return args.Error(0)
}

func (m *MockPortfolioRepository) FindByID(id string) (*models.Portfolio, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioRepository) FindByUserID(userID string) ([]*models.Portfolio, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioRepository) FindByUserIDAndName(userID, name string) (*models.Portfolio, error) {
	args := m.Called(userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioRepository) Update(portfolio *models.Portfolio) error {
	args := m.Called(portfolio)
	return args.Error(0)
}

func (m *MockPortfolioRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPortfolioRepository) ExistsByUserIDAndName(userID, name string) (bool, error) {
	args := m.Called(userID, name)
	return args.Bool(0), args.Error(1)
}

// MockPerformanceSnapshotRepository for testing
type MockPerformanceSnapshotRepository struct {
	mock.Mock
}

func (m *MockPerformanceSnapshotRepository) Create(snapshot *models.PerformanceSnapshot) error {
	args := m.Called(snapshot)
	return args.Error(0)
}

func (m *MockPerformanceSnapshotRepository) FindByID(id string) (*models.PerformanceSnapshot, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PerformanceSnapshot), args.Error(1)
}

func (m *MockPerformanceSnapshotRepository) FindByPortfolioID(portfolioID string, limit, offset int) ([]*models.PerformanceSnapshot, error) {
	args := m.Called(portfolioID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PerformanceSnapshot), args.Error(1)
}

func (m *MockPerformanceSnapshotRepository) FindByPortfolioIDAndDateRange(portfolioID string, startDate, endDate time.Time) ([]*models.PerformanceSnapshot, error) {
	args := m.Called(portfolioID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PerformanceSnapshot), args.Error(1)
}

func (m *MockPerformanceSnapshotRepository) FindLatestByPortfolioID(portfolioID string) (*models.PerformanceSnapshot, error) {
	args := m.Called(portfolioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PerformanceSnapshot), args.Error(1)
}

func (m *MockPerformanceSnapshotRepository) FindByPortfolioIDAndDate(portfolioID string, date time.Time) (*models.PerformanceSnapshot, error) {
	args := m.Called(portfolioID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PerformanceSnapshot), args.Error(1)
}

func (m *MockPerformanceSnapshotRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPerformanceSnapshotRepository) DeleteByPortfolioID(portfolioID string) error {
	args := m.Called(portfolioID)
	return args.Error(0)
}

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

// MockTransactionRepository for testing
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) FindByID(id string) (*models.Transaction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) FindByPortfolioID(portfolioID string) ([]*models.Transaction, error) {
	args := m.Called(portfolioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) FindByPortfolioIDAndSymbol(portfolioID, symbol string) ([]*models.Transaction, error) {
	args := m.Called(portfolioID, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) FindByPortfolioIDWithFilters(portfolioID string, symbol *string, startDate, endDate *time.Time) ([]*models.Transaction, error) {
	args := m.Called(portfolioID, symbol, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) Update(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTransactionRepository) DeleteByImportBatchID(batchID string) error {
	args := m.Called(batchID)
	return args.Error(0)
}

// MockMarketDataService for testing
type MockMarketDataService struct {
	mock.Mock
}

func (m *MockMarketDataService) GetQuote(symbol string) (*Quote, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Quote), args.Error(1)
}

func (m *MockMarketDataService) GetQuotes(symbols []string) (map[string]*Quote, error) {
	args := m.Called(symbols)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*Quote), args.Error(1)
}

func (m *MockMarketDataService) GetHistoricalPrices(symbol string, startDate, endDate time.Time) ([]*HistoricalPrice, error) {
	args := m.Called(symbol, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*HistoricalPrice), args.Error(1)
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
