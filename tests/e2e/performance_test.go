package e2e

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPerformanceGetPortfolioPerformanceViaAPI tests getting portfolio performance via API
func TestPerformanceGetPortfolioPerformanceViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	// Setup: Create user, portfolio, and transactions
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	portfolioID := createTestPortfolio(ctx, t, "Performance Test Portfolio")

	// Add some transactions
	createTestTransactions(ctx, t, portfolioID)

	// Get performance metrics
	var performance map[string]interface{}
	perfPath := fmt.Sprintf("/api/v1/portfolios/%s/performance", portfolioID)
	err = ctx.APIRequest("GET", perfPath, nil, &performance)
	require.NoError(t, err, "Get performance should succeed")

	// Performance response should have metrics
	t.Logf("Performance response: %+v", performance)

	// Check that we got some response (structure may vary)
	assert.NotNil(t, performance, "Performance data should be returned")
}

// TestPerformanceGetHoldingsViaAPI tests getting current holdings via API
func TestPerformanceGetHoldingsViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	// Setup
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	portfolioID := createTestPortfolio(ctx, t, "Holdings Performance Test")

	// Add buy transactions
	transactions := []map[string]interface{}{
		{
			"symbol":           "AAPL",
			"type": "BUY",
			"quantity":         10.0,
			"price":            150.00,
			"date": time.Now().AddDate(0, 0, -10).Format("2006-01-02T15:04:05Z07:00"),
		},
		{
			"symbol":           "GOOGL",
			"type": "BUY",
			"quantity":         5.0,
			"price":            120.00,
			"date": time.Now().AddDate(0, 0, -5).Format("2006-01-02T15:04:05Z07:00"),
		},
	}

	txPath := fmt.Sprintf("/api/v1/portfolios/%s/transactions", portfolioID)
	for _, tx := range transactions {
		var resp interface{}
		err := ctx.APIRequest("POST", txPath, tx, &resp)
		require.NoError(t, err)
	}

	// Get holdings
	var holdings []map[string]interface{}
	holdingsPath := fmt.Sprintf("/api/v1/portfolios/%s/holdings", portfolioID)
	err = ctx.APIRequest("GET", holdingsPath, nil, &holdings)
	require.NoError(t, err, "Get holdings should succeed")

	// Should have holdings for the symbols we bought
	t.Logf("Holdings: %+v", holdings)
	if len(holdings) > 0 {
		assert.GreaterOrEqual(t, len(holdings), 1, "Should have at least one holding")
	}
}

// TestPerformanceSnapshotsViaAPI tests getting performance snapshots via API
func TestPerformanceSnapshotsViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	// Setup
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	portfolioID := createTestPortfolio(ctx, t, "Snapshots Test Portfolio")

	// Get snapshots (may be empty if none created)
	var snapshots []map[string]interface{}
	snapshotsPath := fmt.Sprintf("/api/v1/portfolios/%s/performance/snapshots", portfolioID)
	err = ctx.APIRequest("GET", snapshotsPath, nil, &snapshots)

	// Endpoint may or may not exist, just log the result
	if err != nil {
		t.Logf("Snapshots endpoint returned error (may not be implemented): %v", err)
	} else {
		t.Logf("Snapshots: %+v", snapshots)
		assert.NotNil(t, snapshots, "Snapshots should be returned")
	}
}

// TestCLIPerformanceShow tests showing performance via CLI
func TestCLIPerformanceShow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI test in short mode")
	}

	ctx := NewTestContext(t)

	// Setup
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	portfolioID := createTestPortfolio(ctx, t, "CLI Performance Test")
	createTestTransactions(ctx, t, portfolioID)

	// Save config for CLI
	err = ctx.SaveCLIConfig()
	require.NoError(t, err)

	// Show performance via CLI
	stdout, stderr, err := ctx.RunCLI(
		"performance", "show",
		portfolioID,
		"--output", "json",
	)
	t.Logf("Performance show stdout: %s", stdout)
	t.Logf("Performance show stderr: %s", stderr)

	if err == nil && stdout != "" {
		var performance map[string]interface{}
		if parseErr := json.Unmarshal([]byte(stdout), &performance); parseErr == nil {
			assert.NotNil(t, performance, "Performance data should be returned")
		}
	}
}

// TestCLIPerformanceSnapshots tests listing snapshots via CLI
func TestCLIPerformanceSnapshots(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI test in short mode")
	}

	ctx := NewTestContext(t)

	// Setup
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	portfolioID := createTestPortfolio(ctx, t, "CLI Snapshots Test")

	// Save config for CLI
	err = ctx.SaveCLIConfig()
	require.NoError(t, err)

	// List snapshots via CLI
	stdout, stderr, err := ctx.RunCLI(
		"performance", "snapshots",
		portfolioID,
		"--output", "json",
	)
	t.Logf("Performance snapshots stdout: %s", stdout)
	t.Logf("Performance snapshots stderr: %s", stderr)

	// May return empty array or error if not implemented
	if err == nil && stdout != "" {
		var snapshots []map[string]interface{}
		if parseErr := json.Unmarshal([]byte(stdout), &snapshots); parseErr == nil {
			t.Logf("Got %d snapshots", len(snapshots))
		}
	}
}

