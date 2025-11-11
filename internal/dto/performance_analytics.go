package dto

import (
	"time"

	"github.com/lenon/portfolios/internal/services"
	"github.com/shopspring/decimal"
)

// PerformanceMetricsRequest represents request parameters for performance metrics
type PerformanceMetricsRequest struct {
	StartDate time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate   time.Time `form:"end_date" time_format:"2006-01-02"`
}

// BenchmarkComparisonRequest represents request parameters for benchmark comparison
type BenchmarkComparisonRequest struct {
	StartDate       time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate         time.Time `form:"end_date" time_format:"2006-01-02"`
	BenchmarkSymbol string    `form:"benchmark_symbol"`
}

// PerformanceMetricsResponse represents comprehensive performance metrics
type PerformanceMetricsResponse struct {
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

// TWRResponse represents Time-Weighted Return response
type TWRResponse struct {
	StartDate     time.Time       `json:"start_date"`
	EndDate       time.Time       `json:"end_date"`
	TWR           decimal.Decimal `json:"twr"`
	TWRPercent    decimal.Decimal `json:"twr_percent"`
	AnnualizedTWR decimal.Decimal `json:"annualized_twr"`
	NumPeriods    int             `json:"num_periods"`
	StartingValue decimal.Decimal `json:"starting_value"`
	EndingValue   decimal.Decimal `json:"ending_value"`
}

// MWRResponse represents Money-Weighted Return response
type MWRResponse struct {
	StartDate     time.Time       `json:"start_date"`
	EndDate       time.Time       `json:"end_date"`
	MWR           decimal.Decimal `json:"mwr"`
	MWRPercent    decimal.Decimal `json:"mwr_percent"`
	AnnualizedMWR decimal.Decimal `json:"annualized_mwr"`
	TotalCashFlow decimal.Decimal `json:"total_cash_flow"`
	StartingValue decimal.Decimal `json:"starting_value"`
	EndingValue   decimal.Decimal `json:"ending_value"`
}

// AnnualizedReturnResponse represents annualized return response
type AnnualizedReturnResponse struct {
	StartDate        time.Time       `json:"start_date"`
	EndDate          time.Time       `json:"end_date"`
	TotalReturn      decimal.Decimal `json:"total_return"`
	TotalReturnPct   decimal.Decimal `json:"total_return_pct"`
	AnnualizedReturn decimal.Decimal `json:"annualized_return"`
	Years            float64         `json:"years"`
}

// BenchmarkComparisonResponse represents benchmark comparison response
type BenchmarkComparisonResponse struct {
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

// ToPerformanceMetricsResponse converts service result to DTO
func ToPerformanceMetricsResponse(metrics *services.PerformanceMetrics) *PerformanceMetricsResponse {
	if metrics == nil {
		return nil
	}

	return &PerformanceMetricsResponse{
		StartDate:           metrics.StartDate,
		EndDate:             metrics.EndDate,
		StartingValue:       metrics.StartingValue,
		EndingValue:         metrics.EndingValue,
		TotalReturn:         metrics.TotalReturn,
		TotalReturnPct:      metrics.TotalReturnPct,
		TimeWeightedReturn:  metrics.TimeWeightedReturn,
		MoneyWeightedReturn: metrics.MoneyWeightedReturn,
		AnnualizedReturn:    metrics.AnnualizedReturn,
		TotalDeposits:       metrics.TotalDeposits,
		TotalWithdrawals:    metrics.TotalWithdrawals,
		NetCashFlow:         metrics.NetCashFlow,
		Years:               metrics.Years,
	}
}

// ToTWRResponse converts service result to DTO
func ToTWRResponse(twr *services.TWRResult) *TWRResponse {
	if twr == nil {
		return nil
	}

	return &TWRResponse{
		StartDate:     twr.StartDate,
		EndDate:       twr.EndDate,
		TWR:           twr.TWR,
		TWRPercent:    twr.TWRPercent,
		AnnualizedTWR: twr.AnnualizedTWR,
		NumPeriods:    twr.NumPeriods,
		StartingValue: twr.StartingValue,
		EndingValue:   twr.EndingValue,
	}
}

// ToMWRResponse converts service result to DTO
func ToMWRResponse(mwr *services.MWRResult) *MWRResponse {
	if mwr == nil {
		return nil
	}

	return &MWRResponse{
		StartDate:     mwr.StartDate,
		EndDate:       mwr.EndDate,
		MWR:           mwr.MWR,
		MWRPercent:    mwr.MWRPercent,
		AnnualizedMWR: mwr.AnnualizedMWR,
		TotalCashFlow: mwr.TotalCashFlow,
		StartingValue: mwr.StartingValue,
		EndingValue:   mwr.EndingValue,
	}
}

// ToAnnualizedReturnResponse converts service result to DTO
func ToAnnualizedReturnResponse(result *services.AnnualizedReturnResult) *AnnualizedReturnResponse {
	if result == nil {
		return nil
	}

	return &AnnualizedReturnResponse{
		StartDate:        result.StartDate,
		EndDate:          result.EndDate,
		TotalReturn:      result.TotalReturn,
		TotalReturnPct:   result.TotalReturnPct,
		AnnualizedReturn: result.AnnualizedReturn,
		Years:            result.Years,
	}
}

// ToBenchmarkComparisonResponse converts service result to DTO
func ToBenchmarkComparisonResponse(comparison *services.BenchmarkComparisonResult) *BenchmarkComparisonResponse {
	if comparison == nil {
		return nil
	}

	return &BenchmarkComparisonResponse{
		StartDate:           comparison.StartDate,
		EndDate:             comparison.EndDate,
		BenchmarkSymbol:     comparison.BenchmarkSymbol,
		PortfolioReturn:     comparison.PortfolioReturn,
		BenchmarkReturn:     comparison.BenchmarkReturn,
		Alpha:               comparison.Alpha,
		PortfolioAnnualized: comparison.PortfolioAnnualized,
		BenchmarkAnnualized: comparison.BenchmarkAnnualized,
		Outperformance:      comparison.Outperformance,
	}
}
