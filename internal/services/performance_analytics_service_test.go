package services

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewPerformanceAnalyticsService(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	)

	assert.NotNil(t, svc)
}

func TestCalculateTWR_Success(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: uuid.MustParse(userID),
		Name:   "Test Portfolio",
	}

	snapshots := []*models.PerformanceSnapshot{
		{
			PortfolioID:    uuid.MustParse(portfolioID),
			Date:           startDate,
			TotalValue:     decimal.NewFromInt(10000),
			TotalCostBasis: decimal.NewFromInt(9000),
			TotalReturn:    decimal.NewFromInt(1000),
			TotalReturnPct: decimal.NewFromFloat(11.11),
		},
		{
			PortfolioID:    uuid.MustParse(portfolioID),
			Date:           startDate.AddDate(0, 6, 0),
			TotalValue:     decimal.NewFromInt(11000),
			TotalCostBasis: decimal.NewFromInt(9500),
			TotalReturn:    decimal.NewFromInt(1500),
			TotalReturnPct: decimal.NewFromFloat(15.79),
		},
		{
			PortfolioID:    uuid.MustParse(portfolioID),
			Date:           endDate,
			TotalValue:     decimal.NewFromInt(12000),
			TotalCostBasis: decimal.NewFromInt(10000),
			TotalReturn:    decimal.NewFromInt(2000),
			TotalReturnPct: decimal.NewFromFloat(20.00),
		},
	}

	transactions := []*models.Transaction{}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)
	snapshotRepo.On("FindByPortfolioIDAndDateRange", portfolioID, startDate, endDate).Return(snapshots, nil)
	transactionRepo.On("FindByPortfolioIDWithFilters", portfolioID, mock.Anything, &startDate, &endDate).Return(transactions, nil)

	result, err := svc.CalculateTWR(portfolioID, userID, startDate, endDate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, startDate, result.StartDate)
	assert.Equal(t, endDate, result.EndDate)
	assert.Equal(t, decimal.NewFromInt(10000), result.StartingValue)
	assert.Equal(t, decimal.NewFromInt(12000), result.EndingValue)
	assert.Equal(t, 2, result.NumPeriods)

	portfolioRepo.AssertExpectations(t)
	snapshotRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

func TestCalculateTWR_UnauthorizedAccess(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()
	differentUserID := uuid.New().String()
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: uuid.MustParse(differentUserID),
		Name:   "Test Portfolio",
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)

	result, err := svc.CalculateTWR(portfolioID, userID, startDate, endDate)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)

	portfolioRepo.AssertExpectations(t)
}

func TestCalculateTWR_PortfolioNotFound(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	portfolioRepo.On("FindByID", portfolioID).Return(nil, models.ErrPortfolioNotFound)

	result, err := svc.CalculateTWR(portfolioID, userID, startDate, endDate)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, models.ErrPortfolioNotFound, err)

	portfolioRepo.AssertExpectations(t)
}

func TestCalculateTWR_InsufficientSnapshots(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: uuid.MustParse(userID),
		Name:   "Test Portfolio",
	}

	snapshots := []*models.PerformanceSnapshot{
		{
			PortfolioID: uuid.MustParse(portfolioID),
			Date:        startDate,
			TotalValue:  decimal.NewFromInt(10000),
		},
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)
	snapshotRepo.On("FindByPortfolioIDAndDateRange", portfolioID, startDate, endDate).Return(snapshots, nil)

	result, err := svc.CalculateTWR(portfolioID, userID, startDate, endDate)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "insufficient data")

	portfolioRepo.AssertExpectations(t)
	snapshotRepo.AssertExpectations(t)
}

