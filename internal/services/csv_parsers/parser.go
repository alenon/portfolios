package csv_parsers

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/models"
)

// CSVParser defines the interface that all CSV parsers must implement
type CSVParser interface {
	// Parse parses CSV data and returns import transaction requests
	Parse(data io.Reader) ([]dto.ImportTransactionRequest, []dto.ImportError, error)

	// GetFormat returns the format this parser handles
	GetFormat() dto.ImportFormat

	// ValidateHeaders validates that the CSV has the expected headers
	ValidateHeaders(headers []string) error
}

// BaseParser provides common functionality for all parsers
type BaseParser struct{}

// ParseCSV reads CSV data and returns all rows
func (p *BaseParser) ParseCSV(data io.Reader) ([][]string, error) {
	reader := csv.NewReader(data)
	reader.TrimLeadingSpace = true

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	return rows, nil
}

// ParseDate attempts to parse a date string using common date formats
func (p *BaseParser) ParseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)

	// Common date formats used by brokers
	formats := []string{
		"2006-01-02",                    // ISO 8601 (YYYY-MM-DD)
		"01/02/2006",                    // US format (MM/DD/YYYY)
		"01/02/06",                      // US short (MM/DD/YY)
		"02/01/2006",                    // UK format (DD/MM/YYYY)
		"2006-01-02 15:04:05",           // ISO with time
		"01/02/2006 15:04:05",           // US with time
		"01/02/2006 3:04:05 PM",         // US with 12-hour time
		"January 2, 2006",               // Long format
		"Jan 2, 2006",                   // Short month format
		"2006/01/02",                    // ISO with slashes
		time.RFC3339,                    // RFC3339
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			// Normalize to UTC at start of day
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// ParseDecimal parses a decimal value from a string
func (p *BaseParser) ParseDecimal(value string) (decimal.Decimal, error) {
	value = strings.TrimSpace(value)

	// Remove common currency symbols and formatting
	value = strings.ReplaceAll(value, "$", "")
	value = strings.ReplaceAll(value, "€", "")
	value = strings.ReplaceAll(value, "£", "")
	value = strings.ReplaceAll(value, ",", "")
	value = strings.TrimSpace(value)

	if value == "" || value == "-" {
		return decimal.Zero, nil
	}

	// Handle parentheses for negative numbers (accounting format)
	if strings.HasPrefix(value, "(") && strings.HasSuffix(value, ")") {
		value = "-" + strings.Trim(value, "()")
	}

	dec, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid decimal value: %s", value)
	}

	return dec, nil
}

// ParseTransactionType maps common transaction type strings to our internal types
func (p *BaseParser) ParseTransactionType(typeStr string) (models.TransactionType, error) {
	typeStr = strings.ToUpper(strings.TrimSpace(typeStr))

	// Map common variations to our transaction types
	typeMap := map[string]models.TransactionType{
		"BUY":                models.TransactionTypeBuy,
		"BOUGHT":             models.TransactionTypeBuy,
		"PURCHASE":           models.TransactionTypeBuy,
		"SELL":               models.TransactionTypeSell,
		"SOLD":               models.TransactionTypeSell,
		"SALE":               models.TransactionTypeSell,
		"DIVIDEND":           models.TransactionTypeDividend,
		"DIV":                models.TransactionTypeDividend,
		"CASH DIVIDEND":      models.TransactionTypeDividend,
		"SPLIT":              models.TransactionTypeSplit,
		"STOCK SPLIT":        models.TransactionTypeSplit,
		"MERGER":             models.TransactionTypeMerger,
		"SPINOFF":            models.TransactionTypeSpinoff,
		"SPIN-OFF":           models.TransactionTypeSpinoff,
		"DIVIDEND REINVEST":  models.TransactionTypeDividendReinvest,
		"DRIP":               models.TransactionTypeDividendReinvest,
		"REINVEST":           models.TransactionTypeDividendReinvest,
		"TICKER CHANGE":      models.TransactionTypeTickerChange,
		"SYMBOL CHANGE":      models.TransactionTypeTickerChange,
	}

	if txType, ok := typeMap[typeStr]; ok {
		return txType, nil
	}

	return "", fmt.Errorf("unknown transaction type: %s", typeStr)
}

// NormalizeSymbol normalizes a stock symbol
func (p *BaseParser) NormalizeSymbol(symbol string) string {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	// Remove common suffixes that might be in broker exports
	symbol = strings.TrimSuffix(symbol, ".O")  // NASDAQ suffix
	symbol = strings.TrimSuffix(symbol, ".N")  // NYSE suffix
	return symbol
}

// GetColumnIndex finds the index of a column by name (case-insensitive)
func (p *BaseParser) GetColumnIndex(headers []string, columnNames ...string) int {
	for _, name := range columnNames {
		for i, header := range headers {
			if strings.EqualFold(strings.TrimSpace(header), strings.TrimSpace(name)) {
				return i
			}
		}
	}
	return -1
}

// GetColumnValue safely gets a column value by index
func (p *BaseParser) GetColumnValue(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

// CreateImportError creates a standardized import error
func (p *BaseParser) CreateImportError(line int, field, message, rawData string) dto.ImportError {
	return dto.ImportError{
		Line:    line,
		Field:   field,
		Message: message,
		RawData: rawData,
	}
}

// ValidateTransaction performs basic validation on an import transaction
func (p *BaseParser) ValidateTransaction(tx *dto.ImportTransactionRequest, lineNum int) []dto.ImportError {
	var errors []dto.ImportError

	if tx.Symbol == "" {
		errors = append(errors, dto.ImportError{
			Line:    lineNum,
			Field:   "symbol",
			Message: "symbol is required",
			RawData: tx.RawData,
		})
	}

	if tx.Date.IsZero() {
		errors = append(errors, dto.ImportError{
			Line:    lineNum,
			Field:   "date",
			Message: "date is required",
			RawData: tx.RawData,
		})
	}

	if tx.Quantity.IsZero() {
		errors = append(errors, dto.ImportError{
			Line:    lineNum,
			Field:   "quantity",
			Message: "quantity must be greater than zero",
			RawData: tx.RawData,
		})
	}

	if tx.Quantity.IsNegative() {
		errors = append(errors, dto.ImportError{
			Line:    lineNum,
			Field:   "quantity",
			Message: "quantity cannot be negative",
			RawData: tx.RawData,
		})
	}

	// Price is required for BUY and SELL transactions
	if (tx.Type == models.TransactionTypeBuy || tx.Type == models.TransactionTypeSell) {
		if tx.Price == nil || tx.Price.IsZero() {
			errors = append(errors, dto.ImportError{
				Line:    lineNum,
				Field:   "price",
				Message: "price is required for buy/sell transactions",
				RawData: tx.RawData,
			})
		}
	}

	if tx.Commission.IsNegative() {
		errors = append(errors, dto.ImportError{
			Line:    lineNum,
			Field:   "commission",
			Message: "commission cannot be negative",
			RawData: tx.RawData,
		})
	}

	return errors
}

// ParseInt parses an integer from a string
func (p *BaseParser) ParseInt(value string) (int, error) {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, ",", "")

	if value == "" {
		return 0, nil
	}

	return strconv.Atoi(value)
}

// IsEmptyRow checks if a CSV row is empty or contains only whitespace
func (p *BaseParser) IsEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

// JoinRow joins a CSV row back into a string for error reporting
func (p *BaseParser) JoinRow(row []string) string {
	return strings.Join(row, ",")
}
