package cmd

import (
	"fmt"

	"github.com/lenon/portfolios/internal/cli"
	"github.com/spf13/cobra"
)

var (
	startDate string
	endDate   string
	benchmark string
)

var performanceCmd = &cobra.Command{
	Use:     "performance",
	Aliases: []string{"perf", "pf"},
	Short:   "View performance analytics",
	Long:    "Analyze portfolio performance with metrics like TWR, MWR, and benchmark comparisons",
}

var performanceShowCmd = &cobra.Command{
	Use:   "show <portfolio-id>",
	Short: "Show performance metrics",
	Long:  "Display comprehensive performance metrics for a portfolio",
	Args:  cobra.ExactArgs(1),
	RunE:  runPerformanceShow,
}

var performanceCompareCmd = &cobra.Command{
	Use:   "compare <portfolio-id-1> <portfolio-id-2>",
	Short: "Compare two portfolios",
	Long:  "Compare performance metrics between two portfolios",
	Args:  cobra.ExactArgs(2),
	RunE:  runPerformanceCompare,
}

var performanceBenchmarkCmd = &cobra.Command{
	Use:   "benchmark <portfolio-id> <symbol>",
	Short: "Compare against benchmark",
	Long:  "Compare portfolio performance against a benchmark (e.g., SPY for S&P 500)",
	Args:  cobra.ExactArgs(2),
	RunE:  runPerformanceBenchmark,
}

var performanceSnapshotsCmd = &cobra.Command{
	Use:   "snapshots <portfolio-id>",
	Short: "View performance snapshots",
	Long:  "Display historical performance snapshots for a portfolio",
	Args:  cobra.ExactArgs(1),
	RunE:  runPerformanceSnapshots,
}

func init() {
	performanceCmd.AddCommand(performanceShowCmd)
	performanceCmd.AddCommand(performanceCompareCmd)
	performanceCmd.AddCommand(performanceBenchmarkCmd)
	performanceCmd.AddCommand(performanceSnapshotsCmd)

	// Add date range flags
	performanceShowCmd.Flags().StringVar(&startDate, "start", "", "Start date (YYYY-MM-DD)")
	performanceShowCmd.Flags().StringVar(&endDate, "end", "", "End date (YYYY-MM-DD)")
	performanceCompareCmd.Flags().StringVar(&startDate, "start", "", "Start date (YYYY-MM-DD)")
	performanceCompareCmd.Flags().StringVar(&endDate, "end", "", "End date (YYYY-MM-DD)")
	performanceBenchmarkCmd.Flags().StringVar(&startDate, "start", "", "Start date (YYYY-MM-DD)")
	performanceBenchmarkCmd.Flags().StringVar(&endDate, "end", "", "End date (YYYY-MM-DD)")
}

