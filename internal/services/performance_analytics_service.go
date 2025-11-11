package services

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
	"github.com/shopspring/decimal"
)

// PerformanceAnalyticsService defines the interface for performance analytics operations
type PerformanceAnalyticsService interface {
	CalculateTWR(portfolioID, userID string, startDate, endDate time.Time) (*TWRResult, error)
	CalculateMWR(portfolioID, userID string, startDate, endDate time.Time) (*MWRResult, error)
	CalculateAnnualizedReturn(portfolioID, userID string, startDate, endDate time.Time) (*AnnualizedReturnResult, error)
	CompareToBenchmark(portfolioID, userID, benchmarkSymbol string, startDate, endDate time.Time) (*BenchmarkComparisonResult, error)
	GetPerformanceMetrics(portfolioID, userID string, startDate, endDate time.Time) (*PerformanceMetrics, error)
}

// TWRResult represents the Time-Weighted Return calculation result
type TWRResult struct {
	StartDate     time.Time       `json:"start_date"`
	EndDate       time.Time       `json:"end_date"`
	TWR           decimal.Decimal `json:"twr"`
	TWRPercent    decimal.Decimal `json:"twr_percent"`
	AnnualizedTWR decimal.Decimal `json:"annualized_twr"`
	NumPeriods    int             `json:"num_periods"`
	StartingValue decimal.Decimal `json:"starting_value"`
	EndingValue   decimal.Decimal `json:"ending_value"`
}

// MWRResult represents the Money-Weighted Return (IRR) calculation result
type MWRResult struct {
	StartDate     time.Time       `json:"start_date"`
	EndDate       time.Time       `json:"end_date"`
	MWR           decimal.Decimal `json:"mwr"`
	MWRPercent    decimal.Decimal `json:"mwr_percent"`
	AnnualizedMWR decimal.Decimal `json:"annualized_mwr"`
	TotalCashFlow decimal.Decimal `json:"total_cash_flow"`
	StartingValue decimal.Decimal `json:"starting_value"`
	EndingValue   decimal.Decimal `json:"ending_value"`
}

// AnnualizedReturnResult represents annualized return calculation
type AnnualizedReturnResult struct {
	StartDate        time.Time       `json:"start_date"`
	EndDate          time.Time       `json:"end_date"`
	TotalReturn      decimal.Decimal `json:"total_return"`
	TotalReturnPct   decimal.Decimal `json:"total_return_pct"`
	AnnualizedReturn decimal.Decimal `json:"annualized_return"`
	Years            float64         `json:"years"`
}

// BenchmarkComparisonResult represents portfolio vs benchmark comparison
type BenchmarkComparisonResult struct {
	StartDate           time.Time       `json:"start_date"`
	EndDate             time.Time       `json:"end_date"`
	BenchmarkSymbol     string          `json:"benchmark_symbol"`
	PortfolioReturn     decimal.Decimal `json:"portfolio_return"`
	BenchmarkReturn     decimal.Decimal `json:"benchmark_return"`
	Alpha               decimal.Decimal `json:"alpha"`
	PortfolioAnnualized decimal.Decimal `json:"portfolio_annualized"`
	BenchmarkAnnualized decimal.Decimal `json:"benchmark_annualized"`
	Outperformance      decimal.Decimal `json:"outperformance"`
}

// PerformanceMetrics represents comprehensive performance metrics
type PerformanceMetrics struct {
	StartDate           time.Time       `json:"start_date"`
	EndDate             time.Time       `json:"end_date"`
	StartingValue       decimal.Decimal `json:"starting_value"`
	EndingValue         decimal.Decimal `json:"ending_value"`
	TotalReturn         decimal.Decimal `json:"total_return"`
	TotalReturnPct      decimal.Decimal `json:"total_return_pct"`
	TimeWeightedReturn  decimal.Decimal `json:"time_weighted_return"`
	MoneyWeightedReturn decimal.Decimal `json:"money_weighted_return"`
	AnnualizedReturn    decimal.Decimal `json:"annualized_return"`
	TotalDeposits       decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals    decimal.Decimal `json:"total_withdrawals"`
	NetCashFlow         decimal.Decimal `json:"net_cash_flow"`
	Years               float64         `json:"years"`
}

// performanceAnalyticsService implements PerformanceAnalyticsService interface
type performanceAnalyticsService struct {
	portfolioRepo   repository.PortfolioRepository
	transactionRepo repository.TransactionRepository
	snapshotRepo    repository.PerformanceSnapshotRepository
	marketDataSvc   MarketDataService
}

