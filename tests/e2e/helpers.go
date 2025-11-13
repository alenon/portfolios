package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	// API endpoint - should match docker-compose.e2e.yml backend service
	APIBaseURL = "http://backend-e2e:8080"
	// CLI binary path
	CLIPath = "/usr/local/bin/portfolios"
	// Config directory for CLI
	ConfigDir = "/tmp/portfolios-e2e"
)

// TestContext holds the state for e2e tests
type TestContext struct {
	T            *testing.T
	ConfigDir    string
	APIBaseURL   string
	AccessToken  string
	RefreshToken string
	UserEmail    string
	UserPassword string
	HTTPClient   *http.Client
	CleanupFuncs []func()
}

// NewTestContext creates a new test context
func NewTestContext(t *testing.T) *TestContext {
	ctx := &TestContext{
		T:          t,
		ConfigDir:  ConfigDir,
		APIBaseURL: APIBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		CleanupFuncs: make([]func(), 0),
	}

	// Ensure config directory exists
	err := os.MkdirAll(ctx.ConfigDir, 0750)
	require.NoError(t, err, "Failed to create config directory")

	// Add cleanup function
	t.Cleanup(func() {
		ctx.Cleanup()
	})

	return ctx
}

// Cleanup runs all cleanup functions
func (ctx *TestContext) Cleanup() {
	for i := len(ctx.CleanupFuncs) - 1; i >= 0; i-- {
		ctx.CleanupFuncs[i]()
	}
}

// AddCleanup adds a cleanup function
func (ctx *TestContext) AddCleanup(f func()) {
	ctx.CleanupFuncs = append(ctx.CleanupFuncs, f)
}

// RunCLI executes a CLI command and returns stdout, stderr, and error
func (ctx *TestContext) RunCLI(args ...string) (stdout, stderr string, err error) {
	cmd := exec.Command(CLIPath, args...)

	// Set environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PORTFOLIOS_API_URL=%s", ctx.APIBaseURL),
		fmt.Sprintf("PORTFOLIOS_CONFIG_DIR=%s", ctx.ConfigDir),
		"PORTFOLIOS_OUTPUT_FORMAT=json", // Default to JSON for easier parsing
	)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	return stdout, stderr, err
}

// RunCLIWithInput executes a CLI command with stdin input
func (ctx *TestContext) RunCLIWithInput(input string, args ...string) (stdout, stderr string, err error) {
	cmd := exec.Command(CLIPath, args...)

	// Set environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PORTFOLIOS_API_URL=%s", ctx.APIBaseURL),
		fmt.Sprintf("PORTFOLIOS_CONFIG_DIR=%s", ctx.ConfigDir),
		"PORTFOLIOS_OUTPUT_FORMAT=json",
	)

	cmd.Stdin = strings.NewReader(input)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	return stdout, stderr, err
}

// APIRequest makes a direct HTTP request to the backend API
func (ctx *TestContext) APIRequest(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, ctx.APIBaseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if ctx.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+ctx.AccessToken)
	}

	resp, err := ctx.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// WaitForBackend waits for the backend to be healthy
func (ctx *TestContext) WaitForBackend(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := ctx.HTTPClient.Get(ctx.APIBaseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			_ = resp.Body.Close()
			return nil
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("backend did not become healthy within %v", timeout)
}

// CreateTestUser creates a test user via API
func (ctx *TestContext) CreateTestUser(email, password string) error {
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}

	var respBody struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID    uint   `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
	}

	if err := ctx.APIRequest("POST", "/api/v1/auth/register", reqBody, &respBody); err != nil {
		return err
	}

	ctx.AccessToken = respBody.AccessToken
	ctx.RefreshToken = respBody.RefreshToken
	ctx.UserEmail = email
	ctx.UserPassword = password

	return nil
}

// Login logs in via API
func (ctx *TestContext) Login(email, password string) error {
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}

	var respBody struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := ctx.APIRequest("POST", "/api/v1/auth/login", reqBody, &respBody); err != nil {
		return err
	}

	ctx.AccessToken = respBody.AccessToken
	ctx.RefreshToken = respBody.RefreshToken
	ctx.UserEmail = email
	ctx.UserPassword = password

	return nil
}

// SaveCLIConfig saves the config file for CLI commands
func (ctx *TestContext) SaveCLIConfig() error {
	configPath := filepath.Join(ctx.ConfigDir, "config.yaml")

	// Create a simple YAML format
	configContent := fmt.Sprintf(`api_base_url: %s
access_token: %s
refresh_token: %s
output_format: json
`, ctx.APIBaseURL, ctx.AccessToken, ctx.RefreshToken)

	return os.WriteFile(configPath, []byte(configContent), 0600)
}

// ParseJSONOutput parses JSON output from CLI command
func ParseJSONOutput(output string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return result, nil
}

// ParseJSONArrayOutput parses JSON array output from CLI command
func ParseJSONArrayOutput(output string) ([]interface{}, error) {
	var result []interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON array: %w", err)
	}
	return result, nil
}

// GenerateUniqueEmail generates a unique email for testing
func GenerateUniqueEmail() string {
	return fmt.Sprintf("test-%d@e2e.test", time.Now().UnixNano())
}

// GetFixturePath returns the path to a fixture file
func GetFixturePath(filename string) string {
	return filepath.Join("/fixtures", filename)
}
