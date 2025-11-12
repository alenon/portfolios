package csv_parsers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lenon/portfolios/internal/dto"
)

func TestInteractiveBrokersParser_GetFormat(t *testing.T) {
	parser := NewInteractiveBrokersParser()
	assert.Equal(t, dto.ImportFormatInteractiveBrokers, parser.GetFormat())
}

func TestInteractiveBrokersParser_ValidateHeaders(t *testing.T) {
	parser := &InteractiveBrokersParser{}

	tests := []struct {
		name    string
		headers []string
		wantErr bool
	}{
		{
			name:    "valid IB headers",
			headers: []string{"DataDiscriminator", "Code", "Symbol", "Quantity", "Date/Time"},
			wantErr: false,
		},
		{
			name:    "missing Date/Time",
			headers: []string{"DataDiscriminator", "Code", "Symbol", "Quantity"},
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

func TestInteractiveBrokersParser_Parse_Success(t *testing.T) {
	parser := NewInteractiveBrokersParser()

	csvData := `DataDiscriminator,Code,Symbol,Quantity,Price,Commission,Date/Time
Trade,O,AAPL,100,150.50,1.00,2024-01-15
Trade,C,GOOGL,50,2800.00,1.00,2024-01-16`

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

func TestInteractiveBrokersParser_Parse_Codes(t *testing.T) {
	tests := []struct {
		code         string
		expectedType string
	}{
		{"O", "BUY"},
		{"C", "SELL"},
		{"DIV", "DIVIDEND"},
		{"PL", "SPLIT"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			parser := NewInteractiveBrokersParser()
			csvData := "DataDiscriminator,Code,Symbol,Quantity,Price,Date/Time\nTrade," + tt.code + ",AAPL,100,150.00,2024-01-15"
			reader := strings.NewReader(csvData)
			transactions, errors, err := parser.Parse(reader)

			require.NoError(t, err)
			assert.Empty(t, errors)
			require.Len(t, transactions, 1)
			assert.Equal(t, tt.expectedType, string(transactions[0].Type))
		})
	}
}