func TestCalculateMWR_Success(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: uuid.MustParse(userID),
		Name:   "Test Portfolio",
	}

	startSnapshot := &models.PerformanceSnapshot{
		PortfolioID: uuid.MustParse(portfolioID),
		Date:        startDate,
		TotalValue:  decimal.NewFromInt(10000),
	}

	endSnapshot := &models.PerformanceSnapshot{
		PortfolioID: uuid.MustParse(portfolioID),
		Date:        endDate,
		TotalValue:  decimal.NewFromInt(12000),
	}

	price := decimal.NewFromInt(100)
	transactions := []*models.Transaction{
		{
			PortfolioID: uuid.MustParse(portfolioID),
			Symbol:      "AAPL",
			Type:        models.TransactionTypeBuy,
			Quantity:    decimal.NewFromInt(10),
			Price:       &price,
			Date:        time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)
	snapshotRepo.On("FindByPortfolioIDAndDate", portfolioID, startDate).Return(startSnapshot, nil)
	snapshotRepo.On("FindByPortfolioIDAndDate", portfolioID, endDate).Return(endSnapshot, nil)
	transactionRepo.On("FindByPortfolioIDWithFilters", portfolioID, mock.Anything, &startDate, &endDate).Return(transactions, nil)

	result, err := svc.CalculateMWR(portfolioID, userID, startDate, endDate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, startDate, result.StartDate)
	assert.Equal(t, endDate, result.EndDate)
	assert.Equal(t, decimal.NewFromInt(10000), result.StartingValue)
	assert.Equal(t, decimal.NewFromInt(12000), result.EndingValue)

	portfolioRepo.AssertExpectations(t)
	snapshotRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

func TestCalculateMWR_UnauthorizedAccess(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()
	differentUserID := uuid.New().String()
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: uuid.MustParse(differentUserID),
		Name:   "Test Portfolio",
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)

	result, err := svc.CalculateMWR(portfolioID, userID, startDate, endDate)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)

	portfolioRepo.AssertExpectations(t)
}

func TestCalculateMWR_NoStartSnapshot(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: uuid.MustParse(userID),
		Name:   "Test Portfolio",
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)
	transactionRepo.On("FindByPortfolioIDWithFilters", portfolioID, mock.Anything, &startDate, &endDate).Return([]*models.Transaction{}, nil)
	snapshotRepo.On("FindByPortfolioIDAndDate", portfolioID, startDate).Return(nil, errors.New("snapshot not found"))
	snapshotRepo.On("FindByPortfolioIDAndDateRange", portfolioID, mock.Anything, mock.Anything).Return([]*models.PerformanceSnapshot{}, nil)

	result, err := svc.CalculateMWR(portfolioID, userID, startDate, endDate)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "starting snapshot")

	portfolioRepo.AssertExpectations(t)
	snapshotRepo.AssertExpectations(t)
}

func TestCalculateAnnualizedReturn_Success(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: uuid.MustParse(userID),
		Name:   "Test Portfolio",
	}

	startSnapshot := &models.PerformanceSnapshot{
		PortfolioID: uuid.MustParse(portfolioID),
		Date:        startDate,
		TotalValue:  decimal.NewFromInt(10000),
	}

	endSnapshot := &models.PerformanceSnapshot{
		PortfolioID: uuid.MustParse(portfolioID),
		Date:        endDate,
		TotalValue:  decimal.NewFromInt(12000),
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)
	snapshotRepo.On("FindByPortfolioIDAndDate", portfolioID, startDate).Return(startSnapshot, nil)
	snapshotRepo.On("FindByPortfolioIDAndDate", portfolioID, endDate).Return(endSnapshot, nil)

	result, err := svc.CalculateAnnualizedReturn(portfolioID, userID, startDate, endDate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, startDate, result.StartDate)
	assert.Equal(t, endDate, result.EndDate)
	assert.True(t, result.TotalReturn.Equal(decimal.NewFromInt(2000)))
	assert.True(t, result.TotalReturnPct.GreaterThan(decimal.Zero))

	portfolioRepo.AssertExpectations(t)
	snapshotRepo.AssertExpectations(t)
}

func TestCalculateAnnualizedReturn_ZeroStartingValue(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: uuid.MustParse(userID),
		Name:   "Test Portfolio",
	}

	startSnapshot := &models.PerformanceSnapshot{
		PortfolioID: uuid.MustParse(portfolioID),
		Date:        startDate,
		TotalValue:  decimal.Zero,
	}

	endSnapshot := &models.PerformanceSnapshot{
		PortfolioID: uuid.MustParse(portfolioID),
		Date:        endDate,
		TotalValue:  decimal.NewFromInt(12000),
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)
	snapshotRepo.On("FindByPortfolioIDAndDate", portfolioID, startDate).Return(startSnapshot, nil)
	snapshotRepo.On("FindByPortfolioIDAndDate", portfolioID, endDate).Return(endSnapshot, nil)

	result, err := svc.CalculateAnnualizedReturn(portfolioID, userID, startDate, endDate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.TotalReturnPct.Equal(decimal.Zero))
	assert.True(t, result.AnnualizedReturn.Equal(decimal.Zero))

	portfolioRepo.AssertExpectations(t)
	snapshotRepo.AssertExpectations(t)
}

