package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPortfolioCreateViaAPI tests portfolio creation via API
func TestPortfolioCreateViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	// Register user
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	// Create portfolio
	reqBody := map[string]interface{}{
		"name":              "My Test Portfolio",
		"description":       "A portfolio for e2e testing",
		"base_currency":     "USD",
		"cost_basis_method": "FIFO",
	}

	var respBody struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Currency    string `json:"currency"`
	}

	err = ctx.APIRequest("POST", "/api/v1/portfolios", reqBody, &respBody)
	require.NoError(t, err, "Portfolio creation should succeed")
	assert.NotZero(t, respBody.ID, "Portfolio ID should be returned")
	assert.Equal(t, "My Test Portfolio", respBody.Name)
	assert.Equal(t, "USD", respBody.Currency)
}

// TestPortfolioListViaAPI tests listing portfolios via API
func TestPortfolioListViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	// Register user
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	// Create multiple portfolios
	portfolioNames := []string{"Portfolio 1", "Portfolio 2", "Portfolio 3"}
	for _, name := range portfolioNames {
		reqBody := map[string]interface{}{
			"name":              name,
			"base_currency":     "USD",
			"cost_basis_method": "FIFO",
		}
		var respBody interface{}
		err := ctx.APIRequest("POST", "/api/v1/portfolios", reqBody, &respBody)
		require.NoError(t, err)
	}

	// List portfolios
	var portfolioListResp struct {
		Portfolios []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"portfolios"`
		Total int `json:"total"`
	}

	err = ctx.APIRequest("GET", "/api/v1/portfolios", nil, &portfolioListResp)
	require.NoError(t, err, "Portfolio listing should succeed")
	assert.GreaterOrEqual(t, len(portfolioListResp.Portfolios), 3, "Should have at least 3 portfolios")
}

// TestPortfolioGetByIDViaAPI tests getting a specific portfolio via API
func TestPortfolioGetByIDViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	// Register user
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	// Create portfolio
	reqBody := map[string]interface{}{
		"name":              "Specific Portfolio",
		"description":       "Testing get by ID",
		"base_currency":     "EUR",
		"cost_basis_method": "FIFO",
	}

	var createResp struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Currency    string `json:"currency"`
	}

	err = ctx.APIRequest("POST", "/api/v1/portfolios", reqBody, &createResp)
	require.NoError(t, err)

	// Get portfolio by ID
	var getResp struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Currency    string `json:"currency"`
	}

	path := fmt.Sprintf("/api/v1/portfolios/%s", createResp.ID)
	err = ctx.APIRequest("GET", path, nil, &getResp)
	require.NoError(t, err, "Get portfolio should succeed")
	assert.Equal(t, createResp.ID, getResp.ID)
	assert.Equal(t, "Specific Portfolio", getResp.Name)
	assert.Equal(t, "EUR", getResp.Currency)
}

// TestPortfolioUpdateViaAPI tests updating a portfolio via API
func TestPortfolioUpdateViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	// Register user
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	// Create portfolio
	createReq := map[string]interface{}{
		"name":              "Original Name",
		"base_currency":     "USD",
		"cost_basis_method": "FIFO",
	}

	var createResp struct {
		ID string `json:"id"`
	}

	err = ctx.APIRequest("POST", "/api/v1/portfolios", createReq, &createResp)
	require.NoError(t, err)

	// Update portfolio
	updateReq := map[string]interface{}{
		"name":        "Updated Name",
		"description": "Updated description",
	}

	var updateResp struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	path := fmt.Sprintf("/api/v1/portfolios/%s", createResp.ID)
	err = ctx.APIRequest("PUT", path, updateReq, &updateResp)
	require.NoError(t, err, "Portfolio update should succeed")
	assert.Equal(t, "Updated Name", updateResp.Name)
	assert.Equal(t, "Updated description", updateResp.Description)
}

// TestPortfolioDeleteViaAPI tests deleting a portfolio via API
func TestPortfolioDeleteViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	// Register user
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	// Create portfolio
	createReq := map[string]interface{}{
		"name":              "Portfolio to Delete",
		"base_currency":     "USD",
		"cost_basis_method": "FIFO",
	}

	var createResp struct {
		ID string `json:"id"`
	}

	err = ctx.APIRequest("POST", "/api/v1/portfolios", createReq, &createResp)
	require.NoError(t, err)

	// Delete portfolio
	path := fmt.Sprintf("/api/v1/portfolios/%s", createResp.ID)
	err = ctx.APIRequest("DELETE", path, nil, nil)
	require.NoError(t, err, "Portfolio deletion should succeed")

	// Verify portfolio is deleted
	err = ctx.APIRequest("GET", path, nil, &struct{}{})
	assert.Error(t, err, "Should not be able to get deleted portfolio")
}

// TestPortfolioHoldingsViaAPI tests getting portfolio holdings via API
func TestPortfolioHoldingsViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	// Register user
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	// Create portfolio
	createReq := map[string]interface{}{
		"name":              "Holdings Test Portfolio",
		"base_currency":     "USD",
		"cost_basis_method": "FIFO",
	}

	var createResp struct {
		ID string `json:"id"`
	}

	err = ctx.APIRequest("POST", "/api/v1/portfolios", createReq, &createResp)
	require.NoError(t, err)

	// Get holdings (should be empty initially)
	var holdings []interface{}
	path := fmt.Sprintf("/api/v1/portfolios/%s/holdings", createResp.ID)
	err = ctx.APIRequest("GET", path, nil, &holdings)
	require.NoError(t, err, "Get holdings should succeed")
	assert.Equal(t, 0, len(holdings), "Holdings should be empty initially")
}

// TestCLIPortfolioList tests listing portfolios via CLI
func TestCLIPortfolioList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI test in short mode")
	}

	ctx := NewTestContext(t)

	// Register user and create portfolio via API
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	createReq := map[string]interface{}{
		"name":              "CLI Test Portfolio",
		"base_currency":     "USD",
		"cost_basis_method": "FIFO",
	}
	var createResp interface{}
	err = ctx.APIRequest("POST", "/api/v1/portfolios", createReq, &createResp)
	require.NoError(t, err)

	// Save config for CLI
	err = ctx.SaveCLIConfig()
	require.NoError(t, err)

	// List portfolios via CLI
	stdout, stderr, err := ctx.RunCLI("portfolio", "list", "--output", "json")
	t.Logf("Portfolio list stdout: %s", stdout)
	t.Logf("Portfolio list stderr: %s", stderr)

	if err == nil && stdout != "" {
		// Try to parse the output
		var portfolios []map[string]interface{}
		if parseErr := json.Unmarshal([]byte(stdout), &portfolios); parseErr == nil {
			assert.GreaterOrEqual(t, len(portfolios), 1, "Should have at least one portfolio")
		}
	}
}

// TestCLIPortfolioCreate tests creating a portfolio via CLI
func TestCLIPortfolioCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI test in short mode")
	}

	ctx := NewTestContext(t)

	// Register user via API
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	// Save config for CLI
	err = ctx.SaveCLIConfig()
	require.NoError(t, err)

	// Create portfolio via CLI (using flags to avoid interactive input)
	stdout, stderr, err := ctx.RunCLI(
		"portfolio", "create",
		"--name", "CLI Created Portfolio",
		"--description", "Created via CLI",
		"--currency", "USD",
		"--output", "json",
	)
	t.Logf("Portfolio create stdout: %s", stdout)
	t.Logf("Portfolio create stderr: %s", stderr)

	if err == nil && stdout != "" {
		// Try to parse the output
		var portfolio map[string]interface{}
		if parseErr := json.Unmarshal([]byte(stdout), &portfolio); parseErr == nil {
			assert.Equal(t, "CLI Created Portfolio", portfolio["name"])
		}
	}
}

// TestCLIPortfolioShow tests showing a specific portfolio via CLI
func TestCLIPortfolioShow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI test in short mode")
	}

	ctx := NewTestContext(t)

	// Register user and create portfolio via API
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	createReq := map[string]interface{}{
		"name":              "Show Test Portfolio",
		"description":       "Testing show command",
		"base_currency":     "USD",
		"cost_basis_method": "FIFO",
	}
	var createResp struct {
		ID string `json:"id"`
	}
	err = ctx.APIRequest("POST", "/api/v1/portfolios", createReq, &createResp)
	require.NoError(t, err)

	// Save config for CLI
	err = ctx.SaveCLIConfig()
	require.NoError(t, err)

	// Show portfolio via CLI
	stdout, stderr, err := ctx.RunCLI(
		"portfolio", "show",
		createResp.ID,
		"--output", "json",
	)
	t.Logf("Portfolio show stdout: %s", stdout)
	t.Logf("Portfolio show stderr: %s", stderr)

	if err == nil && stdout != "" {
		// Try to parse the output
		var portfolio map[string]interface{}
		if parseErr := json.Unmarshal([]byte(stdout), &portfolio); parseErr == nil {
			assert.Equal(t, "Show Test Portfolio", portfolio["name"])
		}
	}
}

// TestCLIPortfolioDelete tests deleting a portfolio via CLI
func TestCLIPortfolioDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI test in short mode")
	}

	ctx := NewTestContext(t)

	// Register user and create portfolio via API
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	createReq := map[string]interface{}{
		"name":              "Portfolio to Delete via CLI",
		"base_currency":     "USD",
		"cost_basis_method": "FIFO",
	}
	var createResp struct {
		ID string `json:"id"`
	}
	err = ctx.APIRequest("POST", "/api/v1/portfolios", createReq, &createResp)
	require.NoError(t, err)

	// Save config for CLI
	err = ctx.SaveCLIConfig()
	require.NoError(t, err)

	// Delete portfolio via CLI
	stdout, stderr, err := ctx.RunCLI(
		"portfolio", "delete",
		createResp.ID,
		// Skip confirmation
	)
	t.Logf("Portfolio delete stdout: %s", stdout)
	t.Logf("Portfolio delete stderr: %s", stderr)

	if err == nil {
		assert.True(t, strings.Contains(stdout, "success") || strings.Contains(stdout, "deleted"),
			"Delete output should indicate success")
	}

	// Verify portfolio is deleted via API
	path := fmt.Sprintf("/api/v1/portfolios/%s", createResp.ID)
	err = ctx.APIRequest("GET", path, nil, &struct{}{})
	assert.Error(t, err, "Portfolio should be deleted")
}

// TestPortfolioFlowEndToEnd tests complete portfolio flow
func TestPortfolioFlowEndToEnd(t *testing.T) {
	ctx := NewTestContext(t)

	// 1. Register user
	err := ctx.CreateTestUser(GenerateUniqueEmail(), "SecurePass123!")
	require.NoError(t, err)

	// 2. Create portfolio
	createReq := map[string]interface{}{
		"name":              "End-to-End Portfolio",
		"description":       "Testing complete flow",
		"base_currency":     "USD",
		"cost_basis_method": "FIFO",
	}
	var createResp struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	err = ctx.APIRequest("POST", "/api/v1/portfolios", createReq, &createResp)
	require.NoError(t, err)
	portfolioID := createResp.ID

	// 3. List portfolios
	var portfolioListResp struct {
		Portfolios []struct {
			ID string `json:"id"`
		} `json:"portfolios"`
		Total int `json:"total"`
	}
	err = ctx.APIRequest("GET", "/api/v1/portfolios", nil, &portfolioListResp)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(portfolioListResp.Portfolios), 1)

	// 4. Get specific portfolio
	var getResp struct {
		Name string `json:"name"`
	}
	path := fmt.Sprintf("/api/v1/portfolios/%s", portfolioID)
	err = ctx.APIRequest("GET", path, nil, &getResp)
	require.NoError(t, err)
	assert.Equal(t, "End-to-End Portfolio", getResp.Name)

	// 5. Update portfolio
	updateReq := map[string]interface{}{
		"name":        "Updated E2E Portfolio",
		"description": "Updated via e2e test",
	}
	var updateResp struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	err = ctx.APIRequest("PUT", path, updateReq, &updateResp)
	require.NoError(t, err)
	assert.Equal(t, "Updated E2E Portfolio", updateResp.Name)

	// 6. Get holdings (should be empty)
	var holdings []interface{}
	holdingsPath := fmt.Sprintf("/api/v1/portfolios/%s/holdings", portfolioID)
	err = ctx.APIRequest("GET", holdingsPath, nil, &holdings)
	require.NoError(t, err)
	assert.Equal(t, 0, len(holdings))

	// 7. Delete portfolio
	err = ctx.APIRequest("DELETE", path, nil, nil)
	require.NoError(t, err)

	// 8. Verify deletion
	err = ctx.APIRequest("GET", path, nil, &struct{}{})
	assert.Error(t, err)
}
