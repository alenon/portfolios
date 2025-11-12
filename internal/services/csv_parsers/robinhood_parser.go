package csv_parsers

import (
	"fmt"
	"io"

	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/models"
)

// RobinhoodParser handles Robinhood CSV format imports
// Robinhood CSV format typically includes:
// Activity Date,Process Date,Settle Date,Instrument,Description,Trans Code,Quantity,Price,Amount
type RobinhoodParser struct {
	BaseParser
}

// NewRobinhoodParser creates a new Robinhood CSV parser
func NewRobinhoodParser() CSVParser {
	return &RobinhoodParser{}
}

// GetFormat returns the format this parser handles
func (p *RobinhoodParser) GetFormat() dto.ImportFormat {
	return dto.ImportFormatRobinhood
}

// ValidateHeaders validates that the CSV has the expected Robinhood headers
func (p *RobinhoodParser) ValidateHeaders(headers []string) error {
	// Look for key Robinhood columns
	if p.GetColumnIndex(headers, "Activity Date", "Trans Date") == -1 {
		return fmt.Errorf("missing Robinhood 'Activity Date' column")
	}
	if p.GetColumnIndex(headers, "Trans Code", "Description") == -1 {
		return fmt.Errorf("missing Robinhood 'Trans Code' or 'Description' column")
	}
	return nil
}

// Parse parses Robinhood CSV data and returns import transaction requests
func (p *RobinhoodParser) Parse(data io.Reader) ([]dto.ImportTransactionRequest, []dto.ImportError, error) {
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

	// Get column indices for Robinhood-specific headers
	dateIdx := p.GetColumnIndex(headers, "Activity Date", "Trans Date")
	transCodeIdx := p.GetColumnIndex(headers, "Trans Code")
	descIdx := p.GetColumnIndex(headers, "Description")
	instrumentIdx := p.GetColumnIndex(headers, "Instrument", "Symbol")
	quantityIdx := p.GetColumnIndex(headers, "Quantity")
	priceIdx := p.GetColumnIndex(headers, "Price")

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

		// Parse transaction code/description
		transCode := p.GetColumnValue(row, transCodeIdx)
		description := p.GetColumnValue(row, descIdx)
		txType, err := p.parseRobinhoodTransCode(transCode, description)
		if err != nil {
			errors = append(errors, p.CreateImportError(lineNum, "trans_code", err.Error(), rawData))
			continue
		}

		// Parse symbol from instrument
		symbol := p.NormalizeSymbol(p.GetColumnValue(row, instrumentIdx))
		if symbol == "" {
			errors = append(errors, p.CreateImportError(lineNum, "instrument", "symbol is required", rawData))
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

		// Robinhood typically doesn't show commission (it's zero)
		commission := decimal.Zero

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

// parseRobinhoodTransCode maps Robinhood transaction codes to our transaction types
func (p *RobinhoodParser) parseRobinhoodTransCode(transCode, description string) (models.TransactionType, error) {
	// Robinhood-specific trans code mapping
	codeMap := map[string]models.TransactionType{
		"BUY":      models.TransactionTypeBuy,
		"SELL":     models.TransactionTypeSell,
		"DIV":      models.TransactionTypeDividend,
		"DIVIDEND": models.TransactionTypeDividend,
		"CDIV":     models.TransactionTypeDividend,
		"SPLIT":    models.TransactionTypeSplit,
		"SPINOFF":  models.TransactionTypeSpinoff,
		"MERGER":   models.TransactionTypeMerger,
	}

	// Try exact match on trans code first
	if txType, ok := codeMap[transCode]; ok {
		return txType, nil
	}

	// Try to parse from description
	return p.ParseTransactionType(description)
}
