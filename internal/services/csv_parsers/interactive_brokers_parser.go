package csv_parsers

import (
	"fmt"
	"io"

	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/models"
)

// InteractiveBrokersParser handles Interactive Brokers CSV format imports
// Interactive Brokers CSV format (Activity Flex Query) typically includes:
// DataDiscriminator,Asset Category,Currency,Symbol,Date/Time,Quantity,T. Price,C. Price,Proceeds,Comm/Fee,Basis,Code,Description
type InteractiveBrokersParser struct {
	BaseParser
}

// NewInteractiveBrokersParser creates a new Interactive Brokers CSV parser
func NewInteractiveBrokersParser() CSVParser {
	return &InteractiveBrokersParser{}
}

// GetFormat returns the format this parser handles
func (p *InteractiveBrokersParser) GetFormat() dto.ImportFormat {
	return dto.ImportFormatInteractiveBrokers
}

// ValidateHeaders validates that the CSV has the expected Interactive Brokers headers
func (p *InteractiveBrokersParser) ValidateHeaders(headers []string) error {
	// Look for key Interactive Brokers columns
	if p.GetColumnIndex(headers, "Date/Time", "Date") == -1 {
		return fmt.Errorf("missing Interactive Brokers 'Date/Time' column")
	}
	if p.GetColumnIndex(headers, "Symbol") == -1 {
		return fmt.Errorf("missing Interactive Brokers 'Symbol' column")
	}
	return nil
}

// Parse parses Interactive Brokers CSV data and returns import transaction requests
func (p *InteractiveBrokersParser) Parse(data io.Reader) ([]dto.ImportTransactionRequest, []dto.ImportError, error) {
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

	// Get column indices for Interactive Brokers-specific headers
	dateIdx := p.GetColumnIndex(headers, "Date/Time", "Date")
	symbolIdx := p.GetColumnIndex(headers, "Symbol")
	quantityIdx := p.GetColumnIndex(headers, "Quantity")
	priceIdx := p.GetColumnIndex(headers, "T. Price", "Price")
	commissionIdx := p.GetColumnIndex(headers, "Comm/Fee", "Commission")
	codeIdx := p.GetColumnIndex(headers, "Code", "Transaction Type")
	descIdx := p.GetColumnIndex(headers, "Description")
	currencyIdx := p.GetColumnIndex(headers, "Currency")

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

		// Parse Interactive Brokers transaction code
		code := p.GetColumnValue(row, codeIdx)
		description := p.GetColumnValue(row, descIdx)
		txType, err := p.parseIBCode(code, description)
		if err != nil {
			errors = append(errors, p.CreateImportError(lineNum, "code", err.Error(), rawData))
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

		// Handle negative quantities (IB uses negative for sells)
		if quantity.IsNegative() {
			quantity = quantity.Abs()
			if txType == models.TransactionTypeBuy {
				txType = models.TransactionTypeSell
			}
		}

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

		// Parse currency
		currency := "USD"
		if currencyIdx >= 0 {
			currencyVal := p.GetColumnValue(row, currencyIdx)
			if currencyVal != "" {
				currency = currencyVal
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

// parseIBCode maps Interactive Brokers transaction codes to our transaction types
func (p *InteractiveBrokersParser) parseIBCode(code, description string) (models.TransactionType, error) {
	// Interactive Brokers-specific code mapping
	codeMap := map[string]models.TransactionType{
		"O":   models.TransactionTypeBuy,   // Open position (buy)
		"C":   models.TransactionTypeSell,  // Close position (sell)
		"DIV": models.TransactionTypeDividend,
		"PL":  models.TransactionTypeSplit, // Stock split
		"TC":  models.TransactionTypeTickerChange,
	}

	// Try exact match on code first
	if txType, ok := codeMap[code]; ok {
		return txType, nil
	}

	// Try to parse from description
	return p.ParseTransactionType(description)
}