// NewPerformanceAnalyticsService creates a new PerformanceAnalyticsService instance
func NewPerformanceAnalyticsService(
	portfolioRepo repository.PortfolioRepository,
	transactionRepo repository.TransactionRepository,
	snapshotRepo repository.PerformanceSnapshotRepository,
	marketDataSvc MarketDataService,
) PerformanceAnalyticsService {
	return &performanceAnalyticsService{
		portfolioRepo:   portfolioRepo,
		transactionRepo: transactionRepo,
		snapshotRepo:    snapshotRepo,
		marketDataSvc:   marketDataSvc,
	}
}

// CalculateTWR calculates the Time-Weighted Return for a portfolio
// TWR eliminates the effect of cash flows and measures the compound rate of growth
func (s *performanceAnalyticsService) CalculateTWR(
	portfolioID, userID string,
	startDate, endDate time.Time,
) (*TWRResult, error) {
	// Verify portfolio ownership
	if err := s.verifyPortfolioOwnership(portfolioID, userID); err != nil {
		return nil, err
	}

	// Get performance snapshots for the period
	snapshots, err := s.snapshotRepo.FindByPortfolioIDAndDateRange(portfolioID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve snapshots: %w", err)
	}

	if len(snapshots) < 2 {
		return nil, fmt.Errorf("insufficient data: need at least 2 snapshots for TWR calculation")
	}

	// Get all transactions in the period to identify cash flow dates
	transactions, err := s.transactionRepo.FindByPortfolioIDWithFilters(portfolioID, nil, &startDate, &endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transactions: %w", err)
	}

	// Sort snapshots by date
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Date.Before(snapshots[j].Date)
	})

	startingValue := snapshots[0].TotalValue
	endingValue := snapshots[len(snapshots)-1].TotalValue

	// Calculate TWR using sub-period returns
	// TWR = [(1 + R1) × (1 + R2) × ... × (1 + Rn)] - 1
	twrProduct := decimal.NewFromInt(1)

	for i := 1; i < len(snapshots); i++ {
		prevSnapshot := snapshots[i-1]
		currSnapshot := snapshots[i]

		// Calculate cash flows between periods
		cashFlow := s.calculateCashFlowBetweenDates(transactions, prevSnapshot.Date, currSnapshot.Date)

		// Adjust for cash flows: Return = (EndValue - CashFlow) / StartValue - 1
		adjustedEndValue := currSnapshot.TotalValue.Sub(cashFlow)

		if prevSnapshot.TotalValue.IsZero() {
			continue
		}

		periodReturn := adjustedEndValue.Div(prevSnapshot.TotalValue).Sub(decimal.NewFromInt(1))
		twrProduct = twrProduct.Mul(periodReturn.Add(decimal.NewFromInt(1)))
	}

	twr := twrProduct.Sub(decimal.NewFromInt(1))
	twrPercent := twr.Mul(decimal.NewFromInt(100))

	// Calculate annualized TWR
	years := endDate.Sub(startDate).Hours() / 24 / 365.25
	annualizedTWR := decimal.Zero
	if years > 0 {
		// Annualized TWR = (1 + TWR)^(1/years) - 1
		twrFloat, _ := twr.Add(decimal.NewFromInt(1)).Float64()
		annualizedFloat := math.Pow(twrFloat, 1/years) - 1
		annualizedTWR = decimal.NewFromFloat(annualizedFloat).Mul(decimal.NewFromInt(100))
	}

	return &TWRResult{
		StartDate:     startDate,
		EndDate:       endDate,
		TWR:           twr,
		TWRPercent:    twrPercent,
		AnnualizedTWR: annualizedTWR,
		NumPeriods:    len(snapshots) - 1,
		StartingValue: startingValue,
		EndingValue:   endingValue,
	}, nil
}

