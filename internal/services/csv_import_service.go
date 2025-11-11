package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
	"github.com/lenon/portfolios/internal/services/csv_parsers"
)

// CSVImportService defines the interface for CSV import operations
type CSVImportService interface {
	// ImportFromCSV imports transactions from CSV data
	ImportFromCSV(portfolioID, userID string, req dto.CSVImportRequest) (*dto.ImportResult, error)

	// ImportBulk imports a list of pre-parsed transactions
	ImportBulk(portfolioID, userID string, req dto.BulkImportRequest) (*dto.ImportResult, error)

	// GetImportBatches retrieves all import batches for a portfolio
	GetImportBatches(portfolioID, userID string) (*dto.ImportBatchListResponse, error)

	// DeleteImportBatch deletes all transactions from a specific import batch
	DeleteImportBatch(portfolioID, userID string, batchID uuid.UUID) error
}

// csvImportService implements CSVImportService interface
type csvImportService struct {
	transactionRepo repository.TransactionRepository
	portfolioRepo   repository.PortfolioRepository
	holdingRepo     repository.HoldingRepository
	parsers         map[dto.ImportFormat]csv_parsers.CSVParser
}

// NewCSVImportService creates a new CSVImportService instance
func NewCSVImportService(
	transactionRepo repository.TransactionRepository,
	portfolioRepo repository.PortfolioRepository,
	holdingRepo repository.HoldingRepository,
) CSVImportService {
	// Initialize all parsers
	parsers := map[dto.ImportFormat]csv_parsers.CSVParser{
		dto.ImportFormatGeneric:           csv_parsers.NewGenericParser(),
		dto.ImportFormatFidelity:          csv_parsers.NewFidelityParser(),
		dto.ImportFormatSchwab:            csv_parsers.NewSchwabParser(),
		dto.ImportFormatTDAmeritrade:      csv_parsers.NewTDAmeritradeParser(),
		dto.ImportFormatETrade:            csv_parsers.NewETradeParser(),
		dto.ImportFormatInteractiveBrokers: csv_parsers.NewInteractiveBrokersParser(),
		dto.ImportFormatRobinhood:         csv_parsers.NewRobinhoodParser(),
	}

	return &csvImportService{
		transactionRepo: transactionRepo,
		portfolioRepo:   portfolioRepo,
		holdingRepo:     holdingRepo,
		parsers:         parsers,
	}
}

// ImportFromCSV imports transactions from CSV data
func (s *csvImportService) ImportFromCSV(portfolioID, userID string, req dto.CSVImportRequest) (*dto.ImportResult, error) {
	// Verify portfolio exists and user has access
	if err := s.verifyPortfolioAccess(portfolioID, userID); err != nil {
		return nil, err
	}

	// Get appropriate parser for the format
	parser, ok := s.parsers[req.Format]
	if !ok {
		return nil, fmt.Errorf("unsupported import format: %s", req.Format)
	}

	// Decode CSV data (support both raw text and base64)
	var csvData []byte
	if isBase64(req.CSVData) {
		decoded, err := base64.StdEncoding.DecodeString(req.CSVData)
		if err != nil {
			// If base64 decoding fails, assume it's raw text
			csvData = []byte(req.CSVData)
		} else {
			csvData = decoded
		}
	} else {
		csvData = []byte(req.CSVData)
	}

	// Parse CSV using the appropriate parser
	reader := bytes.NewReader(csvData)
	transactions, parseErrors, err := parser.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	// Create bulk import request from parsed transactions
	bulkReq := dto.BulkImportRequest{
		Format:       req.Format,
		Transactions: transactions,
		DryRun:       req.DryRun,
		SkipInvalid:  req.SkipInvalid,
		Notes:        req.Notes,
	}

	// Import transactions
	result, err := s.ImportBulk(portfolioID, userID, bulkReq)
	if err != nil {
		return nil, err
	}

	// Add parse errors to result
	if len(parseErrors) > 0 {
		result.Errors = append(result.Errors, parseErrors...)
		result.ErrorCount += len(parseErrors)
	}

	return result, nil
}

