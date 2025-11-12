package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// OutputFormat represents the output format type
type OutputFormat string

const (
	OutputFormatTable OutputFormat = "table"
	OutputFormatJSON  OutputFormat = "json"
	OutputFormatCSV   OutputFormat = "csv"
)

// OutputTable renders data as a formatted table
func OutputTable(headers []string, rows [][]string) {
	if len(rows) == 0 {
		fmt.Println("No data to display")
		return
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		Headers(headers...).
		Rows(rows...)

	fmt.Println(t)
}

// OutputJSON renders data as formatted JSON
func OutputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// OutputCSV renders data as CSV
func OutputCSV(headers []string, rows [][]string) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// Output renders data in the specified format
func Output(format OutputFormat, headers []string, rows [][]string, jsonData interface{}) error {
	switch format {
	case OutputFormatTable:
		OutputTable(headers, rows)
		return nil
	case OutputFormatJSON:
		return OutputJSON(jsonData)
	case OutputFormatCSV:
		return OutputCSV(headers, rows)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)
	fmt.Println(style.Render("✓ " + message))
}

// PrintError prints an error message
func PrintError(message string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Bold(true)
	fmt.Fprintln(os.Stderr, style.Render("✗ "+message))
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Bold(true)
	fmt.Println(style.Render("⚠ " + message))
}

// PrintInfo prints an info message
func PrintInfo(message string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true)
	fmt.Println(style.Render("ℹ " + message))
}

// RenderKeyValue renders a key-value pair with styling
func RenderKeyValue(key, value string) string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true)
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))

	return keyStyle.Render(key+":") + " " + valueStyle.Render(value)
}

// RenderSection renders a section header
func RenderSection(title string) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("13")).
		Bold(true).
		Underline(true)
	return style.Render(title)
}

// Confirm prompts the user for confirmation
func Confirm(message string) bool {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Bold(true)

	fmt.Print(style.Render(message + " (y/N): "))

	var response string
	fmt.Scanln(&response)

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// ReadInput prompts for user input
func ReadInput(prompt string) (string, error) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12"))

	fmt.Print(style.Render(prompt + ": "))

	var input string
	if _, err := fmt.Scanln(&input); err != nil && err != io.EOF {
		return "", err
	}

	return strings.TrimSpace(input), nil
}