func TestCompareToBenchmark_Success(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	benchmarkSymbol := "SPY"

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: uuid.MustParse(userID),
		Name:   "Test Portfolio",
	}

	startSnapshot := &models.PerformanceSnapshot{
		PortfolioID: uuid.MustParse(portfolioID),
		Date:        startDate,
		TotalValue:  decimal.NewFromInt(10000),
	}

	endSnapshot := &models.PerformanceSnapshot{
		PortfolioID: uuid.MustParse(portfolioID),
		Date:        endDate,
		TotalValue:  decimal.NewFromInt(12000),
	}

	benchmarkPrices := []*HistoricalPrice{
		{
			Date:  startDate,
			Close: decimal.NewFromInt(400),
		},
		{
			Date:  endDate,
			Close: decimal.NewFromInt(440),
		},
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil).Times(2)
	snapshotRepo.On("FindByPortfolioIDAndDate", portfolioID, startDate).Return(startSnapshot, nil)
	snapshotRepo.On("FindByPortfolioIDAndDate", portfolioID, endDate).Return(endSnapshot, nil)
	marketDataSvc.On("GetHistoricalPrices", benchmarkSymbol, startDate, endDate).Return(benchmarkPrices, nil)

	result, err := svc.CompareToBenchmark(portfolioID, userID, benchmarkSymbol, startDate, endDate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, benchmarkSymbol, result.BenchmarkSymbol)
	assert.True(t, result.PortfolioReturn.GreaterThan(decimal.Zero))
	assert.True(t, result.BenchmarkReturn.GreaterThan(decimal.Zero))

	portfolioRepo.AssertExpectations(t)
	snapshotRepo.AssertExpectations(t)
	marketDataSvc.AssertExpectations(t)
}

func TestCompareToBenchmark_InsufficientBenchmarkData(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	benchmarkSymbol := "SPY"

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: uuid.MustParse(userID),
		Name:   "Test Portfolio",
	}

	startSnapshot := &models.PerformanceSnapshot{
		PortfolioID: uuid.MustParse(portfolioID),
		Date:        startDate,
		TotalValue:  decimal.NewFromInt(10000),
	}

	endSnapshot := &models.PerformanceSnapshot{
		PortfolioID: uuid.MustParse(portfolioID),
		Date:        endDate,
		TotalValue:  decimal.NewFromInt(12000),
	}

	benchmarkPrices := []*HistoricalPrice{
		{
			Date:  startDate,
			Close: decimal.NewFromInt(400),
		},
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil).Times(2)
	snapshotRepo.On("FindByPortfolioIDAndDate", portfolioID, startDate).Return(startSnapshot, nil)
	snapshotRepo.On("FindByPortfolioIDAndDate", portfolioID, endDate).Return(endSnapshot, nil)
	marketDataSvc.On("GetHistoricalPrices", benchmarkSymbol, startDate, endDate).Return(benchmarkPrices, nil)

	result, err := svc.CompareToBenchmark(portfolioID, userID, benchmarkSymbol, startDate, endDate)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "insufficient benchmark data")

	portfolioRepo.AssertExpectations(t)
	snapshotRepo.AssertExpectations(t)
	marketDataSvc.AssertExpectations(t)
}

