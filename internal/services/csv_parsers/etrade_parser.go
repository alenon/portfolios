package csv_parsers

import (
	"fmt"
	"io"

	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/models"
)

// ETradeParser handles E*TRADE CSV format imports
// E*TRADE CSV format typically includes:
// TransactionDate,TransactionType,SecurityType,Symbol,Quantity,Amount,Price,Commission,Description
type ETradeParser struct {
	BaseParser
}

// NewETradeParser creates a new E*TRADE CSV parser
func NewETradeParser() CSVParser {
	return &ETradeParser{}
}

// GetFormat returns the format this parser handles
func (p *ETradeParser) GetFormat() dto.ImportFormat {
	return dto.ImportFormatETrade
}

// ValidateHeaders validates that the CSV has the expected E*TRADE headers
func (p *ETradeParser) ValidateHeaders(headers []string) error {
	// Look for key E*TRADE columns
	if p.GetColumnIndex(headers, "TransactionDate", "Transaction Date") == -1 {
		return fmt.Errorf("missing E*TRADE 'TransactionDate' column")
	}
	if p.GetColumnIndex(headers, "TransactionType", "Transaction Type") == -1 {
		return fmt.Errorf("missing E*TRADE 'TransactionType' column")
	}
	return nil
}

// Parse parses E*TRADE CSV data and returns import transaction requests
func (p *ETradeParser) Parse(data io.Reader) ([]dto.ImportTransactionRequest, []dto.ImportError, error) {
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

	// Get column indices for E*TRADE-specific headers
	dateIdx := p.GetColumnIndex(headers, "TransactionDate", "Transaction Date")
	typeIdx := p.GetColumnIndex(headers, "TransactionType", "Transaction Type")
	symbolIdx := p.GetColumnIndex(headers, "Symbol")
	quantityIdx := p.GetColumnIndex(headers, "Quantity")
	priceIdx := p.GetColumnIndex(headers, "Price")
	commissionIdx := p.GetColumnIndex(headers, "Commission")
	descIdx := p.GetColumnIndex(headers, "Description")

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

		// Parse E*TRADE transaction type
		typeStr := p.GetColumnValue(row, typeIdx)
		txType, err := p.parseETradeType(typeStr)
		if err != nil {
			errors = append(errors, p.CreateImportError(lineNum, "type", err.Error(), rawData))
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

		// Parse commission
		commission := decimal.Zero
		if commissionIdx >= 0 {
			commissionStr := p.GetColumnValue(row, commissionIdx)
			if commissionStr != "" {
				commissionVal, err := p.ParseDecimal(commissionStr)
				if err == nil {
					commission = commissionVal.Abs()
				}
			}
		}

		// Get description for notes
		notes := ""
		if descIdx >= 0 {
			notes = p.GetColumnValue(row, descIdx)
		}

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

// parseETradeType maps E*TRADE transaction types to our transaction types
func (p *ETradeParser) parseETradeType(txType string) (models.TransactionType, error) {
	// E*TRADE-specific type mapping
	typeMap := map[string]models.TransactionType{
		"BOUGHT":                  models.TransactionTypeBuy,
		"SOLD":                    models.TransactionTypeSell,
		"DIVIDEND":                models.TransactionTypeDividend,
		"CASH DIVIDEND":           models.TransactionTypeDividend,
		"DIVIDEND REINVESTED":     models.TransactionTypeDividendReinvest,
		"REINVEST DIVIDEND":       models.TransactionTypeDividendReinvest,
		"STOCK SPLIT":             models.TransactionTypeSplit,
		"MERGER":                  models.TransactionTypeMerger,
		"SPINOFF":                 models.TransactionTypeSpinoff,
		"SYMBOL CHANGE":           models.TransactionTypeTickerChange,
	}

	// Try exact match first
	if transType, ok := typeMap[txType]; ok {
		return transType, nil
	}

	// Try to use base parser for common variations
	return p.ParseTransactionType(txType)
}