// CalculateMWR calculates the Money-Weighted Return (Internal Rate of Return)
// MWR considers the timing and size of cash flows
func (s *performanceAnalyticsService) CalculateMWR(
	portfolioID, userID string,
	startDate, endDate time.Time,
) (*MWRResult, error) {
	// Verify portfolio ownership
	if err := s.verifyPortfolioOwnership(portfolioID, userID); err != nil {
		return nil, err
	}

	// Get transactions for cash flow analysis
	transactions, err := s.transactionRepo.FindByPortfolioIDWithFilters(portfolioID, nil, &startDate, &endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transactions: %w", err)
	}

	// Get starting and ending values
	startSnapshot, err := s.getSnapshotNearDate(portfolioID, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get starting snapshot: %w", err)
	}

	endSnapshot, err := s.getSnapshotNearDate(portfolioID, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get ending snapshot: %w", err)
	}

	startingValue := startSnapshot.TotalValue
	endingValue := endSnapshot.TotalValue

	// Build cash flow series
	cashFlows := s.buildCashFlowSeries(transactions, startDate, endDate, startingValue, endingValue)

	if len(cashFlows) < 2 {
		return nil, fmt.Errorf("insufficient cash flows for MWR calculation")
	}

	// Calculate IRR using Newton-Raphson method
	mwr := s.calculateIRR(cashFlows)
	mwrPercent := mwr.Mul(decimal.NewFromInt(100))

	// Calculate total cash flow (excluding start and end)
	totalCashFlow := decimal.Zero
	for i := 1; i < len(cashFlows)-1; i++ {
		totalCashFlow = totalCashFlow.Add(cashFlows[i].Amount)
	}

	// Annualize MWR
	years := endDate.Sub(startDate).Hours() / 24 / 365.25
	annualizedMWR := decimal.Zero
	if years > 0 {
		mwrFloat, _ := mwr.Add(decimal.NewFromInt(1)).Float64()
		annualizedFloat := math.Pow(mwrFloat, 1/years) - 1
		annualizedMWR = decimal.NewFromFloat(annualizedFloat).Mul(decimal.NewFromInt(100))
	}

	return &MWRResult{
		StartDate:     startDate,
		EndDate:       endDate,
		MWR:           mwr,
		MWRPercent:    mwrPercent,
		AnnualizedMWR: annualizedMWR,
		TotalCashFlow: totalCashFlow,
		StartingValue: startingValue,
		EndingValue:   endingValue,
	}, nil
}

// CalculateAnnualizedReturn calculates the annualized return for a portfolio
func (s *performanceAnalyticsService) CalculateAnnualizedReturn(
	portfolioID, userID string,
	startDate, endDate time.Time,
) (*AnnualizedReturnResult, error) {
	// Verify portfolio ownership
	if err := s.verifyPortfolioOwnership(portfolioID, userID); err != nil {
		return nil, err
	}

	startSnapshot, err := s.getSnapshotNearDate(portfolioID, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get starting snapshot: %w", err)
	}

	endSnapshot, err := s.getSnapshotNearDate(portfolioID, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get ending snapshot: %w", err)
	}

	startingValue := startSnapshot.TotalValue
	endingValue := endSnapshot.TotalValue

	totalReturn := endingValue.Sub(startingValue)
	totalReturnPct := decimal.Zero
	if !startingValue.IsZero() {
		totalReturnPct = totalReturn.Div(startingValue).Mul(decimal.NewFromInt(100))
	}

	// Calculate annualized return
	years := endDate.Sub(startDate).Hours() / 24 / 365.25
	annualizedReturn := decimal.Zero

	if years > 0 && !startingValue.IsZero() {
		// Annualized Return = (EndValue / StartValue)^(1/years) - 1
		ratio, _ := endingValue.Div(startingValue).Float64()
		annualizedFloat := (math.Pow(ratio, 1/years) - 1) * 100
		annualizedReturn = decimal.NewFromFloat(annualizedFloat)
	}

	return &AnnualizedReturnResult{
		StartDate:        startDate,
		EndDate:          endDate,
		TotalReturn:      totalReturn,
		TotalReturnPct:   totalReturnPct,
		AnnualizedReturn: annualizedReturn,
		Years:            years,
	}, nil
}

