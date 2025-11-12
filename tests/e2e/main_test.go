package e2e

import (
	"net/http"
	"os"
	"testing"
	"time"
)

// TestMain sets up and tears down the e2e test environment
func TestMain(m *testing.M) {
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
		// Backend is not ready - skip tests
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup happens automatically via test cleanup functions

	os.Exit(code)
}