// ImportBulk imports a list of pre-parsed transactions
func (s *csvImportService) ImportBulk(portfolioID, userID string, req dto.BulkImportRequest) (*dto.ImportResult, error) {
	// Verify portfolio exists and user has access
	if err := s.verifyPortfolioAccess(portfolioID, userID); err != nil {
		return nil, err
	}

	// Parse portfolio ID
	portfolioUUID, err := uuid.Parse(portfolioID)
	if err != nil {
		return nil, models.ErrInvalidPortfolioID
	}

	// Generate batch ID for this import
	batchID := uuid.New()

	// Initialize result
	result := &dto.ImportResult{
		Success:          true,
		BatchID:          batchID,
		TotalRows:        len(req.Transactions),
		SuccessCount:     0,
		ErrorCount:       0,
		SkippedCount:     0,
		Errors:           []dto.ImportError{},
		Transactions:     []*dto.TransactionResponse{},
		ValidationOnly:   req.DryRun,
		ValidationResults: []dto.ImportValidationResult{},
	}

	// Process each transaction
	for i, txReq := range req.Transactions {
		validationResult := dto.ImportValidationResult{
			Index:  i,
			Valid:  true,
			Errors: []dto.ImportError{},
		}

		// Validate transaction
		if err := s.validateImportTransaction(&txReq); err != nil {
			validationResult.Valid = false
			validationResult.Errors = append(validationResult.Errors, dto.ImportError{
				Line:    i + 1,
				Message: err.Error(),
				RawData: txReq.RawData,
			})

			result.ValidationResults = append(result.ValidationResults, validationResult)
			result.Errors = append(result.Errors, validationResult.Errors...)
			result.ErrorCount++

			if !req.SkipInvalid {
				result.Success = false
				return result, nil
			}
			result.SkippedCount++
			continue
		}

		// If dry run, just validate
		if req.DryRun {
			result.ValidationResults = append(result.ValidationResults, validationResult)
			result.SuccessCount++
			continue
		}

		// Create transaction model
		transaction := &models.Transaction{
			PortfolioID:   portfolioUUID,
			Type:          txReq.Type,
			Symbol:        txReq.Symbol,
			Date:          txReq.Date,
			Quantity:      txReq.Quantity,
			Price:         txReq.Price,
			Commission:    txReq.Commission,
			Currency:      txReq.Currency,
			Notes:         txReq.Notes,
			ImportBatchID: &batchID,
		}

		// Save transaction
		err = s.transactionRepo.Create(transaction)
		if err != nil {
			validationResult.Valid = false
			validationResult.Errors = append(validationResult.Errors, dto.ImportError{
				Line:    i + 1,
				Message: fmt.Sprintf("failed to create transaction: %v", err),
				RawData: txReq.RawData,
			})

			result.ValidationResults = append(result.ValidationResults, validationResult)
			result.Errors = append(result.Errors, validationResult.Errors...)
			result.ErrorCount++

			if !req.SkipInvalid {
				result.Success = false
				return result, nil
			}
			result.SkippedCount++
			continue
		}

		// Update holdings based on transaction
		if err := s.updateHoldingsForTransaction(transaction); err != nil {
			// Log error but don't fail the import
			// Holdings can be recalculated later if needed
			log.Printf("Warning: Failed to update holdings for transaction %s: %v", transaction.ID, err)
		}

		result.ValidationResults = append(result.ValidationResults, validationResult)
		result.Transactions = append(result.Transactions, dto.ToTransactionResponse(transaction))
		result.SuccessCount++
	}

	// If no transactions were successfully imported, mark as failed
	if result.SuccessCount == 0 && result.TotalRows > 0 {
		result.Success = false
	}

	return result, nil
}

