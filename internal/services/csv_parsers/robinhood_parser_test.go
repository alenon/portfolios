package csv_parsers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lenon/portfolios/internal/dto"
)

func TestRobinhoodParser_GetFormat(t *testing.T) {
	parser := NewRobinhoodParser()
	assert.Equal(t, dto.ImportFormatRobinhood, parser.GetFormat())
}

func TestRobinhoodParser_ValidateHeaders(t *testing.T) {
	parser := &RobinhoodParser{}

	tests := []struct {
		name    string
		headers []string
		wantErr bool
	}{
		{
			name:    "valid Robinhood headers",
			headers: []string{"Activity Date", "Trans Code", "Instrument", "Quantity"},
			wantErr: false,
		},
		{
			name:    "missing Trans Code",
			headers: []string{"Activity Date", "Instrument", "Quantity"},
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

func TestRobinhoodParser_Parse_Success(t *testing.T) {
	parser := NewRobinhoodParser()

	csvData := `Activity Date,Trans Code,Instrument,Quantity,Price
01/15/2024,BUY,AAPL,100,150.50
01/16/2024,SELL,GOOGL,50,2800.00`

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

func TestRobinhoodParser_Parse_TransCodes(t *testing.T) {
	tests := []struct {
		transCode    string
		expectedType string
	}{
		{"BUY", "BUY"},
		{"SELL", "SELL"},
		{"CDIV", "DIVIDEND"},
		{"SPLIT", "SPLIT"},
	}

	for _, tt := range tests {
		t.Run(tt.transCode, func(t *testing.T) {
			parser := NewRobinhoodParser()
			csvData := "Activity Date,Trans Code,Instrument,Quantity,Price\n01/15/2024," + tt.transCode + ",AAPL,100,150.00"
			reader := strings.NewReader(csvData)
			transactions, errors, err := parser.Parse(reader)

			require.NoError(t, err)
			assert.Empty(t, errors)
			require.Len(t, transactions, 1)
			assert.Equal(t, tt.expectedType, string(transactions[0].Type))
		})
	}
}