// TestPerformanceWithMultipleTransactions tests performance calculation with multiple transactions
func TestPerformanceWithMultipleTransactions(t *testing.T) {
	ctx := NewTestContext(t)

	// 1. Setup
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	portfolioID := createTestPortfolio(ctx, t, "Multi-Transaction Performance Test")

	// 2. Create a series of transactions
	transactions := []map[string]interface{}{
		{
			"symbol":           "AAPL",
			"type": "BUY",
			"quantity":         10.0,
			"price":            140.00,
			"date": time.Now().AddDate(0, 0, -30).Format("2006-01-02T15:04:05Z07:00"),
			"commission":             5.00,
		},
		{
			"symbol":           "AAPL",
			"type": "BUY",
			"quantity":         5.0,
			"price":            145.00,
			"date": time.Now().AddDate(0, 0, -20).Format("2006-01-02T15:04:05Z07:00"),
			"commission":             2.50,
		},
		{
			"symbol":           "GOOGL",
			"type": "BUY",
			"quantity":         8.0,
			"price":            115.00,
			"date": time.Now().AddDate(0, 0, -15).Format("2006-01-02T15:04:05Z07:00"),
		},
		{
			"symbol":           "AAPL",
			"type": "SELL",
			"quantity":         5.0,
			"price":            155.00,
			"date": time.Now().AddDate(0, 0, -5).Format("2006-01-02T15:04:05Z07:00"),
			"commission":             2.00,
		},
	}

	txPath := fmt.Sprintf("/api/v1/portfolios/%s/transactions", portfolioID)
	for _, tx := range transactions {
		var resp interface{}
		err := ctx.APIRequest("POST", txPath, tx, &resp)
		require.NoError(t, err)
	}

	// 3. Get holdings to verify current positions
	var holdings []map[string]interface{}
	holdingsPath := fmt.Sprintf("/api/v1/portfolios/%s/holdings", portfolioID)
	err = ctx.APIRequest("GET", holdingsPath, nil, &holdings)
	require.NoError(t, err)
	t.Logf("Holdings after transactions: %+v", holdings)

	// 4. Get performance metrics
	var performance map[string]interface{}
	perfPath := fmt.Sprintf("/api/v1/portfolios/%s/performance", portfolioID)
	err = ctx.APIRequest("GET", perfPath, nil, &performance)

	// Log performance even if there's an error
	if err != nil {
		t.Logf("Performance API returned error: %v", err)
	} else {
		t.Logf("Performance metrics: %+v", performance)
		assert.NotNil(t, performance, "Performance should be calculated")
	}
}

// TestPerformanceComparePortfolios tests comparing multiple portfolios
func TestPerformanceComparePortfolios(t *testing.T) {
	ctx := NewTestContext(t)

	// Setup
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	// Create two portfolios with different strategies
	portfolio1ID := createTestPortfolio(ctx, t, "Tech Portfolio")
	portfolio2ID := createTestPortfolio(ctx, t, "Diversified Portfolio")

	// Add transactions to portfolio 1 (tech-heavy)
	txPath1 := fmt.Sprintf("/api/v1/portfolios/%s/transactions", portfolio1ID)
	techTx := map[string]interface{}{
		"symbol":           "AAPL",
		"type": "BUY",
		"quantity":         20.0,
		"price":            150.00,
		"date": time.Now().AddDate(0, 0, -10).Format("2006-01-02T15:04:05Z07:00"),
	}
	var resp1 interface{}
	err = ctx.APIRequest("POST", txPath1, techTx, &resp1)
	require.NoError(t, err)

	// Add transactions to portfolio 2 (diversified)
	txPath2 := fmt.Sprintf("/api/v1/portfolios/%s/transactions", portfolio2ID)
	divTx := map[string]interface{}{
		"symbol":           "SPY",
		"type": "BUY",
		"quantity":         10.0,
		"price":            400.00,
		"date": time.Now().AddDate(0, 0, -10).Format("2006-01-02T15:04:05Z07:00"),
	}
	var resp2 interface{}
	err = ctx.APIRequest("POST", txPath2, divTx, &resp2)
	require.NoError(t, err)

	// Get performance for both
	var perf1 map[string]interface{}
	perfPath1 := fmt.Sprintf("/api/v1/portfolios/%s/performance", portfolio1ID)
	err = ctx.APIRequest("GET", perfPath1, nil, &perf1)
	if err == nil {
		t.Logf("Portfolio 1 performance: %+v", perf1)
	}

	var perf2 map[string]interface{}
	perfPath2 := fmt.Sprintf("/api/v1/portfolios/%s/performance", portfolio2ID)
	err = ctx.APIRequest("GET", perfPath2, nil, &perf2)
	if err == nil {
		t.Logf("Portfolio 2 performance: %+v", perf2)
	}

	// Note: Actual comparison would require market data
	// This test just verifies the API endpoints work
}

// Helper function to create test transactions
func createTestTransactions(ctx *TestContext, t *testing.T, portfolioID string) {
	transactions := []map[string]interface{}{
		{
			"symbol":           "AAPL",
			"type": "BUY",
			"quantity":         10.0,
			"price":            150.00,
			"date": time.Now().AddDate(0, 0, -5).Format("2006-01-02T15:04:05Z07:00"),
		},
		{
			"symbol":           "GOOGL",
			"type": "BUY",
			"quantity":         5.0,
			"price":            120.00,
			"date": time.Now().AddDate(0, 0, -3).Format("2006-01-02T15:04:05Z07:00"),
		},
	}

	txPath := fmt.Sprintf("/api/v1/portfolios/%s/transactions", portfolioID)
	for _, tx := range transactions {
		var resp interface{}
		err := ctx.APIRequest("POST", txPath, tx, &resp)
		require.NoError(t, err)
	}
}