// GetImportBatches retrieves all import batches for a portfolio
func (s *csvImportService) GetImportBatches(portfolioID, userID string) (*dto.ImportBatchListResponse, error) {
	// Verify portfolio exists and user has access
	if err := s.verifyPortfolioAccess(portfolioID, userID); err != nil {
		return nil, err
	}

	// Get all transactions for this portfolio
	transactions, err := s.transactionRepo.FindByPortfolioID(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Group transactions by import batch ID
	batchMap := make(map[uuid.UUID][]*models.Transaction)
	for _, tx := range transactions {
		if tx.ImportBatchID != nil {
			batchMap[*tx.ImportBatchID] = append(batchMap[*tx.ImportBatchID], tx)
		}
	}

	// Create batch info list
	batches := make([]*dto.ImportBatchInfo, 0, len(batchMap))
	portfolioUUID, _ := uuid.Parse(portfolioID)

	for batchID, batchTxs := range batchMap {
		if len(batchTxs) == 0 {
			continue
		}

		// Use the earliest transaction date as the import date
		importedAt := batchTxs[0].CreatedAt
		for _, tx := range batchTxs {
			if tx.CreatedAt.Before(importedAt) {
				importedAt = tx.CreatedAt
			}
		}

		// Get notes from first transaction (if any)
		notes := ""
		if len(batchTxs) > 0 && batchTxs[0].Notes != "" {
			notes = "Batch import"
		}

		batchInfo := dto.ToImportBatchInfo(
			batchID,
			portfolioUUID,
			dto.ImportFormatGeneric, // We don't store format, so default to generic
			len(batchTxs),
			notes,
			importedAt,
		)
		batches = append(batches, batchInfo)
	}

	return &dto.ImportBatchListResponse{
		Batches: batches,
		Total:   len(batches),
	}, nil
}

// DeleteImportBatch deletes all transactions from a specific import batch
func (s *csvImportService) DeleteImportBatch(portfolioID, userID string, batchID uuid.UUID) error {
	// Verify portfolio exists and user has access
	if err := s.verifyPortfolioAccess(portfolioID, userID); err != nil {
		return err
	}

	// Get all transactions in this batch
	transactions, err := s.transactionRepo.FindByPortfolioID(portfolioID)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	// Delete transactions that belong to this batch
	affectedSymbols := make(map[string]bool)
	for _, tx := range transactions {
		if tx.ImportBatchID != nil && *tx.ImportBatchID == batchID {
			if err := s.transactionRepo.Delete(tx.ID.String()); err != nil {
				return fmt.Errorf("failed to delete transaction: %w", err)
			}
			affectedSymbols[tx.Symbol] = true
		}
	}

	// Recalculate holdings for affected symbols
	for symbol := range affectedSymbols {
		if err := s.recalculateHoldingsForSymbol(portfolioID, symbol); err != nil {
			// Log error but don't fail the deletion
			log.Printf("Warning: Failed to recalculate holdings for symbol %s in portfolio %s: %v", symbol, portfolioID, err)
		}
	}

	return nil
}

// verifyPortfolioAccess verifies that the portfolio exists and the user has access
func (s *csvImportService) verifyPortfolioAccess(portfolioID, userID string) error {
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return err
	}

	if portfolio.UserID.String() != userID {
		return models.ErrPortfolioNotFound // Don't leak existence of other users' portfolios
	}

	return nil
}

// validateImportTransaction validates an import transaction request
func (s *csvImportService) validateImportTransaction(tx *dto.ImportTransactionRequest) error {
	if tx.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if tx.Date.IsZero() {
		return fmt.Errorf("date is required")
	}

	if tx.Date.After(time.Now().Add(24 * time.Hour)) {
		return fmt.Errorf("date cannot be in the future")
	}

	if tx.Quantity.IsZero() || tx.Quantity.IsNegative() {
		return fmt.Errorf("quantity must be greater than zero")
	}

	// Price is required for BUY and SELL transactions
	if (tx.Type == models.TransactionTypeBuy || tx.Type == models.TransactionTypeSell) {
		if tx.Price == nil || tx.Price.IsZero() || tx.Price.IsNegative() {
			return fmt.Errorf("price is required and must be greater than zero for buy/sell transactions")
		}
	}

	if tx.Commission.IsNegative() {
		return fmt.Errorf("commission cannot be negative")
	}

	return nil
}

