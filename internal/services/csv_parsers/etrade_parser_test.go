package csv_parsers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lenon/portfolios/internal/dto"
)

func TestETradeParser_GetFormat(t *testing.T) {
	parser := NewETradeParser()
	assert.Equal(t, dto.ImportFormatETrade, parser.GetFormat())
}

func TestETradeParser_ValidateHeaders(t *testing.T) {
	parser := &ETradeParser{}

	tests := []struct {
		name    string
		headers []string
		wantErr bool
	}{
		{
			name:    "valid E*TRADE headers",
			headers: []string{"TransactionDate", "TransactionType", "Symbol", "Quantity"},
			wantErr: false,
		},
		{
			name:    "missing TransactionType",
			headers: []string{"TransactionDate", "Symbol", "Quantity"},
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

func TestETradeParser_Parse_Success(t *testing.T) {
	parser := NewETradeParser()

	csvData := `TransactionDate,TransactionType,Symbol,Quantity,Price,Commission
01/15/2024,Bought,AAPL,100,150.50,4.95
01/16/2024,Sold,GOOGL,50,2800.00,9.95`

	reader := strings.NewReader(csvData)
	transactions, errors, err := parser.Parse(reader)

	require.NoError(t, err)
	assert.Empty(t, errors)
	require.Len(t, transactions, 2)

	tx1 := transactions[0]
	assert.Equal(t, "BUY", string(tx1.Type))
	assert.Equal(t, "AAPL", tx1.Symbol)

	tx2 := transactions[1]
	assert.Equal(t, "SELL", string(tx2.Type))
	assert.Equal(t, "GOOGL", tx2.Symbol)
}
