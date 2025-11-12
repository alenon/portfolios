package csv_parsers

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lenon/portfolios/internal/dto"
)

func TestFidelityParser_GetFormat(t *testing.T) {
	parser := NewFidelityParser()
	assert.Equal(t, dto.ImportFormatFidelity, parser.GetFormat())
}

func TestFidelityParser_ValidateHeaders(t *testing.T) {
	parser := &FidelityParser{}

	tests := []struct {
		name    string
		headers []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid Fidelity headers",
			headers: []string{"Run Date", "Action", "Symbol", "Quantity", "Price", "Commission"},
			wantErr: false,
		},
		{
			name:    "missing Action column",
			headers: []string{"Run Date", "Symbol", "Quantity"},
			wantErr: true,
			errMsg:  "missing Fidelity 'Action' column",
		},
		{
			name:    "missing Symbol column",
			headers: []string{"Run Date", "Action", "Quantity"},
			wantErr: true,
			errMsg:  "missing Fidelity 'Symbol' column",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateHeaders(tt.headers)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFidelityParser_Parse_Success(t *testing.T) {
	parser := NewFidelityParser()

	csvData := `Run Date,Action,Symbol,Security Description,Quantity,Price,Commission,Fees
01/15/2024,YOU BOUGHT,AAPL,Apple Inc,100,150.50,4.95,0.50
01/16/2024,YOU SOLD,GOOGL,Alphabet Inc,50,2800.00,9.95,1.00`

	reader := strings.NewReader(csvData)
	transactions, errors, err := parser.Parse(reader)

	require.NoError(t, err)
	assert.Empty(t, errors)
	require.Len(t, transactions, 2)

	// Verify BUY transaction
	tx1 := transactions[0]
	assert.Equal(t, "BUY", string(tx1.Type))
	assert.Equal(t, "AAPL", tx1.Symbol)
	assert.Equal(t, "100", tx1.Quantity.String())
	assert.Equal(t, "150.5", tx1.Price.String())
	// Commission + Fees = 4.95 + 0.50 = 5.45
	assert.Equal(t, "5.45", tx1.Commission.String())
	assert.Equal(t, "Apple Inc", tx1.Notes)

	// Verify SELL transaction
	tx2 := transactions[1]
	assert.Equal(t, "SELL", string(tx2.Type))
	assert.Equal(t, "GOOGL", tx2.Symbol)
	// Commission + Fees = 9.95 + 1.00 = 10.95
	assert.Equal(t, "10.95", tx2.Commission.String())
}

func TestFidelityParser_Parse_FidelityActions(t *testing.T) {
	tests := []struct {
		name         string
		action       string
		expectedType string
		quantity     string
		expectedQty  decimal.Decimal
	}{
		{
			name:         "YOU BOUGHT",
			action:       "YOU BOUGHT",
			expectedType: "BUY",
			quantity:     "100",
			expectedQty:  decimal.NewFromInt(100),
		},
		{
			name:         "YOU SOLD",
			action:       "YOU SOLD",
			expectedType: "SELL",
			quantity:     "50",
			expectedQty:  decimal.NewFromInt(50),
		},
		{
			name:         "DIVIDEND REINVESTMENT",
			action:       "DIVIDEND REINVESTMENT",
			expectedType: "DIVIDEND_REINVEST",
			quantity:     "10",
			expectedQty:  decimal.NewFromInt(10),
		},
		{
			name:         "STOCK SPLIT",
			action:       "STOCK SPLIT",
			expectedType: "SPLIT",
			quantity:     "100",
			expectedQty:  decimal.NewFromInt(100),
		},
		{
			name:         "MERGER",
			action:       "MERGER",
			expectedType: "MERGER",
			quantity:     "50",
			expectedQty:  decimal.NewFromInt(50),
		},
		{
			name:         "SPINOFF",
			action:       "SPINOFF",
			expectedType: "SPINOFF",
			quantity:     "25",
			expectedQty:  decimal.NewFromInt(25),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewFidelityParser()

			csvData := "Run Date,Action,Symbol,Quantity,Price\n" +
				"01/15/2024," + tt.action + ",AAPL," + tt.quantity + ",150.00"

			reader := strings.NewReader(csvData)
			transactions, errors, err := parser.Parse(reader)

			require.NoError(t, err)
			assert.Empty(t, errors)
			require.Len(t, transactions, 1)

			tx := transactions[0]
			assert.Equal(t, tt.expectedType, string(tx.Type))
			assert.Equal(t, tt.expectedQty.String(), tx.Quantity.String())
		})
	}
}

func TestFidelityParser_Parse_NegativeQuantity(t *testing.T) {
	parser := NewFidelityParser()

	// Fidelity sometimes uses negative quantities for sells
	csvData := `Run Date,Action,Symbol,Quantity,Price
01/15/2024,YOU BOUGHT,AAPL,-100,150.50`

	reader := strings.NewReader(csvData)
	transactions, errors, err := parser.Parse(reader)

	require.NoError(t, err)
	assert.Empty(t, errors)
	require.Len(t, transactions, 1)

	tx := transactions[0]
	// Negative quantity with BUY action should convert to SELL with positive quantity
	assert.Equal(t, "SELL", string(tx.Type))
	assert.Equal(t, "AAPL", tx.Symbol)
	assert.True(t, tx.Quantity.IsPositive())
	assert.Equal(t, "100", tx.Quantity.String())
}

func TestFidelityParser_Parse_CommissionAndFees(t *testing.T) {
	parser := NewFidelityParser()

	// Test that commission and fees are combined
	csvData := `Run Date,Action,Symbol,Quantity,Price,Commission,Fees
01/15/2024,YOU BOUGHT,AAPL,100,150.50,4.95,1.05`

	reader := strings.NewReader(csvData)
	transactions, errors, err := parser.Parse(reader)

	require.NoError(t, err)
	assert.Empty(t, errors)
	require.Len(t, transactions, 1)

	tx := transactions[0]
	// Commission + Fees should be combined: 4.95 + 1.05 = 6.00
	assert.Equal(t, "6", tx.Commission.String())
}

func TestFidelityParser_Parse_EmptyRows(t *testing.T) {
	parser := NewFidelityParser()

	csvData := `Run Date,Action,Symbol,Quantity,Price
01/15/2024,YOU BOUGHT,AAPL,100,150.50
01/16/2024,YOU SOLD,GOOGL,50,2800.00`

	reader := strings.NewReader(csvData)
	transactions, errors, err := parser.Parse(reader)

	require.NoError(t, err)
	assert.Empty(t, errors)
	require.Len(t, transactions, 2)
}

func TestFidelityParser_Parse_Errors(t *testing.T) {
	tests := []struct {
		name      string
		csvData   string
		wantErr   bool
		errInData bool
		errField  string
	}{
		{
			name: "invalid date",
			csvData: `Run Date,Action,Symbol,Quantity
invalid-date,YOU BOUGHT,AAPL,100`,
			errInData: true,
			errField:  "date",
		},
		{
			name: "invalid action",
			csvData: `Run Date,Action,Symbol,Quantity
01/15/2024,INVALID_ACTION,AAPL,100`,
			errInData: true,
			errField:  "action",
		},
		{
			name: "missing symbol",
			csvData: `Run Date,Action,Symbol,Quantity
01/15/2024,YOU BOUGHT,,100`,
			errInData: true,
			errField:  "symbol",
		},
		{
			name: "invalid quantity",
			csvData: `Run Date,Action,Symbol,Quantity
01/15/2024,YOU BOUGHT,AAPL,invalid`,
			errInData: true,
			errField:  "quantity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewFidelityParser()
			reader := strings.NewReader(tt.csvData)
			transactions, importErrors, err := parser.Parse(reader)

			if tt.wantErr {
				require.Error(t, err)
			} else if tt.errInData {
				require.NoError(t, err)
				assert.NotEmpty(t, importErrors)
				assert.Equal(t, tt.errField, importErrors[0].Field)
				assert.Empty(t, transactions)
			}
		})
	}
}

func TestFidelityParser_Parse_OnlyHeaderRow(t *testing.T) {
	parser := NewFidelityParser()

	csvData := `Run Date,Action,Symbol,Quantity`

	reader := strings.NewReader(csvData)
	_, _, err := parser.Parse(reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one data row")
}