func TestGetPerformanceMetrics_Success(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: uuid.MustParse(userID),
		Name:   "Test Portfolio",
	}

	startSnapshot := &models.PerformanceSnapshot{
		PortfolioID: uuid.MustParse(portfolioID),
		Date:        startDate,
		TotalValue:  decimal.NewFromInt(10000),
	}

	endSnapshot := &models.PerformanceSnapshot{
		PortfolioID: uuid.MustParse(portfolioID),
		Date:        endDate,
		TotalValue:  decimal.NewFromInt(12000),
	}

	snapshots := []*models.PerformanceSnapshot{
		startSnapshot,
		{
			PortfolioID:    uuid.MustParse(portfolioID),
			Date:           startDate.AddDate(0, 6, 0),
			TotalValue:     decimal.NewFromInt(11000),
			TotalCostBasis: decimal.NewFromInt(9500),
			TotalReturn:    decimal.NewFromInt(1500),
			TotalReturnPct: decimal.NewFromFloat(15.79),
		},
		endSnapshot,
	}

	price := decimal.NewFromInt(100)
	transactions := []*models.Transaction{
		{
			PortfolioID: uuid.MustParse(portfolioID),
			Symbol:      "AAPL",
			Type:        models.TransactionTypeBuy,
			Quantity:    decimal.NewFromInt(10),
			Price:       &price,
			Date:        time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	// GetPerformanceMetrics calls:
	// - verifyPortfolioOwnership: 1 time
	// - getSnapshotNearDate(start): 1 time
	// - getSnapshotNearDate(end): 1 time
	// - CalculateTWR: verifyPortfolioOwnership + FindByPortfolioIDAndDateRange + FindByPortfolioIDWithFilters
	// - CalculateMWR: verifyPortfolioOwnership + 2x getSnapshotNearDate + FindByPortfolioIDWithFilters
	// - FindByPortfolioIDWithFilters: 1 time
	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil).Times(3)
	snapshotRepo.On("FindByPortfolioIDAndDate", portfolioID, startDate).Return(startSnapshot, nil).Times(2)
	snapshotRepo.On("FindByPortfolioIDAndDate", portfolioID, endDate).Return(endSnapshot, nil).Times(2)
	snapshotRepo.On("FindByPortfolioIDAndDateRange", portfolioID, startDate, endDate).Return(snapshots, nil).Times(1)
	transactionRepo.On("FindByPortfolioIDWithFilters", portfolioID, mock.Anything, &startDate, &endDate).Return(transactions, nil).Times(3)

	result, err := svc.GetPerformanceMetrics(portfolioID, userID, startDate, endDate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, startDate, result.StartDate)
	assert.Equal(t, endDate, result.EndDate)
	assert.Equal(t, decimal.NewFromInt(10000), result.StartingValue)
	assert.Equal(t, decimal.NewFromInt(12000), result.EndingValue)
	assert.True(t, result.TotalDeposits.GreaterThan(decimal.Zero))

	portfolioRepo.AssertExpectations(t)
	snapshotRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

func TestGetPerformanceMetrics_UnauthorizedAccess(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()
	differentUserID := uuid.New().String()
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: uuid.MustParse(differentUserID),
		Name:   "Test Portfolio",
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)

	result, err := svc.GetPerformanceMetrics(portfolioID, userID, startDate, endDate)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)

	portfolioRepo.AssertExpectations(t)
}

func TestGetSnapshotNearDate_ExactMatch(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	).(*performanceAnalyticsService)

	portfolioID := uuid.New().String()
	date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	snapshot := &models.PerformanceSnapshot{
		PortfolioID: uuid.MustParse(portfolioID),
		Date:        date,
		TotalValue:  decimal.NewFromInt(10000),
	}

	snapshotRepo.On("FindByPortfolioIDAndDate", portfolioID, date).Return(snapshot, nil)

	result, err := svc.getSnapshotNearDate(portfolioID, date)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, date, result.Date)

	snapshotRepo.AssertExpectations(t)
}

func TestGetSnapshotNearDate_ClosestMatch(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	).(*performanceAnalyticsService)

	portfolioID := uuid.New().String()
	targetDate := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

	snapshots := []*models.PerformanceSnapshot{
		{
			PortfolioID: uuid.MustParse(portfolioID),
			Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			TotalValue:  decimal.NewFromInt(10000),
		},
		{
			PortfolioID: uuid.MustParse(portfolioID),
			Date:        time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC),
			TotalValue:  decimal.NewFromInt(11000),
		},
	}

	snapshotRepo.On("FindByPortfolioIDAndDate", portfolioID, targetDate).Return(nil, errors.New("not found"))
	snapshotRepo.On("FindByPortfolioIDAndDateRange", portfolioID, mock.Anything, mock.Anything).Return(snapshots, nil)

	result, err := svc.getSnapshotNearDate(portfolioID, targetDate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), result.Date)

	snapshotRepo.AssertExpectations(t)
}

func TestCalculateCashFlowBetweenDates(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	).(*performanceAnalyticsService)

	portfolioID := uuid.MustParse(uuid.New().String())
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	price1 := decimal.NewFromInt(100)
	price2 := decimal.NewFromInt(110)
	price3 := decimal.NewFromInt(150)
	transactions := []*models.Transaction{
		{
			PortfolioID: portfolioID,
			Symbol:      "AAPL",
			Type:        models.TransactionTypeBuy,
			Quantity:    decimal.NewFromInt(10),
			Price:       &price1,
			Date:        time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			PortfolioID: portfolioID,
			Symbol:      "AAPL",
			Type:        models.TransactionTypeSell,
			Quantity:    decimal.NewFromInt(5),
			Price:       &price2,
			Date:        time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			PortfolioID: portfolioID,
			Symbol:      "GOOGL",
			Type:        models.TransactionTypeBuy,
			Quantity:    decimal.NewFromInt(2),
			Price:       &price3,
			Date:        time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC), // Before startDate
		},
	}

	cashFlow := svc.calculateCashFlowBetweenDates(transactions, startDate, endDate)

	// Buy transaction: 10 * 100 = 1000
	// Sell transaction: -(5 * 110) = -550
	// Transaction before startDate should be excluded
	// Net cash flow: 1000 - 550 = 450
	assert.True(t, cashFlow.Equal(decimal.NewFromInt(450)))
}

