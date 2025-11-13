package e2e

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestMain sets up and tears down the e2e test environment
func TestMain(m *testing.M) {
	// Check if we're in the E2E environment (Docker)
	// If not, skip all tests gracefully
	if !isE2EEnvironment() {
		fmt.Println("Skipping E2E tests - not in Docker environment")
		fmt.Println("To run E2E tests: make e2e-test")
		os.Exit(0)
	}

	// Setup: Wait for backend to be ready
	ctx := &TestContext{
		T:          &testing.T{},
		ConfigDir:  ConfigDir,
		APIBaseURL: APIBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Wait for backend to be healthy before running tests
	if err := ctx.WaitForBackend(60 * time.Second); err != nil {
		fmt.Printf("Error: Backend is not ready - %v\n", err)
		fmt.Println("Make sure to run: make e2e-up")
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup happens automatically via test cleanup functions

	os.Exit(code)
}

// isE2EEnvironment checks if we're running in the E2E Docker environment
func isE2EEnvironment() bool {
	// Check if we can reach the backend health endpoint
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(APIBaseURL + "/health")
	if err != nil {
		return false
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return resp.StatusCode == http.StatusOK
}
