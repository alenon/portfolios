package csv_parsers

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lenon/portfolios/internal/dto"
)

func TestGenericParser_GetFormat(t *testing.T) {
	parser := NewGenericParser()
	assert.Equal(t, dto.ImportFormatGeneric, parser.GetFormat())
}

func TestGenericParser_ValidateHeaders(t *testing.T) {
	parser := &GenericParser{}

	tests := []struct {
		name    string
		headers []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid headers - all required present",
			headers: []string{"Date", "Type", "Symbol", "Quantity", "Price", "Commission"},
			wantErr: false,
		},
		{
			name:    "valid headers - case insensitive",
			headers: []string{"date", "type", "symbol", "quantity"},
			wantErr: false,
		},
		{
			name:    "valid headers - with optional fields",
			headers: []string{"date", "type", "symbol", "quantity", "price", "commission"},
			wantErr: false,
		},
		{
			name:    "missing date column",
			headers: []string{"Type", "Symbol", "Quantity"},
			wantErr: true,
			errMsg:  "missing required column: date",
		},
		{
			name:    "missing type column",
			headers: []string{"Date", "Symbol", "Quantity"},
			wantErr: true,
			errMsg:  "missing required column: type",
		},
		{
			name:    "missing symbol column",
			headers: []string{"Date", "Type", "Quantity"},
			wantErr: true,
			errMsg:  "missing required column: symbol",
		},
		{
			name:    "missing quantity column",
			headers: []string{"Date", "Type", "Symbol"},
			wantErr: true,
			errMsg:  "missing required column: quantity",
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

func TestGenericParser_Parse_Success(t *testing.T) {
	parser := NewGenericParser()

	csvData := `Date,Type,Symbol,Quantity,Price,Commission,Currency,Notes
2024-01-15,BUY,AAPL,100,150.50,9.99,USD,Test purchase
2024-01-16,SELL,GOOGL,50,2800.00,12.50,USD,Test sale`

	reader := strings.NewReader(csvData)
	transactions, errors, err := parser.Parse(reader)

	require.NoError(t, err)
	assert.Empty(t, errors)
	require.Len(t, transactions, 2)

	// Verify first transaction
	tx1 := transactions[0]
	assert.Equal(t, "BUY", string(tx1.Type))
	assert.Equal(t, "AAPL", tx1.Symbol)
	assert.Equal(t, "100", tx1.Quantity.String())
	assert.Equal(t, "150.5", tx1.Price.String())
	assert.Equal(t, "9.99", tx1.Commission.String())
	assert.Equal(t, "USD", tx1.Currency)
	assert.Equal(t, "Test purchase", tx1.Notes)
	assert.Equal(t, 2024, tx1.Date.Year())
	assert.Equal(t, time.January, tx1.Date.Month())
	assert.Equal(t, 15, tx1.Date.Day())

	// Verify second transaction
	tx2 := transactions[1]
	assert.Equal(t, "SELL", string(tx2.Type))
	assert.Equal(t, "GOOGL", tx2.Symbol)
	assert.Equal(t, "50", tx2.Quantity.String())
}

func TestGenericParser_Parse_WithPriceAndCommission(t *testing.T) {
	parser := NewGenericParser()

	csvData := `date,type,symbol,quantity,price,commission
2024-01-15,BUY,AAPL,100,150.50,9.99`

	reader := strings.NewReader(csvData)
	transactions, errors, err := parser.Parse(reader)

	require.NoError(t, err)
	assert.Empty(t, errors)
	require.Len(t, transactions, 1)

	tx := transactions[0]
	assert.Equal(t, "BUY", string(tx.Type))
	assert.Equal(t, "AAPL", tx.Symbol)
	assert.Equal(t, "100", tx.Quantity.String())
}

func TestGenericParser_Parse_EmptyRows(t *testing.T) {
	parser := NewGenericParser()

	// Empty rows should be skipped if all fields are empty
	csvData := `Date,Type,Symbol,Quantity,Price
2024-01-15,BUY,AAPL,100,150.50
2024-01-16,SELL,GOOGL,50,2800.00`

	reader := strings.NewReader(csvData)
	transactions, errors, err := parser.Parse(reader)

	require.NoError(t, err)
	assert.Empty(t, errors)
	require.Len(t, transactions, 2)
}

func TestGenericParser_Parse_OptionalFields(t *testing.T) {
	parser := NewGenericParser()

	// CSV without optional commission/notes but with required price for BUY
	csvData := `Date,Type,Symbol,Quantity,Price
2024-01-15,BUY,AAPL,100,150.50`

	reader := strings.NewReader(csvData)
	transactions, errors, err := parser.Parse(reader)

	require.NoError(t, err)
	assert.Empty(t, errors)
	require.Len(t, transactions, 1)

	tx := transactions[0]
	assert.Equal(t, "BUY", string(tx.Type))
	assert.Equal(t, "AAPL", tx.Symbol)
	assert.Equal(t, "100", tx.Quantity.String())
	assert.Equal(t, "150.5", tx.Price.String())
	assert.True(t, tx.Commission.IsZero())
	assert.Equal(t, "USD", tx.Currency) // Default
	assert.Equal(t, "", tx.Notes)
}

func TestGenericParser_Parse_Errors(t *testing.T) {
	tests := []struct {
		name      string
		csvData   string
		wantErr   bool
		errInData bool
		errCount  int
		errField  string
	}{
		{
			name:    "missing header row",
			csvData: ``,
			wantErr: true,
		},
		{
			name: "missing required header",
			csvData: `Date,Symbol,Quantity
2024-01-15,AAPL,100`,
			wantErr: true,
		},
		{
			name: "invalid date format",
			csvData: `Date,Type,Symbol,Quantity
invalid-date,BUY,AAPL,100`,
			errInData: true,
			errCount:  1,
			errField:  "date",
		},
		{
			name: "invalid transaction type",
			csvData: `Date,Type,Symbol,Quantity
2024-01-15,INVALID,AAPL,100`,
			errInData: true,
			errCount:  1,
			errField:  "type",
		},
		{
			name: "missing symbol",
			csvData: `Date,Type,Symbol,Quantity
2024-01-15,BUY,,100`,
			errInData: true,
			errCount:  1,
			errField:  "symbol",
		},
		{
			name: "invalid quantity",
			csvData: `Date,Type,Symbol,Quantity
2024-01-15,BUY,AAPL,invalid`,
			errInData: true,
			errCount:  1,
			errField:  "quantity",
		},
		{
			name: "invalid price",
			csvData: `Date,Type,Symbol,Quantity,Price
2024-01-15,BUY,AAPL,100,invalid`,
			errInData: true,
			errCount:  1,
			errField:  "price",
		},
		{
			name: "invalid commission",
			csvData: `Date,Type,Symbol,Quantity,Price,Commission
2024-01-15,BUY,AAPL,100,150.50,invalid`,
			errInData: true,
			errCount:  1,
			errField:  "commission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewGenericParser()
			reader := strings.NewReader(tt.csvData)
			transactions, importErrors, err := parser.Parse(reader)

			if tt.wantErr {
				require.Error(t, err)
			} else if tt.errInData {
				require.NoError(t, err)
				assert.Len(t, importErrors, tt.errCount)
				if tt.errCount > 0 {
					assert.Equal(t, tt.errField, importErrors[0].Field)
				}
				assert.Empty(t, transactions)
			} else {
				require.NoError(t, err)
				assert.Empty(t, importErrors)
			}
		})
	}
}

func TestGenericParser_Parse_NegativeQuantities(t *testing.T) {
	parser := NewGenericParser()

	csvData := `Date,Type,Symbol,Quantity,Price
2024-01-15,SELL,AAPL,-100,150.50`

	reader := strings.NewReader(csvData)
	transactions, importErrors, err := parser.Parse(reader)

	// Negative quantities should produce validation error
	require.NoError(t, err)
	assert.NotEmpty(t, importErrors)
	assert.Equal(t, "quantity", importErrors[0].Field)
	assert.Contains(t, importErrors[0].Message, "negative")
	assert.Empty(t, transactions)
}

func TestGenericParser_Parse_MultipleCurrencies(t *testing.T) {
	parser := NewGenericParser()

	csvData := `Date,Type,Symbol,Quantity,Price,Commission,Currency
2024-01-15,BUY,AAPL,100,150.50,9.99,USD
2024-01-16,BUY,VOW3.DE,50,120.00,5.00,EUR
2024-01-17,BUY,SONY,30,9500,1000,JPY`

	reader := strings.NewReader(csvData)
	transactions, errors, err := parser.Parse(reader)

	require.NoError(t, err)
	assert.Empty(t, errors)
	require.Len(t, transactions, 3)

	assert.Equal(t, "USD", transactions[0].Currency)
	assert.Equal(t, "EUR", transactions[1].Currency)
	assert.Equal(t, "JPY", transactions[2].Currency)
}

func TestGenericParser_Parse_OnlyHeaderRow(t *testing.T) {
	parser := NewGenericParser()

	csvData := `Date,Type,Symbol,Quantity,Price`

	reader := strings.NewReader(csvData)
	_, _, err := parser.Parse(reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one data row")
}