func TestBuildCashFlowSeries(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	).(*performanceAnalyticsService)

	portfolioID := uuid.MustParse(uuid.New().String())
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	startingValue := decimal.NewFromInt(10000)
	endingValue := decimal.NewFromInt(12000)

	price := decimal.NewFromInt(100)
	transactions := []*models.Transaction{
		{
			PortfolioID: portfolioID,
			Symbol:      "AAPL",
			Type:        models.TransactionTypeBuy,
			Quantity:    decimal.NewFromInt(10),
			Price:       &price,
			Date:        time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	cashFlows := svc.buildCashFlowSeries(transactions, startDate, endDate, startingValue, endingValue)

	assert.Len(t, cashFlows, 3) // Start, 1 transaction, end
	assert.Equal(t, startDate, cashFlows[0].Date)
	assert.True(t, cashFlows[0].Amount.Equal(startingValue.Neg()))
	assert.Equal(t, endDate, cashFlows[len(cashFlows)-1].Date)
	assert.True(t, cashFlows[len(cashFlows)-1].Amount.Equal(endingValue))
}

func TestCalculateIRR(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	).(*performanceAnalyticsService)

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	cashFlows := []CashFlow{
		{Date: startDate, Amount: decimal.NewFromInt(-10000)},
		{Date: endDate, Amount: decimal.NewFromInt(12000)},
	}

	irr := svc.calculateIRR(cashFlows)

	assert.True(t, irr.GreaterThan(decimal.Zero))
	assert.True(t, irr.LessThan(decimal.NewFromInt(1)))
}

func TestCalculateIRR_InsufficientCashFlows(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	).(*performanceAnalyticsService)

	cashFlows := []CashFlow{
		{Date: time.Now(), Amount: decimal.NewFromInt(-10000)},
	}

	irr := svc.calculateIRR(cashFlows)

	assert.True(t, irr.Equal(decimal.Zero))
}

func TestVerifyPortfolioOwnership_Success(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	).(*performanceAnalyticsService)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: uuid.MustParse(userID),
		Name:   "Test Portfolio",
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)

	err := svc.verifyPortfolioOwnership(portfolioID, userID)

	assert.NoError(t, err)
	portfolioRepo.AssertExpectations(t)
}

func TestVerifyPortfolioOwnership_NotFound(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	).(*performanceAnalyticsService)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()

	portfolioRepo.On("FindByID", portfolioID).Return(nil, models.ErrPortfolioNotFound)

	err := svc.verifyPortfolioOwnership(portfolioID, userID)

	assert.Error(t, err)
	assert.Equal(t, models.ErrPortfolioNotFound, err)
	portfolioRepo.AssertExpectations(t)
}

func TestVerifyPortfolioOwnership_Unauthorized(t *testing.T) {
	portfolioRepo := new(MockPortfolioRepository)
	transactionRepo := new(MockTransactionRepository)
	snapshotRepo := new(MockPerformanceSnapshotRepository)
	marketDataSvc := new(MockMarketDataService)

	svc := NewPerformanceAnalyticsService(
		portfolioRepo,
		transactionRepo,
		snapshotRepo,
		marketDataSvc,
	).(*performanceAnalyticsService)

	portfolioID := uuid.New().String()
	userID := uuid.New().String()
	differentUserID := uuid.New().String()

	portfolio := &models.Portfolio{
		ID:     uuid.MustParse(portfolioID),
		UserID: uuid.MustParse(differentUserID),
		Name:   "Test Portfolio",
	}

	portfolioRepo.On("FindByID", portfolioID).Return(portfolio, nil)

	err := svc.verifyPortfolioOwnership(portfolioID, userID)

	assert.Error(t, err)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)
	portfolioRepo.AssertExpectations(t)
}
