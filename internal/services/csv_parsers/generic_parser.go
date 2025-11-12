package csv_parsers

import (
	"fmt"
	"io"

	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/dto"
)

// GenericParser handles standard CSV format imports
// Expected format:
// Date,Type,Symbol,Quantity,Price,Commission,Currency,Notes
type GenericParser struct {
	BaseParser
}

// NewGenericParser creates a new generic CSV parser
func NewGenericParser() CSVParser {
	return &GenericParser{}
}

// GetFormat returns the format this parser handles
func (p *GenericParser) GetFormat() dto.ImportFormat {
	return dto.ImportFormatGeneric
}

// ValidateHeaders validates that the CSV has the expected headers
func (p *GenericParser) ValidateHeaders(headers []string) error {
	// Required columns
	requiredColumns := []string{"date", "type", "symbol", "quantity"}

	for _, required := range requiredColumns {
		if p.GetColumnIndex(headers, required) == -1 {
			return fmt.Errorf("missing required column: %s", required)
		}
	}

	return nil
}

// Parse parses CSV data and returns import transaction requests
func (p *GenericParser) Parse(data io.Reader) ([]dto.ImportTransactionRequest, []dto.ImportError, error) {
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

	// Get column indices
	dateIdx := p.GetColumnIndex(headers, "date", "trade date", "transaction date")
	typeIdx := p.GetColumnIndex(headers, "type", "transaction type", "action")
	symbolIdx := p.GetColumnIndex(headers, "symbol", "ticker", "stock symbol")
	quantityIdx := p.GetColumnIndex(headers, "quantity", "shares", "amount")
	priceIdx := p.GetColumnIndex(headers, "price", "unit price", "share price")
	commissionIdx := p.GetColumnIndex(headers, "commission", "fee", "fees")
	currencyIdx := p.GetColumnIndex(headers, "currency")
	notesIdx := p.GetColumnIndex(headers, "notes", "description", "memo")

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

		// Parse transaction type
		typeStr := p.GetColumnValue(row, typeIdx)
		txType, err := p.ParseTransactionType(typeStr)
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

		// Parse price (optional for some transaction types)
		var price *decimal.Decimal
		if priceIdx >= 0 {
			priceStr := p.GetColumnValue(row, priceIdx)
			if priceStr != "" {
				priceVal, err := p.ParseDecimal(priceStr)
				if err != nil {
					errors = append(errors, p.CreateImportError(lineNum, "price", err.Error(), rawData))
					continue
				}
				price = &priceVal
			}
		}

		// Parse commission (optional, defaults to 0)
		commission := decimal.Zero
		if commissionIdx >= 0 {
			commissionStr := p.GetColumnValue(row, commissionIdx)
			if commissionStr != "" {
				commissionVal, err := p.ParseDecimal(commissionStr)
				if err != nil {
					errors = append(errors, p.CreateImportError(lineNum, "commission", err.Error(), rawData))
					continue
				}
				commission = commissionVal
			}
		}

		// Parse currency (optional, defaults to USD)
		currency := "USD"
		if currencyIdx >= 0 {
			currencyVal := p.GetColumnValue(row, currencyIdx)
			if currencyVal != "" {
				currency = currencyVal
			}
		}

		// Parse notes (optional)
		notes := ""
		if notesIdx >= 0 {
			notes = p.GetColumnValue(row, notesIdx)
		}

		// Create import transaction request
		tx := dto.ImportTransactionRequest{
			Type:       txType,
			Symbol:     symbol,
			Date:       date,
			Quantity:   quantity,
			Price:      price,
			Commission: commission,
			Currency:   currency,
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
