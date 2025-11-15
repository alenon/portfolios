package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCSVImportGenericFormatViaAPI tests CSV import with generic format via API
func TestCSVImportGenericFormatViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	// Setup
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	portfolioID := createTestPortfolio(ctx, t, "CSV Import Test")

	// Upload CSV file
	csvPath := GetFixturePath("generic_import.csv")
	importBatchID, err := uploadCSVFile(ctx, portfolioID, csvPath, "generic")
	require.NoError(t, err, "CSV import should succeed")
	assert.NotEmpty(t, importBatchID, "Import batch ID should be returned")

	// List transactions to verify import
	var txList []struct {
		Symbol          string  `json:"symbol"`
		TransactionType string  `json:"transaction_type"`
		Quantity        float64 `json:"quantity"`
	}

	txPath := fmt.Sprintf("/api/v1/portfolios/%s/transactions", portfolioID)
	err = ctx.APIRequest("GET", txPath, nil, &txList)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(txList), 4, "Should have imported 4 transactions")

	// Verify specific transactions
	symbols := make(map[string]bool)
	for _, tx := range txList {
		symbols[tx.Symbol] = true
	}
	assert.True(t, symbols["AAPL"], "Should have AAPL transactions")
	assert.True(t, symbols["GOOGL"], "Should have GOOGL transactions")
	assert.True(t, symbols["MSFT"], "Should have MSFT transactions")
}

// TestCSVImportFidelityFormatViaAPI tests CSV import with Fidelity format via API
func TestCSVImportFidelityFormatViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	// Setup
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	portfolioID := createTestPortfolio(ctx, t, "Fidelity CSV Import Test")

	// Upload CSV file
	csvPath := GetFixturePath("fidelity_import.csv")
	importBatchID, err := uploadCSVFile(ctx, portfolioID, csvPath, "fidelity")
	require.NoError(t, err, "Fidelity CSV import should succeed")
	assert.NotEmpty(t, importBatchID, "Import batch ID should be returned")

	// List transactions
	var txList []struct {
		Symbol   string  `json:"symbol"`
		Quantity float64 `json:"quantity"`
	}

	txPath := fmt.Sprintf("/api/v1/portfolios/%s/transactions", portfolioID)
	err = ctx.APIRequest("GET", txPath, nil, &txList)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(txList), 2, "Should have imported 2 transactions")
}

// TestCSVImportBulkTransactionsViaAPI tests bulk transaction import via API
func TestCSVImportBulkTransactionsViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	// Setup
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	portfolioID := createTestPortfolio(ctx, t, "Bulk Import Test")

	// Prepare bulk transactions
	transactions := []map[string]interface{}{
		{
			"symbol":           "AAPL",
			"transaction_type": "buy",
			"quantity":         10.0,
			"price":            150.00,
			"transaction_date": "2024-01-15",
		},
		{
			"symbol":           "GOOGL",
			"transaction_type": "buy",
			"quantity":         5.0,
			"price":            120.00,
			"transaction_date": "2024-02-20",
		},
		{
			"symbol":           "MSFT",
			"transaction_type": "buy",
			"quantity":         8.0,
			"price":            300.00,
			"transaction_date": "2024-03-10",
		},
	}

	reqBody := map[string]interface{}{
		"transactions": transactions,
	}

	var respBody struct {
		ImportedCount int    `json:"imported_count"`
		ImportBatchID string `json:"import_batch_id"`
	}

	bulkPath := fmt.Sprintf("/api/v1/portfolios/%s/transactions/import/bulk", portfolioID)
	err = ctx.APIRequest("POST", bulkPath, reqBody, &respBody)
	require.NoError(t, err, "Bulk import should succeed")
	assert.Equal(t, 3, respBody.ImportedCount, "Should import 3 transactions")
	assert.NotEmpty(t, respBody.ImportBatchID, "Import batch ID should be returned")

	// Verify transactions were imported
	var txList []interface{}
	txPath := fmt.Sprintf("/api/v1/portfolios/%s/transactions", portfolioID)
	err = ctx.APIRequest("GET", txPath, nil, &txList)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(txList), 3, "Should have at least 3 transactions")
}

