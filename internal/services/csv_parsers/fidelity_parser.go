package csv_parsers

import (
	"fmt"
	"io"

	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/models"
)

// FidelityParser handles Fidelity CSV format imports
// Fidelity CSV format typically includes:
// Run Date,Account,Action,Symbol,Security Description,Security Type,Quantity,Price,Commission,Fees,Accrued Interest,Amount,Settlement Date
type FidelityParser struct {
	BaseParser
}

// NewFidelityParser creates a new Fidelity CSV parser
func NewFidelityParser() CSVParser {
	return &FidelityParser{}
}

// GetFormat returns the format this parser handles
func (p *FidelityParser) GetFormat() dto.ImportFormat {
	return dto.ImportFormatFidelity
}

// ValidateHeaders validates that the CSV has the expected Fidelity headers
func (p *FidelityParser) ValidateHeaders(headers []string) error {
	// Look for key Fidelity columns
	if p.GetColumnIndex(headers, "Action", "Transaction Type") == -1 {
		return fmt.Errorf("missing Fidelity 'Action' column")
	}
	if p.GetColumnIndex(headers, "Symbol") == -1 {
		return fmt.Errorf("missing Fidelity 'Symbol' column")
	}
	return nil
}

// Parse parses Fidelity CSV data and returns import transaction requests
func (p *FidelityParser) Parse(data io.Reader) ([]dto.ImportTransactionRequest, []dto.ImportError, error) {
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

	// Get column indices for Fidelity-specific headers
	dateIdx := p.GetColumnIndex(headers, "Run Date", "Trade Date", "Settlement Date")
	actionIdx := p.GetColumnIndex(headers, "Action", "Transaction Type")
	symbolIdx := p.GetColumnIndex(headers, "Symbol")
	quantityIdx := p.GetColumnIndex(headers, "Quantity")
	priceIdx := p.GetColumnIndex(headers, "Price")
	commissionIdx := p.GetColumnIndex(headers, "Commission")
	feesIdx := p.GetColumnIndex(headers, "Fees")
	descIdx := p.GetColumnIndex(headers, "Security Description", "Description")

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

		// Parse Fidelity action to our transaction type
		actionStr := p.GetColumnValue(row, actionIdx)
		txType, err := p.parseFidelityAction(actionStr)
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

		// Handle negative quantities (Fidelity uses negative for sells)
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
				price = &priceVal
			}
		}

		// Parse commission and fees
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
		if feesIdx >= 0 {
			feesStr := p.GetColumnValue(row, feesIdx)
			if feesStr != "" {
				feesVal, err := p.ParseDecimal(feesStr)
				if err == nil {
					commission = commission.Add(feesVal.Abs())
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

// parseFidelityAction maps Fidelity action codes to our transaction types
func (p *FidelityParser) parseFidelityAction(action string) (models.TransactionType, error) {
	// Fidelity-specific action mapping
	actionMap := map[string]models.TransactionType{
		"YOU BOUGHT":              models.TransactionTypeBuy,
		"YOU SOLD":                models.TransactionTypeSell,
		"DIVIDEND RECEIVED":       models.TransactionTypeDividend,
		"CASH DIVIDEND":           models.TransactionTypeDividend,
		"REINVESTMENT":            models.TransactionTypeDividendReinvest,
		"DIVIDEND REINVESTMENT":   models.TransactionTypeDividendReinvest,
		"STOCK SPLIT":             models.TransactionTypeSplit,
		"EXCHANGE OR EXERCISE":    models.TransactionTypeMerger,
		"MERGER":                  models.TransactionTypeMerger,
		"SPINOFF":                 models.TransactionTypeSpinoff,
		"SYMBOL CHANGE":           models.TransactionTypeTickerChange,
	}

	// Try exact match first
	if txType, ok := actionMap[action]; ok {
		return txType, nil
	}

	// Try to use base parser for common variations
	return p.ParseTransactionType(action)
}
