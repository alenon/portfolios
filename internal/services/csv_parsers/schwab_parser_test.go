package csv_parsers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lenon/portfolios/internal/dto"
)

func TestSchwabParser_GetFormat(t *testing.T) {
	parser := NewSchwabParser()
	assert.Equal(t, dto.ImportFormatSchwab, parser.GetFormat())
}

func TestSchwabParser_ValidateHeaders(t *testing.T) {
	parser := &SchwabParser{}

	tests := []struct {
		name    string
		headers []string
		wantErr bool
	}{
		{
			name:    "valid Schwab headers",
			headers: []string{"Date", "Action", "Symbol", "Quantity"},
			wantErr: false,
		},
		{
			name:    "missing Action",
			headers: []string{"Date", "Symbol", "Quantity"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateHeaders(tt.headers)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSchwabParser_Parse_Success(t *testing.T) {
	parser := NewSchwabParser()

	csvData := `Date,Action,Symbol,Quantity,Price,Fees & Comm
01/15/2024,BUY,AAPL,100,150.50,4.95
01/16/2024,SELL,GOOGL,50,2800.00,9.95`

	reader := strings.NewReader(csvData)
	transactions, errors, err := parser.Parse(reader)

	require.NoError(t, err)
	assert.Empty(t, errors)
	require.Len(t, transactions, 2)

	tx1 := transactions[0]
	assert.Equal(t, "BUY", string(tx1.Type))
	assert.Equal(t, "AAPL", tx1.Symbol)
	assert.Equal(t, "100", tx1.Quantity.String())

	tx2 := transactions[1]
	assert.Equal(t, "SELL", string(tx2.Type))
	assert.Equal(t, "GOOGL", tx2.Symbol)
}

func TestSchwabParser_Parse_Actions(t *testing.T) {
	tests := []struct {
		action       string
		expectedType string
	}{
		{"BUY", "BUY"},
		{"SELL", "SELL"},
		{"REINVEST DIVIDEND", "DIVIDEND_REINVEST"},
		{"STOCK SPLIT", "SPLIT"},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			parser := NewSchwabParser()
			csvData := "Date,Action,Symbol,Quantity,Price\n01/15/2024," + tt.action + ",AAPL,100,150.00"
			reader := strings.NewReader(csvData)
			transactions, errors, err := parser.Parse(reader)

			require.NoError(t, err)
			assert.Empty(t, errors)
			require.Len(t, transactions, 1)
			assert.Equal(t, tt.expectedType, string(transactions[0].Type))
		})
	}
}
