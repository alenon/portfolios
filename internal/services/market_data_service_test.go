package services

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewMarketDataService(t *testing.T) {
	mockProvider := new(MockMarketDataProvider)
	service := NewMarketDataService(mockProvider, 5*time.Minute)
	assert.NotNil(t, service)
}

func TestMarketDataService_GetQuote(t *testing.T) {
	mockProvider := new(MockMarketDataProvider)
	service := NewMarketDataService(mockProvider, 5*time.Minute)

	quote := &Quote{
		Symbol:        "AAPL",
		Price:         decimal.NewFromInt(150),
		Open:          decimal.NewFromInt(148),
		High:          decimal.NewFromInt(152),
		Low:           decimal.NewFromInt(147),
		Volume:        1000000,
		PreviousClose: decimal.NewFromInt(149),
		Change:        decimal.NewFromInt(1),
		ChangePercent: decimal.NewFromFloat(0.67),
		LastUpdated:   time.Now(),
	}

	t.Run("fetch and cache quote", func(t *testing.T) {
		mockProvider.On("GetQuote", mock.Anything, "AAPL").Return(quote, nil).Once()

		result, err := service.GetQuote("AAPL")
		assert.NoError(t, err)
		assert.Equal(t, "AAPL", result.Symbol)
		assert.True(t, result.Price.Equal(decimal.NewFromInt(150)))

		mockProvider.AssertExpectations(t)
	})

	t.Run("use cached quote", func(t *testing.T) {
		// Should not call provider again, use cache
		result, err := service.GetQuote("AAPL")
		assert.NoError(t, err)
		assert.Equal(t, "AAPL", result.Symbol)
		assert.True(t, result.Price.Equal(decimal.NewFromInt(150)))

		// No new expectations, should use cache
	})
}

func TestMarketDataService_GetQuotes(t *testing.T) {
	mockProvider := new(MockMarketDataProvider)
	service := NewMarketDataService(mockProvider, 5*time.Minute)

	quotes := map[string]*Quote{
		"AAPL": {
			Symbol: "AAPL",
			Price:  decimal.NewFromInt(150),
		},
		"GOOGL": {
			Symbol: "GOOGL",
			Price:  decimal.NewFromInt(2800),
		},
	}

	t.Run("fetch multiple quotes", func(t *testing.T) {
		mockProvider.On("GetQuotes", mock.Anything, []string{"AAPL", "GOOGL"}).Return(quotes, nil).Once()

		result, err := service.GetQuotes([]string{"AAPL", "GOOGL"})
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "AAPL", result["AAPL"].Symbol)
		assert.Equal(t, "GOOGL", result["GOOGL"].Symbol)

		mockProvider.AssertExpectations(t)
	})

	t.Run("use cache for some symbols", func(t *testing.T) {
		// AAPL and GOOGL are cached, only fetch MSFT
		newQuotes := map[string]*Quote{
			"MSFT": {
				Symbol: "MSFT",
				Price:  decimal.NewFromInt(300),
			},
		}
		mockProvider.On("GetQuotes", mock.Anything, []string{"MSFT"}).Return(newQuotes, nil).Once()

		result, err := service.GetQuotes([]string{"AAPL", "GOOGL", "MSFT"})
		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, "AAPL", result["AAPL"].Symbol)
		assert.Equal(t, "GOOGL", result["GOOGL"].Symbol)
		assert.Equal(t, "MSFT", result["MSFT"].Symbol)

		mockProvider.AssertExpectations(t)
	})
}

func TestMarketDataService_GetHistoricalPrices(t *testing.T) {
	mockProvider := new(MockMarketDataProvider)
	service := NewMarketDataService(mockProvider, 5*time.Minute)

	startDate := time.Now().Add(-30 * 24 * time.Hour)
	endDate := time.Now()

	prices := []*HistoricalPrice{
		{
			Date:   startDate,
			Open:   decimal.NewFromInt(145),
			High:   decimal.NewFromInt(148),
			Low:    decimal.NewFromInt(144),
			Close:  decimal.NewFromInt(147),
			Volume: 1000000,
		},
		{
			Date:   endDate,
			Open:   decimal.NewFromInt(147),
			High:   decimal.NewFromInt(150),
			Low:    decimal.NewFromInt(146),
			Close:  decimal.NewFromInt(149),
			Volume: 1100000,
		},
	}

	mockProvider.On("GetHistoricalPrices", mock.Anything, "AAPL", startDate, endDate).Return(prices, nil).Once()

	result, err := service.GetHistoricalPrices("AAPL", startDate, endDate)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.True(t, result[0].Close.Equal(decimal.NewFromInt(147)))
	assert.True(t, result[1].Close.Equal(decimal.NewFromInt(149)))

	mockProvider.AssertExpectations(t)
}

func TestMarketDataService_GetExchangeRate(t *testing.T) {
	mockProvider := new(MockMarketDataProvider)
	service := NewMarketDataService(mockProvider, 5*time.Minute)

	t.Run("same currency", func(t *testing.T) {
		rate, err := service.GetExchangeRate("USD", "USD")
		assert.NoError(t, err)
		assert.True(t, rate.Equal(decimal.NewFromInt(1)))
	})

	t.Run("different currencies", func(t *testing.T) {
		mockProvider.On("GetExchangeRate", mock.Anything, "EUR", "USD").Return(decimal.NewFromFloat(1.18), nil).Once()

		rate, err := service.GetExchangeRate("EUR", "USD")
		assert.NoError(t, err)
		assert.True(t, rate.Equal(decimal.NewFromFloat(1.18)))

		mockProvider.AssertExpectations(t)
	})
}

func TestMarketDataService_RefreshCache(t *testing.T) {
	mockProvider := new(MockMarketDataProvider)
	service := NewMarketDataService(mockProvider, 5*time.Minute)

	quote := &Quote{
		Symbol: "AAPL",
		Price:  decimal.NewFromInt(150),
	}

	// First call to cache
	mockProvider.On("GetQuote", mock.Anything, "AAPL").Return(quote, nil).Once()
	_, err := service.GetQuote("AAPL")
	assert.NoError(t, err)

	// Refresh cache - should fetch again
	updatedQuote := &Quote{
		Symbol: "AAPL",
		Price:  decimal.NewFromInt(155),
	}
	mockProvider.On("GetQuote", mock.Anything, "AAPL").Return(updatedQuote, nil).Once()

	err = service.RefreshCache("AAPL")
	assert.NoError(t, err)

	mockProvider.AssertExpectations(t)
}

// MockMarketDataProvider for testing
type MockMarketDataProvider struct {
	mock.Mock
}

func (m *MockMarketDataProvider) GetQuote(ctx context.Context, symbol string) (*Quote, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Quote), args.Error(1)
}

func (m *MockMarketDataProvider) GetQuotes(ctx context.Context, symbols []string) (map[string]*Quote, error) {
	args := m.Called(ctx, symbols)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*Quote), args.Error(1)
}

func (m *MockMarketDataProvider) GetHistoricalPrices(ctx context.Context, symbol string, startDate, endDate time.Time) ([]*HistoricalPrice, error) {
	args := m.Called(ctx, symbol, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*HistoricalPrice), args.Error(1)
}

func (m *MockMarketDataProvider) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error) {
	args := m.Called(ctx, fromCurrency, toCurrency)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockMarketDataProvider) IsAvailable() bool {
	args := m.Called()
	return args.Bool(0)
}