func runPerformanceShow(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	portfolioID := args[0]
	client := cli.NewClientFromConfig(config)

	// Build query params
	path := "/api/v1/portfolios/" + portfolioID + "/performance"
	if startDate != "" || endDate != "" {
		path += "?"
		if startDate != "" {
			path += "start_date=" + startDate
		}
		if endDate != "" {
			if startDate != "" {
				path += "&"
			}
			path += "end_date=" + endDate
		}
	}

	var metrics struct {
		TimeWeightedReturn    float64 `json:"time_weighted_return"`
		MoneyWeightedReturn   float64 `json:"money_weighted_return"`
		AnnualizedReturn      float64 `json:"annualized_return"`
		TotalReturn           float64 `json:"total_return"`
		TotalGain             float64 `json:"total_gain"`
		DividendIncome        float64 `json:"dividend_income"`
		RealizedGain          float64 `json:"realized_gain"`
		UnrealizedGain        float64 `json:"unrealized_gain"`
		BeginningValue        float64 `json:"beginning_value"`
		EndingValue           float64 `json:"ending_value"`
		NetContributions      float64 `json:"net_contributions"`
		Period                string  `json:"period"`
	}

	if err := client.Request("GET", path, nil, &metrics); err != nil {
		return err
	}

	format := cli.OutputFormat(config.OutputFormat)
	if outputFormat != "" {
		format = cli.OutputFormat(outputFormat)
	}

	if format == cli.OutputFormatJSON {
		return cli.OutputJSON(metrics)
	}

	// Display formatted performance metrics
	fmt.Println(cli.RenderSection("Performance Metrics"))
	if metrics.Period != "" {
		fmt.Printf("\nPeriod: %s\n", metrics.Period)
	}
	fmt.Println()

	fmt.Println(cli.RenderSection("Returns"))
	fmt.Println(cli.RenderKeyValue("Time-Weighted Return", fmt.Sprintf("%.2f%%", metrics.TimeWeightedReturn)))
	fmt.Println(cli.RenderKeyValue("Money-Weighted Return", fmt.Sprintf("%.2f%%", metrics.MoneyWeightedReturn)))
	fmt.Println(cli.RenderKeyValue("Annualized Return", fmt.Sprintf("%.2f%%", metrics.AnnualizedReturn)))
	fmt.Println(cli.RenderKeyValue("Total Return", fmt.Sprintf("%.2f%%", metrics.TotalReturn)))

	fmt.Println()
	fmt.Println(cli.RenderSection("Gains & Income"))
	fmt.Println(cli.RenderKeyValue("Total Gain", formatGain(metrics.TotalGain)))
	fmt.Println(cli.RenderKeyValue("Realized Gain", formatGain(metrics.RealizedGain)))
	fmt.Println(cli.RenderKeyValue("Unrealized Gain", formatGain(metrics.UnrealizedGain)))
	fmt.Println(cli.RenderKeyValue("Dividend Income", fmt.Sprintf("$%.2f", metrics.DividendIncome)))

	fmt.Println()
	fmt.Println(cli.RenderSection("Portfolio Value"))
	fmt.Println(cli.RenderKeyValue("Beginning Value", fmt.Sprintf("$%.2f", metrics.BeginningValue)))
	fmt.Println(cli.RenderKeyValue("Ending Value", fmt.Sprintf("$%.2f", metrics.EndingValue)))
	fmt.Println(cli.RenderKeyValue("Net Contributions", formatGain(metrics.NetContributions)))

	return nil
}

func runPerformanceCompare(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	portfolioID1 := args[0]
	portfolioID2 := args[1]
	client := cli.NewClientFromConfig(config)

	// Build query params
	queryParams := fmt.Sprintf("?portfolio_ids=%s,%s", portfolioID1, portfolioID2)
	if startDate != "" {
		queryParams += "&start_date=" + startDate
	}
	if endDate != "" {
		queryParams += "&end_date=" + endDate
	}

	var comparison struct {
		Portfolios []struct {
			ID                 uint    `json:"id"`
			Name               string  `json:"name"`
			TimeWeightedReturn float64 `json:"time_weighted_return"`
			MoneyWeightedReturn float64 `json:"money_weighted_return"`
			AnnualizedReturn   float64 `json:"annualized_return"`
			TotalReturn        float64 `json:"total_return"`
			TotalGain          float64 `json:"total_gain"`
			EndingValue        float64 `json:"ending_value"`
		} `json:"portfolios"`
	}

	if err := client.Request("GET", "/api/v1/portfolios/compare"+queryParams, nil, &comparison); err != nil {
		return err
	}

	format := cli.OutputFormat(config.OutputFormat)
	if outputFormat != "" {
		format = cli.OutputFormat(outputFormat)
	}

	if format == cli.OutputFormatJSON {
		return cli.OutputJSON(comparison)
	}

	if len(comparison.Portfolios) == 0 {
		cli.PrintInfo("No data available for comparison")
		return nil
	}

	// Display as table
	headers := []string{"Portfolio", "TWR %", "MWR %", "Annualized %", "Total Return %", "Total Gain", "Value"}
	rows := make([][]string, len(comparison.Portfolios))

	for i, p := range comparison.Portfolios {
		rows[i] = []string{
			fmt.Sprintf("%s (#%d)", p.Name, p.ID),
			fmt.Sprintf("%.2f%%", p.TimeWeightedReturn),
			fmt.Sprintf("%.2f%%", p.MoneyWeightedReturn),
			fmt.Sprintf("%.2f%%", p.AnnualizedReturn),
			fmt.Sprintf("%.2f%%", p.TotalReturn),
			formatGain(p.TotalGain),
			fmt.Sprintf("$%.2f", p.EndingValue),
		}
	}

	cli.OutputTable(headers, rows)
	return nil
}

