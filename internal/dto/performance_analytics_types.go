package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

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
