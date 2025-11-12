package cmd

import (
	"fmt"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	Version   = "0.1.0"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display version, build date, and other build information",
	Run:   runVersion,
}

func runVersion(cmd *cobra.Command, args []string) {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("13")).
		Bold(true)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))

	fmt.Println(titleStyle.Render("Portfolios CLI"))
	fmt.Println()
	fmt.Printf("%s %s\n", labelStyle.Render("Version:"), valueStyle.Render(Version))
	fmt.Printf("%s %s\n", labelStyle.Render("Build Date:"), valueStyle.Render(BuildDate))
	fmt.Printf("%s %s\n", labelStyle.Render("Git Commit:"), valueStyle.Render(GitCommit))
	fmt.Printf("%s %s\n", labelStyle.Render("Go Version:"), valueStyle.Render(runtime.Version()))
	fmt.Printf("%s %s/%s\n", labelStyle.Render("Platform:"), valueStyle.Render(runtime.GOOS), valueStyle.Render(runtime.GOARCH))
}