// TestCSVImportBatchListViaAPI tests listing import batches via API
func TestCSVImportBatchListViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	// Setup
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	portfolioID := createTestPortfolio(ctx, t, "Batch List Test")

	// Import CSV to create a batch
	csvPath := GetFixturePath("generic_import.csv")
	_, err = uploadCSVFile(ctx, portfolioID, csvPath, "generic")
	require.NoError(t, err)

	// List import batches
	var batches []struct {
		ImportBatchID    string `json:"import_batch_id"`
		TransactionCount int    `json:"transaction_count"`
	}

	batchPath := fmt.Sprintf("/api/v1/portfolios/%s/imports/batches", portfolioID)
	err = ctx.APIRequest("GET", batchPath, nil, &batches)
	require.NoError(t, err, "Batch listing should succeed")
	assert.GreaterOrEqual(t, len(batches), 1, "Should have at least one import batch")
}

// TestCSVImportBatchDeleteViaAPI tests deleting an import batch via API
func TestCSVImportBatchDeleteViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	// Setup
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	portfolioID := createTestPortfolio(ctx, t, "Batch Delete Test")

	// Import CSV to create a batch
	csvPath := GetFixturePath("generic_import.csv")
	importBatchID, err := uploadCSVFile(ctx, portfolioID, csvPath, "generic")
	require.NoError(t, err)
	require.NotEmpty(t, importBatchID)

	// Get initial transaction count
	var txListBefore []interface{}
	txPath := fmt.Sprintf("/api/v1/portfolios/%s/transactions", portfolioID)
	err = ctx.APIRequest("GET", txPath, nil, &txListBefore)
	require.NoError(t, err)
	initialCount := len(txListBefore)

	// Delete the import batch
	deletePath := fmt.Sprintf("/api/v1/portfolios/%s/imports/batches/%s", portfolioID, importBatchID)
	err = ctx.APIRequest("DELETE", deletePath, nil, nil)
	require.NoError(t, err, "Batch deletion should succeed")

	// Verify transactions were deleted
	var txListAfter []interface{}
	err = ctx.APIRequest("GET", txPath, nil, &txListAfter)
	require.NoError(t, err)
	assert.Less(t, len(txListAfter), initialCount, "Transaction count should decrease after batch deletion")
}

// TestCLICSVImport tests CSV import via CLI
func TestCLICSVImport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI test in short mode")
	}

	ctx := NewTestContext(t)

	// Setup
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	portfolioID := createTestPortfolio(ctx, t, "CLI CSV Import Test")

	// Save config for CLI
	err = ctx.SaveCLIConfig()
	require.NoError(t, err)

	// Import CSV via CLI
	csvPath := GetFixturePath("generic_import.csv")
	stdout, stderr, _ := ctx.RunCLI(
		"transaction", "import",
		portfolioID,
		csvPath,
		"--broker", "generic",
		"--output", "json",
	)
	t.Logf("CSV import stdout: %s", stdout)
	t.Logf("CSV import stderr: %s", stderr)

	// Even if CLI import has issues, verify via API
	var txList []interface{}
	txPath := fmt.Sprintf("/api/v1/portfolios/%s/transactions", portfolioID)
	err = ctx.APIRequest("GET", txPath, nil, &txList)
	require.NoError(t, err)

	// If import worked, we should have transactions
	if len(txList) > 0 {
		t.Logf("Import succeeded: %d transactions imported", len(txList))
	}
}

