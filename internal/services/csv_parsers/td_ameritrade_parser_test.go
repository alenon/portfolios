package csv_parsers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lenon/portfolios/internal/dto"
)

func TestTDAmeritradeParser_GetFormat(t *testing.T) {
	parser := NewTDAmeritradeParser()
	assert.Equal(t, dto.ImportFormatTDAmeritrade, parser.GetFormat())
}

func TestTDAmeritradeParser_ValidateHeaders(t *testing.T) {
	parser := &TDAmeritradeParser{}

	tests := []struct {
		name    string
		headers []string
		wantErr bool
	}{
		{
			name:    "valid TD Ameritrade headers",
			headers: []string{"DATE", "DESCRIPTION", "SYMBOL", "QUANTITY"},
			wantErr: false,
		},
		{
			name:    "missing DESCRIPTION",
			headers: []string{"DATE", "SYMBOL", "QUANTITY"},
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

func TestTDAmeritradeParser_Parse_Success(t *testing.T) {
	parser := NewTDAmeritradeParser()

	csvData := `DATE,DESCRIPTION,SYMBOL,QUANTITY,PRICE,COMMISSION,REG FEE
01/15/2024,BOUGHT 100 AAPL,AAPL,100,150.50,4.95,0.01
01/16/2024,SOLD 50 GOOGL,GOOGL,50,2800.00,9.95,0.05`

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

func TestTDAmeritradeParser_Parse_Descriptions(t *testing.T) {
	tests := []struct {
		description  string
		expectedType string
	}{
		{"BOUGHT 100 AAPL", "BUY"},
		{"SOLD 50 AAPL", "SELL"},
		{"DIVIDEND AAPL", "DIVIDEND"},
		{"CASH DIVIDEND", "DIVIDEND"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			parser := NewTDAmeritradeParser()
			csvData := "DATE,DESCRIPTION,SYMBOL,QUANTITY,PRICE\n01/15/2024," + tt.description + ",AAPL,100,150.00"
			reader := strings.NewReader(csvData)
			transactions, errors, err := parser.Parse(reader)

			require.NoError(t, err)
			assert.Empty(t, errors)
			require.Len(t, transactions, 1)
			assert.Equal(t, tt.expectedType, string(transactions[0].Type))
		})
	}
}