// updateHoldingsForTransaction updates holdings based on a transaction
func (s *csvImportService) updateHoldingsForTransaction(tx *models.Transaction) error {
	// Get current holding for this symbol
	holding, err := s.holdingRepo.FindByPortfolioIDAndSymbol(tx.PortfolioID.String(), tx.Symbol)
	if err != nil && err != models.ErrHoldingNotFound {
		return err
	}

	switch tx.Type {
	case models.TransactionTypeBuy, models.TransactionTypeDividendReinvest:
		if holding == nil {
			// Create new holding
			newHolding := &models.Holding{
				PortfolioID:  tx.PortfolioID,
				Symbol:       tx.Symbol,
				Quantity:     tx.Quantity,
				CostBasis:    tx.GetTotalCost(),
				AvgCostPrice: *tx.Price,
			}
			err = s.holdingRepo.Create(newHolding)
			return err
		} else {
			// Update existing holding using AddShares method
			holding.AddShares(tx.Quantity, tx.GetTotalCost())
			return s.holdingRepo.Update(holding)
		}

	case models.TransactionTypeSell:
		if holding == nil {
			return models.ErrInsufficientShares
		}

		holding.Quantity = holding.Quantity.Sub(tx.Quantity)
		if holding.Quantity.IsNegative() {
			return models.ErrInsufficientShares
		}

		if holding.Quantity.IsZero() {
			return s.holdingRepo.Delete(holding.ID.String())
		}

		return s.holdingRepo.Update(holding)
	}

	return nil
}

// recalculateHoldingsForSymbol recalculates holdings for a specific symbol
func (s *csvImportService) recalculateHoldingsForSymbol(portfolioID, symbol string) error {
	// Get all transactions for this symbol, ordered by date
	transactions, err := s.transactionRepo.FindByPortfolioIDAndSymbol(portfolioID, symbol)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	// Calculate holdings based on transactions
	quantity := decimal.Zero
	costBasis := decimal.Zero

	for _, tx := range transactions {
		switch tx.Type {
		case models.TransactionTypeBuy, models.TransactionTypeDividendReinvest:
			quantity = quantity.Add(tx.Quantity)
			costBasis = costBasis.Add(tx.GetTotalCost())
		case models.TransactionTypeSell:
			if quantity.IsZero() {
				return models.ErrInsufficientShares
			}
			avgCostPrice := costBasis.Div(quantity)
			costBasisForSale := avgCostPrice.Mul(tx.Quantity)
			quantity = quantity.Sub(tx.Quantity)
			costBasis = costBasis.Sub(costBasisForSale)
		}
	}

	// Update or delete holding
	holding, err := s.holdingRepo.FindByPortfolioIDAndSymbol(portfolioID, symbol)
	if err != nil && err != models.ErrHoldingNotFound {
		return err
	}

	if quantity.IsZero() {
		if holding != nil {
			return s.holdingRepo.Delete(holding.ID.String())
		}
		return nil
	}

	avgCost := costBasis.Div(quantity)
	if holding == nil {
		// Create new holding
		portfolioUUID, _ := uuid.Parse(portfolioID)
		newHolding := &models.Holding{
			PortfolioID:  portfolioUUID,
			Symbol:       symbol,
			Quantity:     quantity,
			CostBasis:    costBasis,
			AvgCostPrice: avgCost,
		}
		err = s.holdingRepo.Create(newHolding)
		return err
	}

	holding.Quantity = quantity
	holding.CostBasis = costBasis
	holding.AvgCostPrice = avgCost
	return s.holdingRepo.Update(holding)
}

// isBase64 checks if a string is base64 encoded
func isBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
