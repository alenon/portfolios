package dto

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/services"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// Test Portfolio DTOs
func TestToPortfolioResponse(t *testing.T) {
	portfolioID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          userID,
		Name:            "Test Portfolio",
		Description:     "Test Description",
		BaseCurrency:    "USD",
		CostBasisMethod: models.CostBasisFIFO,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	response := ToPortfolioResponse(portfolio)

	assert.NotNil(t, response)
	assert.Equal(t, portfolioID, response.ID)
	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, "Test Portfolio", response.Name)
	assert.Equal(t, "Test Description", response.Description)
	assert.Equal(t, "USD", response.BaseCurrency)
	assert.Equal(t, models.CostBasisFIFO, response.CostBasisMethod)
}

func TestToPortfolioResponse_Nil(t *testing.T) {
	response := ToPortfolioResponse(nil)
	assert.Nil(t, response)
}

func TestToPortfolioListResponse(t *testing.T) {
	portfolios := []*models.Portfolio{
		{
			ID:     uuid.New(),
			UserID: uuid.New(),
			Name:   "Portfolio 1",
		},
		{
			ID:     uuid.New(),
			UserID: uuid.New(),
			Name:   "Portfolio 2",
		},
	}

	response := ToPortfolioListResponse(portfolios)

	assert.NotNil(t, response)
	assert.Equal(t, 2, response.Total)
	assert.Len(t, response.Portfolios, 2)
}

