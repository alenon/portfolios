package csv_parsers

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseParser_ParseDate(t *testing.T) {
	parser := &BaseParser{}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"ISO 8601", "2024-01-15", false},
		{"US format", "01/15/2024", false},
		{"US short", "01/15/24", false},
		{"ISO with time", "2024-01-15 10:30:45", false},
		{"US with time", "01/15/2024 10:30:45", false},
		{"Long format", "January 15, 2024", false},
		{"Short month", "Jan 15, 2024", false},
		{"ISO with slashes", "2024/01/15", false},
		{"Invalid date", "not-a-date", true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseDate(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, 2024, result.Year())
				assert.Equal(t, time.January, result.Month())
				assert.Equal(t, 15, result.Day())
				assert.Equal(t, time.UTC, result.Location())
			}
		})
	}
}

func TestBaseParser_ParseDecimal(t *testing.T) {
	parser := &BaseParser{}

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"simple integer", "100", "100", false},
		{"simple decimal", "150.50", "150.5", false},
		{"with dollar sign", "$150.50", "150.5", false},
		{"with euro sign", "€150.50", "150.5", false},
		{"with pound sign", "£150.50", "150.5", false},
		{"with thousands separator", "1,234.56", "1234.56", false},
		{"negative", "-100", "-100", false},
		{"accounting format", "(100)", "-100", false},
		{"empty string", "", "0", false},
		{"dash only", "-", "0", false},
		{"invalid", "abc", "0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseDecimal(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result.String())
			}
		})
	}
}

func TestBaseParser_ParseTransactionType(t *testing.T) {
	parser := &BaseParser{}

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"buy", "BUY", "BUY", false},
		{"sell", "SELL", "SELL", false},
		{"dividend", "DIVIDEND", "DIVIDEND", false},
		{"buy lowercase", "buy", "BUY", false},
		{"purchase", "PURCHASE", "BUY", false},
		{"sale", "SALE", "SELL", false},
		{"invalid", "INVALID_TYPE", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseTransactionType(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, string(result))
			}
		})
	}
}

func TestBaseParser_NormalizeSymbol(t *testing.T) {
	parser := &BaseParser{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "AAPL", "AAPL"},
		{"lowercase", "aapl", "AAPL"},
		{"with spaces", " AAPL ", "AAPL"},
		{"NASDAQ suffix", "AAPL.O", "AAPL"},
		{"NYSE suffix", "AAPL.N", "AAPL"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.NormalizeSymbol(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBaseParser_GetColumnIndex(t *testing.T) {
	parser := &BaseParser{}

	headers := []string{"Date", "Type", "Symbol", "Quantity"}

	tests := []struct {
		name          string
		columnNames   []string
		expectedIndex int
	}{
		{"exact match", []string{"Date"}, 0},
		{"case insensitive", []string{"date"}, 0},
		{"alternate name first", []string{"Trade Date", "Date"}, 0},
		{"alternate name second", []string{"Transaction Date", "Date"}, 0},
		{"not found", []string{"Price"}, -1},
		{"with spaces", []string{" Symbol "}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.GetColumnIndex(headers, tt.columnNames...)
			assert.Equal(t, tt.expectedIndex, result)
		})
	}
}

func TestBaseParser_GetColumnValue(t *testing.T) {
	parser := &BaseParser{}

	row := []string{"2024-01-15", "BUY", "AAPL", "100"}

	tests := []struct {
		name     string
		index    int
		expected string
	}{
		{"valid index 0", 0, "2024-01-15"},
		{"valid index 2", 2, "AAPL"},
		{"negative index", -1, ""},
		{"index out of bounds", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.GetColumnValue(row, tt.index)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBaseParser_IsEmptyRow(t *testing.T) {
	parser := &BaseParser{}

	tests := []struct {
		name     string
		row      []string
		expected bool
	}{
		{"all empty", []string{"", "", ""}, true},
		{"all spaces", []string{" ", "  ", "   "}, true},
		{"one non-empty", []string{"", "value", ""}, false},
		{"all non-empty", []string{"a", "b", "c"}, false},
		{"empty slice", []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.IsEmptyRow(tt.row)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBaseParser_JoinRow(t *testing.T) {
	parser := &BaseParser{}

	row := []string{"2024-01-15", "BUY", "AAPL", "100"}
	result := parser.JoinRow(row)
	assert.Equal(t, "2024-01-15,BUY,AAPL,100", result)
}

func TestBaseParser_ParseCSV(t *testing.T) {
	parser := &BaseParser{}

	tests := []struct {
		name     string
		csvData  string
		wantRows int
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid CSV",
			csvData: `Date,Type,Symbol
2024-01-15,BUY,AAPL`,
			wantRows: 2,
			wantErr:  false,
		},
		{
			name:     "empty CSV",
			csvData:  "",
			wantRows: 0,
			wantErr:  true,
			errMsg:   "empty",
		},
		{
			name: "with leading spaces",
			csvData: `Date,Type,Symbol
  2024-01-15,  BUY,  AAPL`,
			wantRows: 2,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.csvData)
			rows, err := parser.ParseCSV(reader)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, rows, tt.wantRows)
			}
		})
	}
}

func TestBaseParser_ParseInt(t *testing.T) {
	parser := &BaseParser{}

	tests := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{"simple integer", "123", 123, false},
		{"with comma", "1,234", 1234, false},
		{"negative", "-100", -100, false},
		{"empty", "", 0, false},
		{"invalid", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseInt(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
