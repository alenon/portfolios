package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/lenon/portfolios/internal/cli"
	"github.com/spf13/cobra"
)

var (
	cfgFile      string
	outputFormat string
	apiBaseURL   string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "portfolios",
	Short: "Portfolio management CLI",
	Long: renderBanner() + `

A powerful command-line interface for managing your investment portfolios.

Features:
  • Track stocks, dividends, and corporate actions
  • Import transactions from major brokers
  • Analyze performance with advanced metrics
  • Generate reports and export data
  • Manage tax lots and cost basis

Get started by logging in:
  portfolios auth login

For more information, visit: https://github.com/lenon/portfolios`,
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.portfolios/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "output format (table|json|csv)")
	rootCmd.PersistentFlags().StringVar(&apiBaseURL, "api-url", "", "API base URL (default is http://localhost:8080)")

	// Add subcommands
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(portfolioCmd)
	rootCmd.AddCommand(transactionCmd)
	rootCmd.AddCommand(performanceCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	config, err := cli.LoadConfig()
	if err != nil {
		cli.PrintError(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	// Override config with flags if provided
	if apiBaseURL != "" {
		config.APIBaseURL = apiBaseURL
	}
	if outputFormat != "" {
		config.OutputFormat = outputFormat
	}
}

// renderBanner returns a styled banner for the CLI
func renderBanner() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("13")).
		Bold(true)

	banner := `
 ____            _    __       _ _
|  _ \ ___  _ __| |_ / _| ___ | (_) ___  ___
| |_) / _ \| '__| __| |_ / _ \| | |/ _ \/ __|
|  __/ (_) | |  | |_|  _| (_) | | | (_) \__ \
|_|   \___/|_|   \__|_|  \___/|_|_|\___/|___/`

	return style.Render(banner)
}