func runPerformanceBenchmark(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	portfolioID := args[0]
	benchmarkSymbol := args[1]
	client := cli.NewClientFromConfig(config)

	// Build query params
	queryParams := fmt.Sprintf("?benchmark=%s", benchmarkSymbol)
	if startDate != "" {
		queryParams += "&start_date=" + startDate
	}
	if endDate != "" {
		queryParams += "&end_date=" + endDate
	}

	var comparison struct {
		Portfolio struct {
			TimeWeightedReturn  float64 `json:"time_weighted_return"`
			AnnualizedReturn    float64 `json:"annualized_return"`
			TotalReturn         float64 `json:"total_return"`
		} `json:"portfolio"`
		Benchmark struct {
			Symbol             string  `json:"symbol"`
			TimeWeightedReturn float64 `json:"time_weighted_return"`
			AnnualizedReturn   float64 `json:"annualized_return"`
			TotalReturn        float64 `json:"total_return"`
		} `json:"benchmark"`
		Difference struct {
			TimeWeightedReturn float64 `json:"time_weighted_return"`
			AnnualizedReturn   float64 `json:"annualized_return"`
			TotalReturn        float64 `json:"total_return"`
		} `json:"difference"`
	}

	if err := client.Request("GET", "/api/v1/portfolios/"+portfolioID+"/performance"+queryParams, nil, &comparison); err != nil {
		return err
	}

	format := cli.OutputFormat(config.OutputFormat)
	if outputFormat != "" {
		format = cli.OutputFormat(outputFormat)
	}

	if format == cli.OutputFormatJSON {
		return cli.OutputJSON(comparison)
	}

	// Display comparison
	fmt.Println(cli.RenderSection(fmt.Sprintf("Benchmark Comparison: %s", benchmarkSymbol)))
	fmt.Println()

	headers := []string{"Metric", "Portfolio", "Benchmark", "Difference"}
	rows := [][]string{
		{
			"Time-Weighted Return",
			fmt.Sprintf("%.2f%%", comparison.Portfolio.TimeWeightedReturn),
			fmt.Sprintf("%.2f%%", comparison.Benchmark.TimeWeightedReturn),
			formatDifference(comparison.Difference.TimeWeightedReturn),
		},
		{
			"Annualized Return",
			fmt.Sprintf("%.2f%%", comparison.Portfolio.AnnualizedReturn),
			fmt.Sprintf("%.2f%%", comparison.Benchmark.AnnualizedReturn),
			formatDifference(comparison.Difference.AnnualizedReturn),
		},
		{
			"Total Return",
			fmt.Sprintf("%.2f%%", comparison.Portfolio.TotalReturn),
			fmt.Sprintf("%.2f%%", comparison.Benchmark.TotalReturn),
			formatDifference(comparison.Difference.TotalReturn),
		},
	}

	cli.OutputTable(headers, rows)
	return nil
}

func runPerformanceSnapshots(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	portfolioID := args[0]
	client := cli.NewClientFromConfig(config)

	// Build query params
	path := "/api/v1/portfolios/" + portfolioID + "/snapshots"
	if startDate != "" || endDate != "" {
		path += "?"
		if startDate != "" {
			path += "start_date=" + startDate
		}
		if endDate != "" {
			if startDate != "" {
				path += "&"
			}
			path += "end_date=" + endDate
		}
	}

	var snapshots []struct {
		ID            uint    `json:"id"`
		SnapshotDate  string  `json:"snapshot_date"`
		TotalValue    float64 `json:"total_value"`
		TotalGain     float64 `json:"total_gain"`
		TotalReturn   float64 `json:"total_return"`
		DayChange     float64 `json:"day_change"`
		DayChangePercent float64 `json:"day_change_percent"`
	}

	if err := client.Request("GET", path, nil, &snapshots); err != nil {
		return err
	}

	if len(snapshots) == 0 {
		cli.PrintInfo("No snapshots found for this portfolio")
		return nil
	}

	format := cli.OutputFormat(config.OutputFormat)
	if outputFormat != "" {
		format = cli.OutputFormat(outputFormat)
	}

	headers := []string{"Date", "Total Value", "Total Gain", "Return %", "Day Change", "Day Change %"}
	rows := make([][]string, len(snapshots))

	for i, s := range snapshots {
		rows[i] = []string{
			s.SnapshotDate,
			fmt.Sprintf("$%.2f", s.TotalValue),
			formatGain(s.TotalGain),
			fmt.Sprintf("%.2f%%", s.TotalReturn),
			formatGain(s.DayChange),
			formatDifference(s.DayChangePercent),
		}
	}

	return cli.Output(format, headers, rows, snapshots)
}

func formatDifference(diff float64) string {
	if diff >= 0 {
		return fmt.Sprintf("+%.2f%%", diff)
	}
	return fmt.Sprintf("%.2f%%", diff)
}
