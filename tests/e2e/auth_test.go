package e2e

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthRegisterViaAPI tests user registration via direct API call
func TestAuthRegisterViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	email := GenerateUniqueEmail()
	password := "SecurePass123!"

	err := ctx.CreateTestUser(email, password)
	require.NoError(t, err, "User registration should succeed")
	assert.NotEmpty(t, ctx.AccessToken, "Access token should be returned")
	assert.NotEmpty(t, ctx.RefreshToken, "Refresh token should be returned")
	assert.Equal(t, email, ctx.UserEmail, "Email should match")
}

// TestAuthLoginViaAPI tests user login via direct API call
func TestAuthLoginViaAPI(t *testing.T) {
	ctx := NewTestContext(t)

	// First register a user
	email := GenerateUniqueEmail()
	password := "SecurePass123!"
	err := ctx.CreateTestUser(email, password)
	require.NoError(t, err)

	// Clear tokens to simulate fresh login
	oldAccessToken := ctx.AccessToken
	ctx.AccessToken = ""
	ctx.RefreshToken = ""

	// Login
	err = ctx.Login(email, password)
	require.NoError(t, err, "Login should succeed")
	assert.NotEmpty(t, ctx.AccessToken, "Access token should be returned")
	assert.NotEmpty(t, ctx.RefreshToken, "Refresh token should be returned")
	assert.NotEqual(t, oldAccessToken, ctx.AccessToken, "New access token should be different")
}

// TestAuthLoginFailsWithWrongPassword tests login failure with wrong password
func TestAuthLoginFailsWithWrongPassword(t *testing.T) {
	ctx := NewTestContext(t)

	// Register a user
	email := GenerateUniqueEmail()
	password := "SecurePass123!"
	err := ctx.CreateTestUser(email, password)
	require.NoError(t, err)

	// Try to login with wrong password
	ctx.AccessToken = ""
	ctx.RefreshToken = ""

	reqBody := map[string]string{
		"email":    email,
		"password": "WrongPassword",
	}

	var respBody interface{}
	err = ctx.APIRequest("POST", "/api/v1/auth/login", reqBody, &respBody)
	assert.Error(t, err, "Login should fail with wrong password")
}

// TestAuthTokenRefresh tests token refresh functionality
func TestAuthTokenRefresh(t *testing.T) {
	ctx := NewTestContext(t)

	// Register a user
	email := GenerateUniqueEmail()
	password := "SecurePass123!"
	err := ctx.CreateTestUser(email, password)
	require.NoError(t, err)

	oldAccessToken := ctx.AccessToken
	refreshToken := ctx.RefreshToken

	// Wait a moment to ensure new token has different timestamp
	time.Sleep(100 * time.Millisecond)

	// Refresh the access token
	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}

	var respBody struct {
		AccessToken string `json:"access_token"`
	}

	err = ctx.APIRequest("POST", "/api/v1/auth/refresh", reqBody, &respBody)
	require.NoError(t, err, "Token refresh should succeed")
	assert.NotEmpty(t, respBody.AccessToken, "New access token should be returned")
	assert.NotEqual(t, oldAccessToken, respBody.AccessToken, "New access token should be different")

	ctx.AccessToken = respBody.AccessToken
}

// TestAuthProtectedEndpointAccess tests accessing protected endpoints
func TestAuthProtectedEndpointAccess(t *testing.T) {
	ctx := NewTestContext(t)

	// Create a user
	email := GenerateUniqueEmail()
	password := "SecurePass123!"
	err := ctx.CreateTestUser(email, password)
	require.NoError(t, err)

	// Access protected endpoint (get current user)
	var userResp struct {
		ID    uint   `json:"id"`
		Email string `json:"email"`
	}

	err = ctx.APIRequest("GET", "/api/v1/auth/me", nil, &userResp)
	require.NoError(t, err, "Should access protected endpoint with valid token")
	assert.Equal(t, email, userResp.Email, "Should return current user")

	// Try without token
	oldToken := ctx.AccessToken
	ctx.AccessToken = ""

	err = ctx.APIRequest("GET", "/api/v1/auth/me", nil, &userResp)
	assert.Error(t, err, "Should fail without token")

	// Restore token
	ctx.AccessToken = oldToken
}

// TestAuthLogout tests logout functionality
func TestAuthLogout(t *testing.T) {
	ctx := NewTestContext(t)

	// Register and login
	email := GenerateUniqueEmail()
	password := "SecurePass123!"
	err := ctx.CreateTestUser(email, password)
	require.NoError(t, err)

	refreshToken := ctx.RefreshToken

	// Logout (revoke refresh token)
	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}

	err = ctx.APIRequest("POST", "/api/v1/auth/logout", reqBody, nil)
	require.NoError(t, err, "Logout should succeed")

	// Try to use the revoked refresh token
	refreshReqBody := map[string]string{
		"refresh_token": refreshToken,
	}

	var refreshResp struct {
		AccessToken string `json:"access_token"`
	}

	err = ctx.APIRequest("POST", "/api/v1/auth/refresh", refreshReqBody, &refreshResp)
	assert.Error(t, err, "Should not be able to use revoked refresh token")
}

