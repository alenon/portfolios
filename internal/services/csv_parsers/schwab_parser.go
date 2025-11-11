package csv_parsers

import (
	"fmt"
	"io"

	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/models"
)

// SchwabParser handles Schwab CSV format imports
// Schwab CSV format typically includes:
// Date,Action,Symbol,Description,Quantity,Price,Fees & Comm,Amount
type SchwabParser struct {
	BaseParser
}

// NewSchwabParser creates a new Schwab CSV parser
func NewSchwabParser() CSVParser {
	return &SchwabParser{}
}

// GetFormat returns the format this parser handles
func (p *SchwabParser) GetFormat() dto.ImportFormat {
	return dto.ImportFormatSchwab
}

// ValidateHeaders validates that the CSV has the expected Schwab headers
func (p *SchwabParser) ValidateHeaders(headers []string) error {
	// Look for key Schwab columns
	if p.GetColumnIndex(headers, "Date", "Trade Date") == -1 {
		return fmt.Errorf("missing Schwab 'Date' column")
	}
	if p.GetColumnIndex(headers, "Action") == -1 {
		return fmt.Errorf("missing Schwab 'Action' column")
	}
	return nil
}

// Parse parses Schwab CSV data and returns import transaction requests
func (p *SchwabParser) Parse(data io.Reader) ([]dto.ImportTransactionRequest, []dto.ImportError, error) {
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

	// Get column indices for Schwab-specific headers
	dateIdx := p.GetColumnIndex(headers, "Date", "Trade Date")
	actionIdx := p.GetColumnIndex(headers, "Action")
	symbolIdx := p.GetColumnIndex(headers, "Symbol")
	quantityIdx := p.GetColumnIndex(headers, "Quantity")
	priceIdx := p.GetColumnIndex(headers, "Price")
	feesIdx := p.GetColumnIndex(headers, "Fees & Comm", "Fees", "Commission")
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

		// Parse Schwab action to our transaction type
		actionStr := p.GetColumnValue(row, actionIdx)
		txType, err := p.parseSchwabAction(actionStr)
		if err != nil {
			errors = append(errors, p.CreateImportError(lineNum, "action", err.Error(), rawData))
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
				price = &priceVal
			}
		}

		// Parse fees & commission
		commission := decimal.Zero
		if feesIdx >= 0 {
			feesStr := p.GetColumnValue(row, feesIdx)
			if feesStr != "" {
				feesVal, err := p.ParseDecimal(feesStr)
				if err == nil {
					commission = feesVal.Abs()
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

// parseSchwabAction maps Schwab action codes to our transaction types
func (p *SchwabParser) parseSchwabAction(action string) (models.TransactionType, error) {
	// Schwab-specific action mapping
	actionMap := map[string]models.TransactionType{
		"BUY":                     models.TransactionTypeBuy,
		"SELL":                    models.TransactionTypeSell,
		"DIV":                     models.TransactionTypeDividend,
		"CASH DIVIDEND":           models.TransactionTypeDividend,
		"REINVEST DIVIDEND":       models.TransactionTypeDividendReinvest,
		"REINVEST SHARES":         models.TransactionTypeDividendReinvest,
		"STOCK SPLIT":             models.TransactionTypeSplit,
		"MERGER":                  models.TransactionTypeMerger,
		"SPINOFF":                 models.TransactionTypeSpinoff,
		"SYMBOL CHANGE":           models.TransactionTypeTickerChange,
		"NAME CHANGE":             models.TransactionTypeTickerChange,
	}

	// Try exact match first
	if txType, ok := actionMap[action]; ok {
		return txType, nil
	}

	// Try to use base parser for common variations
	return p.ParseTransactionType(action)
}