// CompareToBenchmark compares portfolio performance to a benchmark index
func (s *performanceAnalyticsService) CompareToBenchmark(
	portfolioID, userID, benchmarkSymbol string,
	startDate, endDate time.Time,
) (*BenchmarkComparisonResult, error) {
	// Verify portfolio ownership
	if err := s.verifyPortfolioOwnership(portfolioID, userID); err != nil {
		return nil, err
	}

	// Calculate portfolio annualized return
	portfolioReturn, err := s.CalculateAnnualizedReturn(portfolioID, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate portfolio return: %w", err)
	}

	// Get benchmark historical data
	benchmarkPrices, err := s.marketDataSvc.GetHistoricalPrices(benchmarkSymbol, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get benchmark prices: %w", err)
	}

	if len(benchmarkPrices) < 2 {
		return nil, fmt.Errorf("insufficient benchmark data")
	}

	// Get starting and ending prices
	benchmarkStart := benchmarkPrices[0].Close
	benchmarkEnd := benchmarkPrices[len(benchmarkPrices)-1].Close

	// Calculate benchmark return
	benchmarkReturn := benchmarkEnd.Sub(benchmarkStart).Div(benchmarkStart)
	benchmarkReturnPct := benchmarkReturn.Mul(decimal.NewFromInt(100))

	// Calculate annualized benchmark return
	years := portfolioReturn.Years
	benchmarkAnnualized := decimal.Zero
	if years > 0 {
		benchmarkFloat, _ := benchmarkReturn.Add(decimal.NewFromInt(1)).Float64()
		annualizedFloat := (math.Pow(benchmarkFloat, 1/years) - 1) * 100
		benchmarkAnnualized = decimal.NewFromFloat(annualizedFloat)
	}

	// Calculate alpha (excess return)
	alpha := portfolioReturn.AnnualizedReturn.Sub(benchmarkAnnualized)
	outperformance := portfolioReturn.TotalReturnPct.Sub(benchmarkReturnPct)

	return &BenchmarkComparisonResult{
		StartDate:           startDate,
		EndDate:             endDate,
		BenchmarkSymbol:     benchmarkSymbol,
		PortfolioReturn:     portfolioReturn.TotalReturnPct,
		BenchmarkReturn:     benchmarkReturnPct,
		Alpha:               alpha,
		PortfolioAnnualized: portfolioReturn.AnnualizedReturn,
		BenchmarkAnnualized: benchmarkAnnualized,
		Outperformance:      outperformance,
	}, nil
}

// GetPerformanceMetrics returns comprehensive performance metrics
func (s *performanceAnalyticsService) GetPerformanceMetrics(
	portfolioID, userID string,
	startDate, endDate time.Time,
) (*PerformanceMetrics, error) {
	// Verify portfolio ownership
	if err := s.verifyPortfolioOwnership(portfolioID, userID); err != nil {
		return nil, err
	}

	// Get snapshots
	startSnapshot, err := s.getSnapshotNearDate(portfolioID, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get starting snapshot: %w", err)
	}

	endSnapshot, err := s.getSnapshotNearDate(portfolioID, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get ending snapshot: %w", err)
	}

	// Calculate basic metrics
	startingValue := startSnapshot.TotalValue
	endingValue := endSnapshot.TotalValue
	totalReturn := endingValue.Sub(startingValue)
	totalReturnPct := decimal.Zero
	if !startingValue.IsZero() {
		totalReturnPct = totalReturn.Div(startingValue).Mul(decimal.NewFromInt(100))
	}

	// Calculate years
	years := endDate.Sub(startDate).Hours() / 24 / 365.25

	// Calculate TWR
	twrResult, err := s.CalculateTWR(portfolioID, userID, startDate, endDate)
	timeWeightedReturn := decimal.Zero
	if err == nil {
		timeWeightedReturn = twrResult.TWRPercent
	}

	// Calculate MWR
	mwrResult, err := s.CalculateMWR(portfolioID, userID, startDate, endDate)
	moneyWeightedReturn := decimal.Zero
	netCashFlow := decimal.Zero
	if err == nil {
		moneyWeightedReturn = mwrResult.MWRPercent
		netCashFlow = mwrResult.TotalCashFlow
	}

	// Calculate annualized return
	annualizedReturn := decimal.Zero
	if years > 0 && !startingValue.IsZero() {
		ratio, _ := endingValue.Div(startingValue).Float64()
		annualizedFloat := (math.Pow(ratio, 1/years) - 1) * 100
		annualizedReturn = decimal.NewFromFloat(annualizedFloat)
	}

	// Get transaction totals
	transactions, err := s.transactionRepo.FindByPortfolioIDWithFilters(portfolioID, nil, &startDate, &endDate)
	totalDeposits := decimal.Zero
	totalWithdrawals := decimal.Zero

	if err == nil {
		for _, tx := range transactions {
			if tx.IsBuy() {
				totalDeposits = totalDeposits.Add(tx.GetTotalCost())
			} else if tx.IsSell() {
				totalWithdrawals = totalWithdrawals.Add(tx.GetProceeds())
			}
		}
	}

	return &PerformanceMetrics{
		StartDate:           startDate,
		EndDate:             endDate,
		StartingValue:       startingValue,
		EndingValue:         endingValue,
		TotalReturn:         totalReturn,
		TotalReturnPct:      totalReturnPct,
		TimeWeightedReturn:  timeWeightedReturn,
		MoneyWeightedReturn: moneyWeightedReturn,
		AnnualizedReturn:    annualizedReturn,
		TotalDeposits:       totalDeposits,
		TotalWithdrawals:    totalWithdrawals,
		NetCashFlow:         netCashFlow,
		Years:               years,
	}, nil
}

// Helper functions

