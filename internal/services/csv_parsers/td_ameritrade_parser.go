package csv_parsers

import (
	"fmt"
	"io"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/models"
)

// TDAmeritradeParser handles TD Ameritrade CSV format imports
// TD Ameritrade CSV format typically includes:
// DATE,TRANSACTION ID,DESCRIPTION,QUANTITY,SYMBOL,PRICE,COMMISSION,AMOUNT,REG FEE,SHORT-TERM RDM FEE,FUND REDEMPTION FEE, DEFERRED SALES CHARGE
type TDAmeritradeParser struct {
	BaseParser
}

// NewTDAmeritradeParser creates a new TD Ameritrade CSV parser
func NewTDAmeritradeParser() CSVParser {
	return &TDAmeritradeParser{}
}

// GetFormat returns the format this parser handles
func (p *TDAmeritradeParser) GetFormat() dto.ImportFormat {
	return dto.ImportFormatTDAmeritrade
}

// ValidateHeaders validates that the CSV has the expected TD Ameritrade headers
func (p *TDAmeritradeParser) ValidateHeaders(headers []string) error {
	// Look for key TD Ameritrade columns
	if p.GetColumnIndex(headers, "DATE") == -1 {
		return fmt.Errorf("missing TD Ameritrade 'DATE' column")
	}
	if p.GetColumnIndex(headers, "DESCRIPTION") == -1 {
		return fmt.Errorf("missing TD Ameritrade 'DESCRIPTION' column")
	}
	return nil
}

// Parse parses TD Ameritrade CSV data and returns import transaction requests
func (p *TDAmeritradeParser) Parse(data io.Reader) ([]dto.ImportTransactionRequest, []dto.ImportError, error) {
	rows, err := p.ParseCSV(data)
	if err != nil {
		return nil, nil, err
	}

	if len(rows) < 2 {
		return nil, nil, fmt.Errorf("CSV must contain header row and at least one data row")
	}

	headers := rows[0]
	if err := p.ValidateHeaders(headers); err != nil {
		return nil, nil, err
	}

	// Get column indices for TD Ameritrade-specific headers
	dateIdx := p.GetColumnIndex(headers, "DATE")
	descIdx := p.GetColumnIndex(headers, "DESCRIPTION")
	symbolIdx := p.GetColumnIndex(headers, "SYMBOL")
	quantityIdx := p.GetColumnIndex(headers, "QUANTITY")
	priceIdx := p.GetColumnIndex(headers, "PRICE")
	commissionIdx := p.GetColumnIndex(headers, "COMMISSION")
	regFeeIdx := p.GetColumnIndex(headers, "REG FEE")

	var transactions []dto.ImportTransactionRequest
	var errors []dto.ImportError

	// Parse each data row
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		lineNum := i + 1

		// Skip empty rows
		if p.IsEmptyRow(row) {
			continue
		}

		rawData := p.JoinRow(row)

		// Parse date
		dateStr := p.GetColumnValue(row, dateIdx)
		date, err := p.ParseDate(dateStr)
		if err != nil {
			errors = append(errors, p.CreateImportError(lineNum, "date", err.Error(), rawData))
			continue
		}

		// Parse TD Ameritrade description to determine transaction type
		description := p.GetColumnValue(row, descIdx)
		txType, err := p.parseTDADescription(description)
		if err != nil {
			errors = append(errors, p.CreateImportError(lineNum, "description", err.Error(), rawData))
			continue
		}

		// Parse symbol
		symbol := p.NormalizeSymbol(p.GetColumnValue(row, symbolIdx))
		if symbol == "" {
			errors = append(errors, p.CreateImportError(lineNum, "symbol", "symbol is required", rawData))
			continue
		}

		// Parse quantity
		quantityStr := p.GetColumnValue(row, quantityIdx)
		quantity, err := p.ParseDecimal(quantityStr)
		if err != nil {
			errors = append(errors, p.CreateImportError(lineNum, "quantity", err.Error(), rawData))
			continue
		}

		// Quantity should always be positive
		quantity = quantity.Abs()

		// Parse price
		var price *decimal.Decimal
		if priceIdx >= 0 {
			priceStr := p.GetColumnValue(row, priceIdx)
			if priceStr != "" {
				priceVal, err := p.ParseDecimal(priceStr)
				if err != nil {
					errors = append(errors, p.CreateImportError(lineNum, "price", err.Error(), rawData))
					continue
				}
				priceVal = priceVal.Abs()
				price = &priceVal
			}
		}

		// Parse commission and regulatory fees
		commission := decimal.Zero
		if commissionIdx >= 0 {
			commissionStr := p.GetColumnValue(row, commissionIdx)
			if commissionStr != "" {
				commissionVal, err := p.ParseDecimal(commissionStr)
				if err == nil {
					commission = commission.Add(commissionVal.Abs())
				}
			}
		}
		if regFeeIdx >= 0 {
			regFeeStr := p.GetColumnValue(row, regFeeIdx)
			if regFeeStr != "" {
				regFeeVal, err := p.ParseDecimal(regFeeStr)
				if err == nil {
					commission = commission.Add(regFeeVal.Abs())
				}
			}
		}

		// Get description for notes
		notes := description

		// Create import transaction request
		tx := dto.ImportTransactionRequest{
			Type:       txType,
			Symbol:     symbol,
			Date:       date,
			Quantity:   quantity,
			Price:      price,
			Commission: commission,
			Currency:   "USD",
			Notes:      notes,
			RawData:    rawData,
		}

		// Validate transaction
		validationErrors := p.ValidateTransaction(&tx, lineNum)
		if len(validationErrors) > 0 {
			errors = append(errors, validationErrors...)
			continue
		}

		transactions = append(transactions, tx)
	}

	return transactions, errors, nil
}

// parseTDADescription maps TD Ameritrade descriptions to our transaction types
func (p *TDAmeritradeParser) parseTDADescription(description string) (models.TransactionType, error) {
	// TD Ameritrade-specific description mapping (descriptions contain action keywords)
	descMap := map[string]models.TransactionType{
		"BOUGHT":                models.TransactionTypeBuy,
		"BUY":                   models.TransactionTypeBuy,
		"SOLD":                  models.TransactionTypeSell,
		"SELL":                  models.TransactionTypeSell,
		"DIVIDEND":              models.TransactionTypeDividend,
		"CASH DIVIDEND":         models.TransactionTypeDividend,
		"QUALIFIED DIVIDEND":    models.TransactionTypeDividend,
		"ORDINARY DIVIDEND":     models.TransactionTypeDividend,
		"REINVEST":              models.TransactionTypeDividendReinvest,
		"DIVIDEND REINVESTMENT": models.TransactionTypeDividendReinvest,
		"STOCK SPLIT":           models.TransactionTypeSplit,
		"MERGER":                models.TransactionTypeMerger,
		"ACQUISITION":           models.TransactionTypeMerger,
		"SPINOFF":               models.TransactionTypeSpinoff,
		"SPIN OFF":              models.TransactionTypeSpinoff,
		"SYMBOL CHANGE":         models.TransactionTypeTickerChange,
		"NAME CHANGE":           models.TransactionTypeTickerChange,
	}

	// Try to find matching keyword in description
	for keyword, txType := range descMap {
		if contains(description, keyword) {
			return txType, nil
		}
	}

	// Try to use base parser for common variations
	return p.ParseTransactionType(description)
}

// contains checks if a string contains a substring (case-insensitive)
func contains(str, substr string) bool {
	str = strings.ToUpper(str)
	substr = strings.ToUpper(substr)
	return strings.Contains(str, substr)
}