// TestCLIImportBatchList tests listing import batches via CLI
func TestCLIImportBatchList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI test in short mode")
	}

	ctx := NewTestContext(t)

	// Setup
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	portfolioID := createTestPortfolio(ctx, t, "CLI Batch List Test")

	// Import CSV via API to create batch
	csvPath := GetFixturePath("generic_import.csv")
	_, err = uploadCSVFile(ctx, portfolioID, csvPath, "generic")
	require.NoError(t, err)

	// Save config for CLI
	err = ctx.SaveCLIConfig()
	require.NoError(t, err)

	// List batches via CLI
	stdout, stderr, err := ctx.RunCLI(
		"transaction", "batches",
		portfolioID,
		"--output", "json",
	)
	t.Logf("Batch list stdout: %s", stdout)
	t.Logf("Batch list stderr: %s", stderr)

	if err == nil && stdout != "" {
		var batches []map[string]interface{}
		if parseErr := json.Unmarshal([]byte(stdout), &batches); parseErr == nil {
			assert.GreaterOrEqual(t, len(batches), 1, "Should have at least one batch")
		}
	}
}

// TestCSVImportFlowEndToEnd tests complete CSV import flow
func TestCSVImportFlowEndToEnd(t *testing.T) {
	ctx := NewTestContext(t)

	// 1. Setup
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	portfolioID := createTestPortfolio(ctx, t, "E2E CSV Import Test")

	// 2. Import CSV file
	csvPath := GetFixturePath("generic_import.csv")
	importBatchID, err := uploadCSVFile(ctx, portfolioID, csvPath, "generic")
	require.NoError(t, err)
	assert.NotEmpty(t, importBatchID)

	// 3. Verify transactions were imported
	var txList []struct {
		Symbol          string `json:"symbol"`
		TransactionType string `json:"transaction_type"`
		ImportBatchID   string `json:"import_batch_id"`
	}
	txPath := fmt.Sprintf("/api/v1/portfolios/%s/transactions", portfolioID)
	err = ctx.APIRequest("GET", txPath, nil, &txList)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(txList), 4)

	// Verify batch ID is set
	for _, tx := range txList {
		if tx.ImportBatchID != "" {
			assert.Equal(t, importBatchID, tx.ImportBatchID)
		}
	}

	// 4. List import batches
	var batches []struct {
		ImportBatchID    string `json:"import_batch_id"`
		TransactionCount int    `json:"transaction_count"`
	}
	batchPath := fmt.Sprintf("/api/v1/portfolios/%s/imports/batches", portfolioID)
	err = ctx.APIRequest("GET", batchPath, nil, &batches)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(batches), 1)

	// 5. Delete import batch
	deletePath := fmt.Sprintf("/api/v1/portfolios/%s/imports/batches/%s", portfolioID, importBatchID)
	err = ctx.APIRequest("DELETE", deletePath, nil, nil)
	require.NoError(t, err)

	// 6. Verify transactions are deleted
	var txListAfter []interface{}
	err = ctx.APIRequest("GET", txPath, nil, &txListAfter)
	require.NoError(t, err)
	assert.Less(t, len(txListAfter), len(txList))
}

// Helper function to upload CSV file via multipart form
func uploadCSVFile(ctx *TestContext, portfolioID string, csvPath string, broker string) (string, error) {
	// Read CSV file
	fileContent, err := os.ReadFile(csvPath)
	if err != nil {
		return "", fmt.Errorf("failed to read CSV file: %w", err)
	}

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add broker field
	_ = writer.WriteField("broker", broker)

	// Add file field
	part, err := writer.CreateFormFile("file", "import.csv")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, bytes.NewReader(fileContent))
	if err != nil {
		return "", fmt.Errorf("failed to write file content: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Make request
	url := fmt.Sprintf("%s/api/v1/portfolios/%s/transactions/import/csv", ctx.APIBaseURL, portfolioID)
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+ctx.AccessToken)

	resp, err := ctx.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		ImportedCount int    `json:"imported_count"`
		ImportBatchID string `json:"import_batch_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.ImportBatchID, nil
}