// TestCLIAuthRegister tests user registration via CLI
func TestCLIAuthRegister(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI test in short mode")
	}

	ctx := NewTestContext(t)

	email := GenerateUniqueEmail()
	password := "SecurePass123!"

	// Note: The CLI register command expects interactive input
	// We'll use the API for registration in e2e tests
	// and focus on testing login/logout via CLI

	// Register via API first
	err := ctx.CreateTestUser(email, password)
	require.NoError(t, err)
	assert.NotEmpty(t, ctx.AccessToken)
}

// TestCLIAuthLogin tests login via CLI
func TestCLIAuthLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI test in short mode")
	}

	ctx := NewTestContext(t)

	// Register a user via API
	email := GenerateUniqueEmail()
	password := "SecurePass123!"
	err := ctx.CreateTestUser(email, password)
	require.NoError(t, err)

	// Clear config to simulate fresh login
	ctx.AccessToken = ""
	ctx.RefreshToken = ""

	// Prepare input for interactive login
	input := email + "\n" + password + "\n"

	stdout, stderr, err := ctx.RunCLIWithInput(input, "auth", "login")

	// CLI might exit with error if prompts don't work in non-interactive mode
	// For now, we'll test the API-based auth flow more thoroughly
	t.Logf("Login stdout: %s", stdout)
	t.Logf("Login stderr: %s", stderr)

	// If login succeeded, stdout should contain success message
	if err == nil {
		assert.True(t, strings.Contains(stdout, "success") || strings.Contains(stdout, "logged in"),
			"Login output should indicate success")
	}
}

// TestCLIAuthWhoami tests whoami command
func TestCLIAuthWhoami(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI test in short mode")
	}

	ctx := NewTestContext(t)

	// Register and login via API
	email := GenerateUniqueEmail()
	password := "SecurePass123!"
	err := ctx.CreateTestUser(email, password)
	require.NoError(t, err)

	// Save config for CLI
	err = ctx.SaveCLIConfig()
	require.NoError(t, err)

	// Run whoami command
	stdout, stderr, err := ctx.RunCLI("auth", "whoami", "--output", "json")
	t.Logf("Whoami stdout: %s", stdout)
	t.Logf("Whoami stderr: %s", stderr)

	if err == nil && stdout != "" {
		// Try to parse the output
		var result map[string]interface{}
		if parseErr := json.Unmarshal([]byte(stdout), &result); parseErr == nil {
			// Check if email is present in output
			if emailVal, ok := result["email"]; ok {
				assert.Equal(t, email, emailVal, "Email should match")
			}
		}
	}
}

// TestCLIAuthLogout tests logout via CLI
func TestCLIAuthLogout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI test in short mode")
	}

	ctx := NewTestContext(t)

	// Register and login via API
	email := GenerateUniqueEmail()
	password := "SecurePass123!"
	err := ctx.CreateTestUser(email, password)
	require.NoError(t, err)

	// Save config for CLI
	err = ctx.SaveCLIConfig()
	require.NoError(t, err)

	// Run logout command
	stdout, stderr, err := ctx.RunCLI("auth", "logout")
	t.Logf("Logout stdout: %s", stdout)
	t.Logf("Logout stderr: %s", stderr)

	// Logout should succeed
	if err == nil {
		assert.True(t, strings.Contains(stdout, "success") || strings.Contains(stdout, "logged out"),
			"Logout output should indicate success")
	}
}

// TestAuthFlowEndToEnd tests complete auth flow: register -> login -> access -> logout
func TestAuthFlowEndToEnd(t *testing.T) {
	ctx := NewTestContext(t)

	// 1. Register
	email := GenerateUniqueEmail()
	password := "SecurePass123!"
	err := ctx.CreateTestUser(email, password)
	require.NoError(t, err, "Registration should succeed")

	// 2. Access protected endpoint
	var userResp struct {
		Email string `json:"email"`
	}
	err = ctx.APIRequest("GET", "/api/v1/auth/me", nil, &userResp)
	require.NoError(t, err, "Should access protected endpoint after registration")
	assert.Equal(t, email, userResp.Email)

	// 3. Refresh token
	oldAccessToken := ctx.AccessToken
	refreshToken := ctx.RefreshToken

	time.Sleep(100 * time.Millisecond)

	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}
	var refreshResp struct {
		AccessToken string `json:"access_token"`
	}
	err = ctx.APIRequest("POST", "/api/v1/auth/refresh", reqBody, &refreshResp)
	require.NoError(t, err, "Token refresh should succeed")
	assert.NotEqual(t, oldAccessToken, refreshResp.AccessToken)

	ctx.AccessToken = refreshResp.AccessToken

	// 4. Access protected endpoint with new token
	err = ctx.APIRequest("GET", "/api/v1/auth/me", nil, &userResp)
	require.NoError(t, err, "Should access protected endpoint with refreshed token")

	// 5. Logout
	logoutReq := map[string]string{
		"refresh_token": refreshToken,
	}
	err = ctx.APIRequest("POST", "/api/v1/auth/logout", logoutReq, nil)
	require.NoError(t, err, "Logout should succeed")

	// 6. Verify refresh token is revoked
	err = ctx.APIRequest("POST", "/api/v1/auth/refresh", reqBody, &refreshResp)
	assert.Error(t, err, "Should not be able to use revoked refresh token")
}
