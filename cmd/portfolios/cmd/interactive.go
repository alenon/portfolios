package cmd

import (
	"fmt"

	"github.com/lenon/portfolios/internal/cli"
	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Aliases: []string{"i", "ui"},
	Short:   "Interactive portfolio selector",
	Long:    "Launch an interactive UI to select and view portfolios",
	RunE:    runInteractive,
}

var portfolioSelectCmd = &cobra.Command{
	Use:   "select",
	Short: "Interactive portfolio selection",
	Long:  "Use an interactive UI to select a portfolio and view its details",
	RunE:  runPortfolioSelect,
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
	portfolioCmd.AddCommand(portfolioSelectCmd)
}

func runInteractive(cmd *cobra.Command, args []string) error {
	return runPortfolioSelect(cmd, args)
}

func runPortfolioSelect(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	client := cli.NewClientFromConfig(config)

	// Fetch portfolios
	var portfoliosResp []struct {
		ID          uint    `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		TotalValue  float64 `json:"total_value"`
	}

	if err := client.Request("GET", "/api/v1/portfolios", nil, &portfoliosResp); err != nil {
		return err
	}

	if len(portfoliosResp) == 0 {
		cli.PrintInfo("No portfolios found. Create one with 'portfolios portfolio create'")
		return nil
	}

	// Convert to selector format
	portfolios := make([]cli.Portfolio, len(portfoliosResp))
	for i, p := range portfoliosResp {
		portfolios[i] = cli.Portfolio{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			TotalValue:  p.TotalValue,
		}
	}

	// Run interactive selector
	selected, err := cli.RunPortfolioSelector(portfolios)
	if err != nil {
		return err
	}

	if selected == nil {
		cli.PrintInfo("No portfolio selected")
		return nil
	}

	// Show selected portfolio details
	cli.PrintSuccess(fmt.Sprintf("Selected: %s", selected.Name))
	fmt.Println()

	// Fetch full portfolio details
	var portfolio struct {
		ID          uint    `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		TotalValue  float64 `json:"total_value"`
		TotalGain   float64 `json:"total_gain"`
		TotalReturn float64 `json:"total_return"`
		CostBasis   float64 `json:"cost_basis"`
	}

	if err := client.Request("GET", fmt.Sprintf("/api/v1/portfolios/%d", selected.ID), nil, &portfolio); err != nil {
		return err
	}

	fmt.Println(cli.RenderSection("Portfolio Details"))
	fmt.Println()
	fmt.Println(cli.RenderKeyValue("ID", fmt.Sprintf("%d", portfolio.ID)))
	fmt.Println(cli.RenderKeyValue("Name", portfolio.Name))
	if portfolio.Description != "" {
		fmt.Println(cli.RenderKeyValue("Description", portfolio.Description))
	}
	fmt.Println()
	fmt.Println(cli.RenderSection("Performance"))
	fmt.Println()
	fmt.Println(cli.RenderKeyValue("Total Value", fmt.Sprintf("$%.2f", portfolio.TotalValue)))
	fmt.Println(cli.RenderKeyValue("Cost Basis", fmt.Sprintf("$%.2f", portfolio.CostBasis)))
	fmt.Println(cli.RenderKeyValue("Total Gain", formatGain(portfolio.TotalGain)))
	fmt.Println(cli.RenderKeyValue("Total Return", fmt.Sprintf("%.2f%%", portfolio.TotalReturn)))

	return nil
}
