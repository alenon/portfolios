package cmd

import (
	"fmt"

	"github.com/lenon/portfolios/internal/cli"
	"github.com/spf13/cobra"
)

var portfolioCmd = &cobra.Command{
	Use:     "portfolio",
	Aliases: []string{"port", "p"},
	Short:   "Manage portfolios",
	Long:    "Create, list, view, and delete investment portfolios",
}

var portfolioListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List all portfolios",
	Long:    "Display all portfolios with their current values and performance metrics",
	RunE:    runPortfolioList,
}

var portfolioCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new portfolio",
	Long:  "Create a new investment portfolio with a name and optional description",
	RunE:  runPortfolioCreate,
}

var portfolioShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show portfolio details",
	Long:  "Display detailed information about a specific portfolio",
	Args:  cobra.ExactArgs(1),
	RunE:  runPortfolioShow,
}

var portfolioDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a portfolio",
	Long:    "Delete a portfolio and all its associated data",
	Args:    cobra.ExactArgs(1),
	RunE:    runPortfolioDelete,
}

var portfolioHoldingsCmd = &cobra.Command{
	Use:     "holdings <id>",
	Aliases: []string{"hold"},
	Short:   "Show portfolio holdings",
	Long:    "Display current holdings for a specific portfolio",
	Args:    cobra.ExactArgs(1),
	RunE:    runPortfolioHoldings,
}

func init() {
	portfolioCmd.AddCommand(portfolioListCmd)
	portfolioCmd.AddCommand(portfolioCreateCmd)
	portfolioCmd.AddCommand(portfolioShowCmd)
	portfolioCmd.AddCommand(portfolioDeleteCmd)
	portfolioCmd.AddCommand(portfolioHoldingsCmd)
}

func runPortfolioList(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	client := cli.NewClientFromConfig(config)

	var portfolios []struct {
		ID          uint    `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		TotalValue  float64 `json:"total_value"`
		TotalGain   float64 `json:"total_gain"`
		TotalReturn float64 `json:"total_return"`
		CreatedAt   string  `json:"created_at"`
	}

	if err := client.Request("GET", "/api/v1/portfolios", nil, &portfolios); err != nil {
		return err
	}

	if len(portfolios) == 0 {
		cli.PrintInfo("No portfolios found. Create one with 'portfolios portfolio create'")
		return nil
	}

	format := cli.OutputFormat(config.OutputFormat)
	if outputFormat != "" {
		format = cli.OutputFormat(outputFormat)
	}

	headers := []string{"ID", "Name", "Description", "Total Value", "Total Gain", "Total Return %"}
	rows := make([][]string, len(portfolios))

	for i, p := range portfolios {
		rows[i] = []string{
			fmt.Sprintf("%d", p.ID),
			p.Name,
			truncate(p.Description, 40),
			fmt.Sprintf("$%.2f", p.TotalValue),
			formatGain(p.TotalGain),
			fmt.Sprintf("%.2f%%", p.TotalReturn),
		}
	}

	return cli.Output(format, headers, rows, portfolios)
}

func runPortfolioCreate(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	// Prompt for portfolio details
	name, err := cli.ReadInput("Portfolio Name")
	if err != nil {
		return err
	}

	description, err := cli.ReadInput("Description (optional)")
	if err != nil {
		return err
	}

	client := cli.NewClientFromConfig(config)

	createReq := map[string]string{
		"name":        name,
		"description": description,
	}

	var portfolio struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := client.Request("POST", "/api/v1/portfolios", createReq, &portfolio); err != nil {
		return err
	}

	cli.PrintSuccess(fmt.Sprintf("Portfolio '%s' created successfully!", portfolio.Name))
	fmt.Println()
	fmt.Println(cli.RenderKeyValue("ID", fmt.Sprintf("%d", portfolio.ID)))
	fmt.Println(cli.RenderKeyValue("Name", portfolio.Name))
	if portfolio.Description != "" {
		fmt.Println(cli.RenderKeyValue("Description", portfolio.Description))
	}

	return nil
}

func runPortfolioShow(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	portfolioID := args[0]
	client := cli.NewClientFromConfig(config)

	var portfolio struct {
		ID          uint    `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		TotalValue  float64 `json:"total_value"`
		TotalGain   float64 `json:"total_gain"`
		TotalReturn float64 `json:"total_return"`
		CostBasis   float64 `json:"cost_basis"`
		CreatedAt   string  `json:"created_at"`
		UpdatedAt   string  `json:"updated_at"`
	}

	if err := client.Request("GET", "/api/v1/portfolios/"+portfolioID, nil, &portfolio); err != nil {
		return err
	}

	format := cli.OutputFormat(config.OutputFormat)
	if outputFormat != "" {
		format = cli.OutputFormat(outputFormat)
	}

	if format == cli.OutputFormatJSON {
		return cli.OutputJSON(portfolio)
	}

	// Display as formatted text
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
	fmt.Println()
	fmt.Println(cli.RenderKeyValue("Created", portfolio.CreatedAt))
	fmt.Println(cli.RenderKeyValue("Last Updated", portfolio.UpdatedAt))

	return nil
}

func runPortfolioDelete(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	portfolioID := args[0]

	if !cli.Confirm(fmt.Sprintf("Are you sure you want to delete portfolio %s?", portfolioID)) {
		cli.PrintInfo("Deletion cancelled")
		return nil
	}

	client := cli.NewClientFromConfig(config)

	if err := client.Request("DELETE", "/api/v1/portfolios/"+portfolioID, nil, nil); err != nil {
		return err
	}

	cli.PrintSuccess("Portfolio deleted successfully!")
	return nil
}

func runPortfolioHoldings(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	portfolioID := args[0]
	client := cli.NewClientFromConfig(config)

	var holdings []struct {
		ID                    uint    `json:"id"`
		Symbol                string  `json:"symbol"`
		Quantity              float64 `json:"quantity"`
		AverageCost           float64 `json:"average_cost"`
		TotalCost             float64 `json:"total_cost"`
		CurrentValue          float64 `json:"current_value"`
		UnrealizedGain        float64 `json:"unrealized_gain"`
		UnrealizedGainPercent float64 `json:"unrealized_gain_percent"`
	}

	if err := client.Request("GET", "/api/v1/portfolios/"+portfolioID+"/holdings", nil, &holdings); err != nil {
		return err
	}

	if len(holdings) == 0 {
		cli.PrintInfo("No holdings found for this portfolio")
		return nil
	}

	format := cli.OutputFormat(config.OutputFormat)
	if outputFormat != "" {
		format = cli.OutputFormat(outputFormat)
	}

	headers := []string{"Symbol", "Quantity", "Avg Cost", "Total Cost", "Current Value", "Unrealized Gain", "Return %"}
	rows := make([][]string, len(holdings))

	for i, h := range holdings {
		rows[i] = []string{
			h.Symbol,
			fmt.Sprintf("%.4f", h.Quantity),
			fmt.Sprintf("$%.2f", h.AverageCost),
			fmt.Sprintf("$%.2f", h.TotalCost),
			fmt.Sprintf("$%.2f", h.CurrentValue),
			formatGain(h.UnrealizedGain),
			fmt.Sprintf("%.2f%%", h.UnrealizedGainPercent),
		}
	}

	return cli.Output(format, headers, rows, holdings)
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatGain(gain float64) string {
	if gain >= 0 {
		return fmt.Sprintf("+$%.2f", gain)
	}
	return fmt.Sprintf("-$%.2f", -gain)
}
