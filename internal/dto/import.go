package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/models"
)

// ImportFormat represents the format of the import file
type ImportFormat string

const (
	ImportFormatGeneric           ImportFormat = "GENERIC"
	ImportFormatFidelity          ImportFormat = "FIDELITY"
	ImportFormatSchwab            ImportFormat = "SCHWAB"
	ImportFormatTDAmeritrade      ImportFormat = "TD_AMERITRADE"
	ImportFormatETrade            ImportFormat = "ETRADE"
	ImportFormatInteractiveBrokers ImportFormat = "INTERACTIVE_BROKERS"
	ImportFormatRobinhood         ImportFormat = "ROBINHOOD"
)

// ImportTransactionRequest represents a single transaction in the import
type ImportTransactionRequest struct {
	Type       models.TransactionType `json:"type"`
	Symbol     string                 `json:"symbol"`
	Date       time.Time              `json:"date"`
	Quantity   decimal.Decimal        `json:"quantity"`
	Price      *decimal.Decimal       `json:"price,omitempty"`
	Commission decimal.Decimal        `json:"commission"`
	Currency   string                 `json:"currency,omitempty"`
	Notes      string                 `json:"notes,omitempty"`
	RawData    string                 `json:"raw_data,omitempty"` // Original CSV line for debugging
}

// BulkImportRequest represents the request to import multiple transactions
type BulkImportRequest struct {
	Format       ImportFormat               `json:"format" binding:"required,oneof=GENERIC FIDELITY SCHWAB TD_AMERITRADE ETRADE INTERACTIVE_BROKERS ROBINHOOD"`
	Transactions []ImportTransactionRequest `json:"transactions" binding:"required,min=1"`
	DryRun       bool                       `json:"dry_run"`       // If true, validate but don't save
	SkipInvalid  bool                       `json:"skip_invalid"`  // If true, skip invalid transactions and continue
	Notes        string                     `json:"notes"`         // Optional notes about this import batch
}

// CSVImportRequest represents the request to import transactions from a CSV file
type CSVImportRequest struct {
	Format      ImportFormat `json:"format" binding:"required,oneof=GENERIC FIDELITY SCHWAB TD_AMERITRADE ETRADE INTERACTIVE_BROKERS ROBINHOOD"`
	CSVData     string       `json:"csv_data" binding:"required"`       // Base64 encoded CSV data or raw CSV text
	DryRun      bool         `json:"dry_run"`                           // If true, validate but don't save
	SkipInvalid bool         `json:"skip_invalid"`                      // If true, skip invalid transactions and continue
	Notes       string       `json:"notes"`                             // Optional notes about this import batch
}

// ImportError represents an error that occurred during import
type ImportError struct {
	Line    int    `json:"line"`              // Line number in the CSV (0 for general errors)
	Field   string `json:"field,omitempty"`   // Field that caused the error
	Message string `json:"message"`           // Error message
	RawData string `json:"raw_data,omitempty"` // Original data for debugging
}

// ImportValidationResult represents the validation result of a single transaction
type ImportValidationResult struct {
	Index  int           `json:"index"`
	Valid  bool          `json:"valid"`
	Errors []ImportError `json:"errors,omitempty"`
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	Success         bool                      `json:"success"`
	BatchID         uuid.UUID                 `json:"batch_id,omitempty"` // Import batch ID if successful
	TotalRows       int                       `json:"total_rows"`
	SuccessCount    int                       `json:"success_count"`
	ErrorCount      int                       `json:"error_count"`
	SkippedCount    int                       `json:"skipped_count"`
	Errors          []ImportError             `json:"errors,omitempty"`
	Transactions    []*TransactionResponse    `json:"transactions,omitempty"`    // Created transactions (if not dry run)
	ValidationOnly  bool                      `json:"validation_only"`           // True if this was a dry run
	ValidationResults []ImportValidationResult `json:"validation_results,omitempty"` // Detailed validation results
}

// ImportBatchInfo represents information about an import batch
type ImportBatchInfo struct {
	BatchID          uuid.UUID  `json:"batch_id"`
	PortfolioID      uuid.UUID  `json:"portfolio_id"`
	Format           ImportFormat `json:"format"`
	TransactionCount int        `json:"transaction_count"`
	Notes            string     `json:"notes,omitempty"`
	ImportedAt       time.Time  `json:"imported_at"`
}

// ImportBatchListResponse represents a list of import batches
type ImportBatchListResponse struct {
	Batches []*ImportBatchInfo `json:"batches"`
	Total   int                `json:"total"`
}

// ToImportBatchInfo converts import information to ImportBatchInfo DTO
func ToImportBatchInfo(batchID, portfolioID uuid.UUID, format ImportFormat, count int, notes string, importedAt time.Time) *ImportBatchInfo {
	return &ImportBatchInfo{
		BatchID:          batchID,
		PortfolioID:      portfolioID,
		Format:           format,
		TransactionCount: count,
		Notes:            notes,
		ImportedAt:       importedAt,
	}
}