func (s *performanceAnalyticsService) verifyPortfolioOwnership(portfolioID, userID string) error {
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return models.ErrUnauthorizedAccess
	}
	return nil
}

func (s *performanceAnalyticsService) getSnapshotNearDate(portfolioID string, date time.Time) (*models.PerformanceSnapshot, error) {
	// Try to get exact date first
	snapshot, err := s.snapshotRepo.FindByPortfolioIDAndDate(portfolioID, date)
	if err == nil {
		return snapshot, nil
	}

	// If not found, get snapshots in a range around the date
	startRange := date.AddDate(0, 0, -7) // 7 days before
	endRange := date.AddDate(0, 0, 7)    // 7 days after

	snapshots, err := s.snapshotRepo.FindByPortfolioIDAndDateRange(portfolioID, startRange, endRange)
	if err != nil || len(snapshots) == 0 {
		return nil, fmt.Errorf("no snapshot found near date %s", date.Format("2006-01-02"))
	}

	// Return the closest snapshot
	closest := snapshots[0]
	minDiff := math.Abs(date.Sub(snapshots[0].Date).Hours())

	for _, snap := range snapshots {
		diff := math.Abs(date.Sub(snap.Date).Hours())
		if diff < minDiff {
			minDiff = diff
			closest = snap
		}
	}

	return closest, nil
}

func (s *performanceAnalyticsService) calculateCashFlowBetweenDates(
	transactions []*models.Transaction,
	startDate, endDate time.Time,
) decimal.Decimal {
	cashFlow := decimal.Zero

	for _, tx := range transactions {
		if tx.Date.After(startDate) && tx.Date.Before(endDate) || tx.Date.Equal(endDate) {
			if tx.IsBuy() {
				// Deposits are positive cash flows
				cashFlow = cashFlow.Add(tx.GetTotalCost())
			} else if tx.IsSell() {
				// Withdrawals are negative cash flows
				cashFlow = cashFlow.Sub(tx.GetProceeds())
			}
		}
	}

	return cashFlow
}

// CashFlow represents a cash flow at a specific date
type CashFlow struct {
	Date   time.Time
	Amount decimal.Decimal
}

func (s *performanceAnalyticsService) buildCashFlowSeries(
	transactions []*models.Transaction,
	startDate, endDate time.Time,
	startingValue, endingValue decimal.Decimal,
) []CashFlow {
	cashFlows := []CashFlow{
		{Date: startDate, Amount: startingValue.Neg()}, // Starting value as negative outflow
	}

	// Add transaction cash flows
	for _, tx := range transactions {
		if (tx.Date.After(startDate) || tx.Date.Equal(startDate)) &&
			(tx.Date.Before(endDate) || tx.Date.Equal(endDate)) {
			amount := decimal.Zero
			if tx.IsBuy() {
				amount = tx.GetTotalCost().Neg() // Deposits are negative (outflows)
			} else if tx.IsSell() {
				amount = tx.GetProceeds() // Withdrawals are positive (inflows)
			}

			if !amount.IsZero() {
				cashFlows = append(cashFlows, CashFlow{
					Date:   tx.Date,
					Amount: amount,
				})
			}
		}
	}

	// Ending value as positive inflow
	cashFlows = append(cashFlows, CashFlow{
		Date:   endDate,
		Amount: endingValue,
	})

	// Sort by date
	sort.Slice(cashFlows, func(i, j int) bool {
		return cashFlows[i].Date.Before(cashFlows[j].Date)
	})

	return cashFlows
}

// calculateIRR calculates the Internal Rate of Return using Newton-Raphson method
func (s *performanceAnalyticsService) calculateIRR(cashFlows []CashFlow) decimal.Decimal {
	if len(cashFlows) < 2 {
		return decimal.Zero
	}

	// Initial guess
	rate := 0.1
	maxIterations := 100
	tolerance := 0.0001

	baseDate := cashFlows[0].Date

	for i := 0; i < maxIterations; i++ {
		npv := 0.0
		derivative := 0.0

		for _, cf := range cashFlows {
			years := cf.Date.Sub(baseDate).Hours() / 24 / 365.25
			amount, _ := cf.Amount.Float64()

			factor := math.Pow(1+rate, years)
			npv += amount / factor
			derivative -= years * amount / (factor * (1 + rate))
		}

		if math.Abs(npv) < tolerance {
			break
		}

		if derivative == 0 {
			break
		}

		rate = rate - npv/derivative

		// Bound the rate to reasonable values
		if rate < -0.99 {
			rate = -0.99
		}
		if rate > 10.0 {
			rate = 10.0
		}
	}

	return decimal.NewFromFloat(rate)
}
