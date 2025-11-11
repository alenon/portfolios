package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewHoldingService(t *testing.T) {
	mockHoldingRepo := new(MockHoldingRepository)
	mockPortfolioRepo := new(MockPortfolioRepository)

	service := NewHoldingService(mockHoldingRepo, mockPortfolioRepo)
	assert.NotNil(t, service)
}

func TestHoldingService_GetByPortfolioID(t *testing.T) {
	mockHoldingRepo := new(MockHoldingRepository)
	mockPortfolioRepo := new(MockPortfolioRepository)
	service := NewHoldingService(mockHoldingRepo, mockPortfolioRepo)

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

	t.Run("successful retrieval", func(t *testing.T) {
		mockPortfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil).Once()
		mockHoldingRepo.On("FindByPortfolioID", portfolioID.String()).Return(holdings, nil).Once()

		result, err := service.GetByPortfolioID(portfolioID.String(), userID.String())
		assert.NoError(t, err)
		assert.Len(t, result, 1)

		mockPortfolioRepo.AssertExpectations(t)
		mockHoldingRepo.AssertExpectations(t)
	})

	t.Run("portfolio not found", func(t *testing.T) {
		mockPortfolioRepo.On("FindByID", portfolioID.String()).Return(nil, models.ErrPortfolioNotFound).Once()

		_, err := service.GetByPortfolioID(portfolioID.String(), userID.String())
		assert.Equal(t, models.ErrPortfolioNotFound, err)

		mockPortfolioRepo.AssertExpectations(t)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		differentUserID := uuid.New()
		mockPortfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil).Once()

		_, err := service.GetByPortfolioID(portfolioID.String(), differentUserID.String())
		assert.Equal(t, models.ErrUnauthorizedAccess, err)

		mockPortfolioRepo.AssertExpectations(t)
	})
}

func TestHoldingService_GetByPortfolioIDAndSymbol(t *testing.T) {
	mockHoldingRepo := new(MockHoldingRepository)
	mockPortfolioRepo := new(MockPortfolioRepository)
	service := NewHoldingService(mockHoldingRepo, mockPortfolioRepo)

	portfolioID := uuid.New()
	userID := uuid.New()

	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          userID,
		Name:            "Test Portfolio",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
	}

	holding := &models.Holding{
		ID:           uuid.New(),
		PortfolioID:  portfolioID,
		Symbol:       "AAPL",
		Quantity:     decimal.NewFromInt(100),
		CostBasis:    decimal.NewFromInt(15000),
		AvgCostPrice: decimal.NewFromInt(150),
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockPortfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil).Once()
		mockHoldingRepo.On("FindByPortfolioIDAndSymbol", portfolioID.String(), "AAPL").Return(holding, nil).Once()

		result, err := service.GetByPortfolioIDAndSymbol(portfolioID.String(), "AAPL", userID.String())
		assert.NoError(t, err)
		assert.Equal(t, "AAPL", result.Symbol)

		mockPortfolioRepo.AssertExpectations(t)
		mockHoldingRepo.AssertExpectations(t)
	})
}

func TestHoldingService_GetPortfolioValue(t *testing.T) {
	mockHoldingRepo := new(MockHoldingRepository)
	mockPortfolioRepo := new(MockPortfolioRepository)
	service := NewHoldingService(mockHoldingRepo, mockPortfolioRepo)

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
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Symbol:       "GOOGL",
			Quantity:     decimal.NewFromInt(50),
			CostBasis:    decimal.NewFromInt(10000),
			AvgCostPrice: decimal.NewFromInt(200),
		},
	}

	t.Run("with all prices", func(t *testing.T) {
		mockPortfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil).Once()
		mockHoldingRepo.On("FindByPortfolioID", portfolioID.String()).Return(holdings, nil).Once()

		prices := map[string]decimal.Decimal{
			"AAPL":  decimal.NewFromInt(180),
			"GOOGL": decimal.NewFromInt(220),
		}

		totalValue, err := service.GetPortfolioValue(portfolioID.String(), userID.String(), prices)
		assert.NoError(t, err)
		assert.True(t, totalValue.Equal(decimal.NewFromInt(29000)))

		mockPortfolioRepo.AssertExpectations(t)
		mockHoldingRepo.AssertExpectations(t)
	})
}
