package cmd

import (
	"fmt"

	"github.com/lenon/portfolios/internal/cli"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  "View and modify CLI configuration settings",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  "Display all current configuration settings",
	RunE:  runConfigShow,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long:  "Set a configuration value (api_base_url, output_format)",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config file path",
	Long:  "Display the path to the configuration file",
	RunE:  runConfigPath,
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configPathCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	fmt.Println(cli.RenderSection("Configuration"))
	fmt.Println()
	fmt.Println(cli.RenderKeyValue("API Base URL", config.APIBaseURL))
	fmt.Println(cli.RenderKeyValue("Output Format", config.OutputFormat))

	if config.AccessToken != "" {
		fmt.Println(cli.RenderKeyValue("Logged In", "Yes"))
	} else {
		fmt.Println(cli.RenderKeyValue("Logged In", "No"))
	}

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	config, err := cli.LoadConfig()
	if err != nil {
		return err
	}

	switch key {
	case "api_base_url":
		config.APIBaseURL = value
	case "output_format":
		if value != "table" && value != "json" && value != "csv" {
			return fmt.Errorf("invalid output format: %s (must be table, json, or csv)", value)
		}
		config.OutputFormat = value
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	if err := cli.SaveConfig(config); err != nil {
		return err
	}

	cli.PrintSuccess(fmt.Sprintf("Configuration updated: %s = %s", key, value))
	return nil
}

func runConfigPath(cmd *cobra.Command, args []string) error {
	path, err := cli.GetConfigPath()
	if err != nil {
		return err
	}

	fmt.Println(path)
	return nil
}
