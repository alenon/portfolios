package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewPerformanceSnapshotService(t *testing.T) {
	mockSnapshotRepo := new(MockPerformanceSnapshotRepository)
	mockPortfolioRepo := new(MockPortfolioRepository)
	mockHoldingRepo := new(MockHoldingRepository)

	service := NewPerformanceSnapshotService(mockSnapshotRepo, mockPortfolioRepo, mockHoldingRepo)
	assert.NotNil(t, service)
}

func TestPerformanceSnapshotService_CreateSnapshot(t *testing.T) {
	mockSnapshotRepo := new(MockPerformanceSnapshotRepository)
	mockPortfolioRepo := new(MockPortfolioRepository)
	mockHoldingRepo := new(MockHoldingRepository)
	service := NewPerformanceSnapshotService(mockSnapshotRepo, mockPortfolioRepo, mockHoldingRepo)

	portfolioID := uuid.New()
	userID := uuid.New()

	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          userID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}

	holdings := []*models.Holding{
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Symbol:       "AAPL",
			Quantity:     decimal.NewFromInt(100),
			CostBasis:    decimal.NewFromInt(15000),
			AvgCostPrice: decimal.NewFromInt(150),
		},
	}

	prices := map[string]decimal.Decimal{
		"AAPL": decimal.NewFromInt(180),
	}

	previousSnapshot := &models.PerformanceSnapshot{
		ID:             uuid.New(),
		PortfolioID:    portfolioID,
		Date:           time.Now().Add(-24 * time.Hour),
		TotalValue:     decimal.NewFromInt(17000),
		TotalCostBasis: decimal.NewFromInt(15000),
	}

	t.Run("successful snapshot creation", func(t *testing.T) {
		mockPortfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil).Once()
		mockHoldingRepo.On("FindByPortfolioID", portfolioID.String()).Return(holdings, nil).Once()
		mockSnapshotRepo.On("FindLatestByPortfolioID", portfolioID.String()).Return(previousSnapshot, nil).Once()
		mockSnapshotRepo.On("Create", mock.AnythingOfType("*models.PerformanceSnapshot")).Return(nil).Once()

		snapshot, err := service.CreateSnapshot(portfolioID.String(), userID.String(), prices)
		assert.NoError(t, err)
		assert.NotNil(t, snapshot)
		assert.True(t, snapshot.TotalValue.Equal(decimal.NewFromInt(18000)))

		mockPortfolioRepo.AssertExpectations(t)
		mockHoldingRepo.AssertExpectations(t)
		mockSnapshotRepo.AssertExpectations(t)
	})

	t.Run("portfolio not found", func(t *testing.T) {
		mockPortfolioRepo.On("FindByID", portfolioID.String()).Return(nil, models.ErrPortfolioNotFound).Once()

		_, err := service.CreateSnapshot(portfolioID.String(), userID.String(), prices)
		assert.Equal(t, models.ErrPortfolioNotFound, err)

		mockPortfolioRepo.AssertExpectations(t)
	})
}

func TestPerformanceSnapshotService_GetByPortfolioID(t *testing.T) {
	mockSnapshotRepo := new(MockPerformanceSnapshotRepository)
	mockPortfolioRepo := new(MockPortfolioRepository)
	mockHoldingRepo := new(MockHoldingRepository)
	service := NewPerformanceSnapshotService(mockSnapshotRepo, mockPortfolioRepo, mockHoldingRepo)

	portfolioID := uuid.New()
	userID := uuid.New()

	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          userID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}

	snapshots := []*models.PerformanceSnapshot{
		{
			ID:             uuid.New(),
			PortfolioID:    portfolioID,
			Date:           time.Now(),
			TotalValue:     decimal.NewFromInt(30000),
			TotalCostBasis: decimal.NewFromInt(25000),
		},
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockPortfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil).Once()
		mockSnapshotRepo.On("FindByPortfolioID", portfolioID.String(), 10, 0).Return(snapshots, nil).Once()

		result, err := service.GetByPortfolioID(portfolioID.String(), userID.String(), 10, 0)
		assert.NoError(t, err)
		assert.Len(t, result, 1)

		mockPortfolioRepo.AssertExpectations(t)
		mockSnapshotRepo.AssertExpectations(t)
	})
}

func TestPerformanceSnapshotService_GetLatest(t *testing.T) {
	mockSnapshotRepo := new(MockPerformanceSnapshotRepository)
	mockPortfolioRepo := new(MockPortfolioRepository)
	mockHoldingRepo := new(MockHoldingRepository)
	service := NewPerformanceSnapshotService(mockSnapshotRepo, mockPortfolioRepo, mockHoldingRepo)

	portfolioID := uuid.New()
	userID := uuid.New()

	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          userID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}

	snapshot := &models.PerformanceSnapshot{
		ID:             uuid.New(),
		PortfolioID:    portfolioID,
		Date:           time.Now(),
		TotalValue:     decimal.NewFromInt(30000),
		TotalCostBasis: decimal.NewFromInt(25000),
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockPortfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil).Once()
		mockSnapshotRepo.On("FindLatestByPortfolioID", portfolioID.String()).Return(snapshot, nil).Once()

		result, err := service.GetLatest(portfolioID.String(), userID.String())
		assert.NoError(t, err)
		assert.NotNil(t, result)

		mockPortfolioRepo.AssertExpectations(t)
		mockSnapshotRepo.AssertExpectations(t)
	})
}