// Test Transaction DTOs
func TestToTransactionResponse(t *testing.T) {
	transactionID := uuid.New()
	portfolioID := uuid.New()
	price := decimal.NewFromFloat(150.25)
	now := time.Now()

	transaction := &models.Transaction{
		ID:          transactionID,
		PortfolioID: portfolioID,
		Type:        models.TransactionTypeBuy,
		Symbol:      "AAPL",
		Date:        now,
		Quantity:    decimal.NewFromInt(10),
		Price:       &price,
		Commission:  decimal.NewFromFloat(9.99),
		Currency:    "USD",
		Notes:       "Test transaction",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	response := ToTransactionResponse(transaction)

	assert.NotNil(t, response)
	assert.Equal(t, transactionID, response.ID)
	assert.Equal(t, portfolioID, response.PortfolioID)
	assert.Equal(t, models.TransactionTypeBuy, response.Type)
	assert.Equal(t, "AAPL", response.Symbol)
	assert.Equal(t, decimal.NewFromInt(10), response.Quantity)
}

func TestToTransactionResponse_Nil(t *testing.T) {
	response := ToTransactionResponse(nil)
	assert.Nil(t, response)
}

func TestToTransactionListResponse(t *testing.T) {
	price := decimal.NewFromFloat(150.25)
	transactions := []*models.Transaction{
		{
			ID:       uuid.New(),
			Symbol:   "AAPL",
			Quantity: decimal.NewFromInt(10),
			Price:    &price,
		},
		{
			ID:       uuid.New(),
			Symbol:   "GOOGL",
			Quantity: decimal.NewFromInt(5),
			Price:    &price,
		},
	}

	response := ToTransactionListResponse(transactions)

	assert.NotNil(t, response)
	assert.Equal(t, 2, response.Total)
	assert.Len(t, response.Transactions, 2)
}

// Test Holding DTOs
func TestToHoldingResponse(t *testing.T) {
	holdingID := uuid.New()
	portfolioID := uuid.New()

	holding := &models.Holding{
		ID:          holdingID,
		PortfolioID: portfolioID,
		Symbol:      "AAPL",
		Quantity:    decimal.NewFromInt(100),
		CostBasis:   decimal.NewFromInt(10000),
	}

	response := ToHoldingResponse(holding)

	assert.NotNil(t, response)
	assert.Equal(t, holdingID, response.ID)
	assert.Equal(t, portfolioID, response.PortfolioID)
	assert.Equal(t, "AAPL", response.Symbol)
	assert.Equal(t, decimal.NewFromInt(100), response.Quantity)
	assert.Equal(t, decimal.NewFromInt(10000), response.CostBasis)
}

func TestToHoldingResponseWithMarketData(t *testing.T) {
	holdingID := uuid.New()
	portfolioID := uuid.New()

	holding := &models.Holding{
		ID:          holdingID,
		PortfolioID: portfolioID,
		Symbol:      "AAPL",
		Quantity:    decimal.NewFromInt(100),
		CostBasis:   decimal.NewFromInt(10000),
	}

	currentPrice := decimal.NewFromFloat(150.25)
	totalPortfolioValue := decimal.NewFromInt(20000)

	response := ToHoldingResponseWithMarketData(holding, currentPrice, totalPortfolioValue)

	assert.NotNil(t, response)
	assert.Equal(t, holdingID, response.ID)
	assert.Equal(t, currentPrice, *response.MarketPrice)
	assert.NotNil(t, response.MarketValue)
	assert.NotNil(t, response.UnrealizedGain)
	assert.NotNil(t, response.UnrealizedGainPct)
}

func TestToHoldingListResponse(t *testing.T) {
	holdings := []*models.Holding{
		{
			ID:       uuid.New(),
			Symbol:   "AAPL",
			Quantity: decimal.NewFromInt(100),
		},
		{
			ID:       uuid.New(),
			Symbol:   "GOOGL",
			Quantity: decimal.NewFromInt(50),
		},
	}

	response := ToHoldingListResponse(holdings)

	assert.NotNil(t, response)
	assert.Equal(t, 2, response.Total)
	assert.Len(t, response.Holdings, 2)
}

// Test Market Data DTOs
func TestToQuoteResponse(t *testing.T) {
	quote := &services.Quote{
		Symbol:        "AAPL",
		Price:         decimal.NewFromFloat(150.25),
		Open:          decimal.NewFromFloat(149.50),
		High:          decimal.NewFromFloat(151.00),
		Low:           decimal.NewFromFloat(149.00),
		Volume:        1000000,
		PreviousClose: decimal.NewFromFloat(149.00),
		Change:        decimal.NewFromFloat(1.25),
		ChangePercent: decimal.NewFromFloat(0.84),
		LastUpdated:   time.Now(),
	}

	response := ToQuoteResponse(quote)

	assert.NotNil(t, response)
	assert.Equal(t, "AAPL", response.Symbol)
	assert.Equal(t, decimal.NewFromFloat(150.25), response.Price)
	assert.Equal(t, int64(1000000), response.Volume)
}

func TestToQuoteResponse_Nil(t *testing.T) {
	response := ToQuoteResponse(nil)
	assert.Nil(t, response)
}

func TestToQuotesResponse(t *testing.T) {
	quotes := map[string]*services.Quote{
		"AAPL": {
			Symbol: "AAPL",
			Price:  decimal.NewFromFloat(150.25),
		},
		"GOOGL": {
			Symbol: "GOOGL",
			Price:  decimal.NewFromFloat(2800.50),
		},
	}

	response := ToQuotesResponse(quotes)

	assert.NotNil(t, response)
	assert.Len(t, response.Quotes, 2)
	assert.NotNil(t, response.Quotes["AAPL"])
	assert.NotNil(t, response.Quotes["GOOGL"])
}

func TestToHistoricalPriceResponse(t *testing.T) {
	now := time.Now()
	price := &services.HistoricalPrice{
		Date:   now,
		Open:   decimal.NewFromFloat(149.50),
		High:   decimal.NewFromFloat(151.00),
		Low:    decimal.NewFromFloat(149.00),
		Close:  decimal.NewFromFloat(150.25),
		Volume: 1000000,
	}

	response := ToHistoricalPriceResponse(price)

	assert.NotNil(t, response)
	assert.Equal(t, now, response.Date)
	assert.Equal(t, decimal.NewFromFloat(150.25), response.Close)
	assert.Equal(t, int64(1000000), response.Volume)
}

func TestToHistoricalPriceResponse_Nil(t *testing.T) {
	response := ToHistoricalPriceResponse(nil)
	assert.Nil(t, response)
}

func TestToHistoricalPricesResponse(t *testing.T) {
	prices := []*services.HistoricalPrice{
		{
			Date:  time.Now(),
			Close: decimal.NewFromFloat(150.25),
		},
		{
			Date:  time.Now().AddDate(0, 0, -1),
			Close: decimal.NewFromFloat(149.50),
		},
	}

	response := ToHistoricalPricesResponse(prices)

	assert.NotNil(t, response)
	assert.Len(t, response.Prices, 2)
}

func TestToExchangeRateResponse(t *testing.T) {
	rate := decimal.NewFromFloat(0.85)

	response := ToExchangeRateResponse("USD", "EUR", rate)

	assert.NotNil(t, response)
	assert.Equal(t, "USD", response.From)
	assert.Equal(t, "EUR", response.To)
	assert.Equal(t, rate, response.Rate)
}

// Test Performance Analytics DTOs
func TestToPerformanceMetricsResponse(t *testing.T) {
	metrics := &services.PerformanceMetrics{
		StartDate:           time.Now().AddDate(-1, 0, 0),
		EndDate:             time.Now(),
		StartingValue:       decimal.NewFromInt(10000),
		EndingValue:         decimal.NewFromInt(12000),
		TotalReturn:         decimal.NewFromInt(2000),
		TotalReturnPct:      decimal.NewFromInt(20),
		TimeWeightedReturn:  decimal.NewFromInt(20),
		MoneyWeightedReturn: decimal.NewFromInt(18),
		AnnualizedReturn:    decimal.NewFromInt(19),
		TotalDeposits:       decimal.NewFromInt(10000),
		TotalWithdrawals:    decimal.Zero,
		NetCashFlow:         decimal.NewFromInt(10000),
		Years:               1.0,
	}

	response := ToPerformanceMetricsResponse(metrics)

	assert.NotNil(t, response)
	assert.Equal(t, decimal.NewFromInt(10000), response.StartingValue)
	assert.Equal(t, decimal.NewFromInt(12000), response.EndingValue)
	assert.Equal(t, decimal.NewFromInt(2000), response.TotalReturn)
	assert.Equal(t, 1.0, response.Years)
}

func TestToPerformanceMetricsResponse_Nil(t *testing.T) {
	response := ToPerformanceMetricsResponse(nil)
	assert.Nil(t, response)
}

func TestToTWRResponse(t *testing.T) {
	twr := &services.TWRResult{
		StartDate:     time.Now().AddDate(-1, 0, 0),
		EndDate:       time.Now(),
		TWR:           decimal.NewFromFloat(0.20),
		TWRPercent:    decimal.NewFromInt(20),
		AnnualizedTWR: decimal.NewFromInt(20),
		NumPeriods:    365,
		StartingValue: decimal.NewFromInt(10000),
		EndingValue:   decimal.NewFromInt(12000),
	}

	response := ToTWRResponse(twr)

	assert.NotNil(t, response)
	assert.Equal(t, decimal.NewFromInt(20), response.TWRPercent)
	assert.Equal(t, 365, response.NumPeriods)
}

func TestToTWRResponse_Nil(t *testing.T) {
	response := ToTWRResponse(nil)
	assert.Nil(t, response)
}

func TestToMWRResponse(t *testing.T) {
	mwr := &services.MWRResult{
		StartDate:     time.Now().AddDate(-1, 0, 0),
		EndDate:       time.Now(),
		MWR:           decimal.NewFromFloat(0.18),
		MWRPercent:    decimal.NewFromInt(18),
		AnnualizedMWR: decimal.NewFromInt(18),
		TotalCashFlow: decimal.NewFromInt(1000),
		StartingValue: decimal.NewFromInt(10000),
		EndingValue:   decimal.NewFromInt(12000),
	}

	response := ToMWRResponse(mwr)

	assert.NotNil(t, response)
	assert.Equal(t, decimal.NewFromInt(18), response.MWRPercent)
	assert.Equal(t, decimal.NewFromInt(1000), response.TotalCashFlow)
}

func TestToMWRResponse_Nil(t *testing.T) {
	response := ToMWRResponse(nil)
	assert.Nil(t, response)
}

func TestToAnnualizedReturnResponse(t *testing.T) {
	result := &services.AnnualizedReturnResult{
		StartDate:        time.Now().AddDate(-1, 0, 0),
		EndDate:          time.Now(),
		TotalReturn:      decimal.NewFromInt(2000),
		TotalReturnPct:   decimal.NewFromInt(20),
		AnnualizedReturn: decimal.NewFromInt(20),
		Years:            1.0,
	}

	response := ToAnnualizedReturnResponse(result)

	assert.NotNil(t, response)
	assert.Equal(t, decimal.NewFromInt(2000), response.TotalReturn)
	assert.Equal(t, decimal.NewFromInt(20), response.AnnualizedReturn)
	assert.Equal(t, 1.0, response.Years)
}

func TestToAnnualizedReturnResponse_Nil(t *testing.T) {
	response := ToAnnualizedReturnResponse(nil)
	assert.Nil(t, response)
}

func TestToBenchmarkComparisonResponse(t *testing.T) {
	comparison := &services.BenchmarkComparisonResult{
		StartDate:           time.Now().AddDate(-1, 0, 0),
		EndDate:             time.Now(),
		BenchmarkSymbol:     "SPY",
		PortfolioReturn:     decimal.NewFromInt(20),
		BenchmarkReturn:     decimal.NewFromInt(15),
		Alpha:               decimal.NewFromInt(5),
		PortfolioAnnualized: decimal.NewFromInt(20),
		BenchmarkAnnualized: decimal.NewFromInt(15),
		Outperformance:      decimal.NewFromInt(5),
	}

	response := ToBenchmarkComparisonResponse(comparison)

	assert.NotNil(t, response)
	assert.Equal(t, "SPY", response.BenchmarkSymbol)
	assert.Equal(t, decimal.NewFromInt(5), response.Alpha)
	assert.Equal(t, decimal.NewFromInt(5), response.Outperformance)
}

func TestToBenchmarkComparisonResponse_Nil(t *testing.T) {
	response := ToBenchmarkComparisonResponse(nil)
	assert.Nil(t, response)
}

// Test Performance Snapshot DTOs
func TestToPerformanceSnapshotResponse(t *testing.T) {
	snapshotID := uuid.New()
	portfolioID := uuid.New()
	now := time.Now()

	snapshot := &models.PerformanceSnapshot{
		ID:             snapshotID,
		PortfolioID:    portfolioID,
		Date:           now,
		TotalValue:     decimal.NewFromInt(12000),
		TotalCostBasis: decimal.NewFromInt(10000),
		TotalReturn:    decimal.NewFromInt(2000),
		TotalReturnPct: decimal.NewFromInt(20),
		CreatedAt:      now,
	}

	response := ToPerformanceSnapshotResponse(snapshot)

	assert.NotNil(t, response)
	assert.Equal(t, snapshotID, response.ID)
	assert.Equal(t, portfolioID, response.PortfolioID)
	assert.Equal(t, decimal.NewFromInt(12000), response.TotalValue)
	assert.Equal(t, decimal.NewFromInt(20), response.TotalReturnPct)
}

func TestToPerformanceSnapshotResponse_Nil(t *testing.T) {
	response := ToPerformanceSnapshotResponse(nil)
	assert.Nil(t, response)
}

func TestToPerformanceSnapshotListResponse(t *testing.T) {
	snapshots := []*models.PerformanceSnapshot{
		{
			ID:          uuid.New(),
			TotalValue:  decimal.NewFromInt(12000),
			TotalReturn: decimal.NewFromInt(2000),
		},
		{
			ID:          uuid.New(),
			TotalValue:  decimal.NewFromInt(13000),
			TotalReturn: decimal.NewFromInt(3000),
		},
	}

	response := ToPerformanceSnapshotListResponse(snapshots)

	assert.NotNil(t, response)
	assert.Equal(t, 2, response.Total)
	assert.Len(t, response.Snapshots, 2)
}
