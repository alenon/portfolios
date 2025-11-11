package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/dto"
	"github.com/lenon/portfolios/internal/models"
)

// Mock services for testing
type mockAuthService struct {
	registerFunc       func(email, password string) (*models.User, string, string, error)
	loginFunc          func(email, password string, rememberMe bool) (*models.User, string, string, error)
	refreshAccessToken func(refreshToken string) (string, error)
	logoutFunc         func(refreshToken string) error
}

func (m *mockAuthService) Register(email, password string) (*models.User, string, string, error) {
	if m.registerFunc != nil {
		return m.registerFunc(email, password)
	}
	return nil, "", "", errors.New("not implemented")
}

func (m *mockAuthService) Login(email, password string, rememberMe bool) (*models.User, string, string, error) {
	if m.loginFunc != nil {
		return m.loginFunc(email, password, rememberMe)
	}
	return nil, "", "", errors.New("not implemented")
}

func (m *mockAuthService) RefreshAccessToken(refreshToken string) (string, error) {
	if m.refreshAccessToken != nil {
		return m.refreshAccessToken(refreshToken)
	}
	return "", errors.New("not implemented")
}

func (m *mockAuthService) Logout(refreshToken string) error {
	if m.logoutFunc != nil {
		return m.logoutFunc(refreshToken)
	}
	return errors.New("not implemented")
}

type mockPasswordResetService struct {
	initiateResetFunc func(email string) error
	resetPasswordFunc func(token, newPassword string) error
}

func (m *mockPasswordResetService) InitiateReset(email string) error {
	if m.initiateResetFunc != nil {
		return m.initiateResetFunc(email)
	}
	return errors.New("not implemented")
}

func (m *mockPasswordResetService) ValidateResetToken(token string) (*models.PasswordResetToken, error) {
	return nil, errors.New("not implemented")
}

func (m *mockPasswordResetService) ResetPassword(token, newPassword string) error {
	if m.resetPasswordFunc != nil {
		return m.resetPasswordFunc(token, newPassword)
	}
	return errors.New("not implemented")
}

type mockUserRepository struct {
	findByIDFunc func(id string) (*models.User, error)
}

func (m *mockUserRepository) Create(user *models.User) error {
	return nil
}

func (m *mockUserRepository) FindByEmail(email string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepository) FindByID(id string) (*models.User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(id)
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) UpdateLastLogin(id string) error {
	return nil
}

func (m *mockUserRepository) UpdatePassword(id string, passwordHash string) error {
	return nil
}

// Test 1: Successful user registration
func TestRegisterSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock services
	mockAuth := &mockAuthService{
		registerFunc: func(email, password string) (*models.User, string, string, error) {
			user := &models.User{
				ID:        uuid.New(),
				Email:     email,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			return user, "access_token_123", "refresh_token_123", nil
		},
	}
	mockPasswordReset := &mockPasswordResetService{}
	mockUserRepo := &mockUserRepository{}

	handler := NewAuthHandler(mockAuth, mockPasswordReset, mockUserRepo, 3600)

	// Create test request
	reqBody := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "SecurePass123",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.Register(c)

	// Assert response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response dto.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.User.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", response.User.Email)
	}
	if response.AccessToken != "access_token_123" {
		t.Errorf("Expected access token, got '%s'", response.AccessToken)
	}
}

// Test 2: Registration with duplicate email
func TestRegisterDuplicateEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock services
	mockAuth := &mockAuthService{
		registerFunc: func(email, password string) (*models.User, string, string, error) {
			return nil, "", "", errors.New("user with email test@example.com already exists")
		},
	}
	mockPasswordReset := &mockPasswordResetService{}
	mockUserRepo := &mockUserRepository{}

	handler := NewAuthHandler(mockAuth, mockPasswordReset, mockUserRepo, 3600)

	// Create test request
	reqBody := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "SecurePass123",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.Register(c)

	// Assert response
	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", w.Code)
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Code != "EMAIL_ALREADY_EXISTS" {
		t.Errorf("Expected error code 'EMAIL_ALREADY_EXISTS', got '%s'", response.Code)
	}
}

// Test 3: Successful login
func TestLoginSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock services
	now := time.Now().UTC()
	mockAuth := &mockAuthService{
		loginFunc: func(email, password string, rememberMe bool) (*models.User, string, string, error) {
			user := &models.User{
				ID:          uuid.New(),
				Email:       email,
				CreatedAt:   now,
				UpdatedAt:   now,
				LastLoginAt: &now,
			}
			return user, "access_token_456", "refresh_token_456", nil
		},
	}
	mockPasswordReset := &mockPasswordResetService{}
	mockUserRepo := &mockUserRepository{}

	handler := NewAuthHandler(mockAuth, mockPasswordReset, mockUserRepo, 3600)

	// Create test request
	reqBody := dto.LoginRequest{
		Email:      "test@example.com",
		Password:   "SecurePass123",
		RememberMe: false,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.Login(c)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response dto.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.User.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", response.User.Email)
	}
	if response.AccessToken != "access_token_456" {
		t.Errorf("Expected access token, got '%s'", response.AccessToken)
	}
}

// Test 4: Login with invalid credentials
func TestLoginInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock services
	mockAuth := &mockAuthService{
		loginFunc: func(email, password string, rememberMe bool) (*models.User, string, string, error) {
			return nil, "", "", errors.New("invalid email or password")
		},
	}
	mockPasswordReset := &mockPasswordResetService{}
	mockUserRepo := &mockUserRepository{}

	handler := NewAuthHandler(mockAuth, mockPasswordReset, mockUserRepo, 3600)

	// Create test request
	reqBody := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.Login(c)

	// Assert response
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Code != "INVALID_CREDENTIALS" {
		t.Errorf("Expected error code 'INVALID_CREDENTIALS', got '%s'", response.Code)
	}
}

// Test 5: Successful token refresh
func TestRefreshTokenSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock services
	mockAuth := &mockAuthService{
		refreshAccessToken: func(refreshToken string) (string, error) {
			return "new_access_token_789", nil
		},
	}
	mockPasswordReset := &mockPasswordResetService{}
	mockUserRepo := &mockUserRepository{}

	handler := NewAuthHandler(mockAuth, mockPasswordReset, mockUserRepo, 3600)

	// Create test request
	reqBody := dto.RefreshRequest{
		RefreshToken: "valid_refresh_token",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.RefreshToken(c)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response dto.RefreshResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.AccessToken != "new_access_token_789" {
		t.Errorf("Expected new access token, got '%s'", response.AccessToken)
	}
	if response.ExpiresIn != 3600 {
		t.Errorf("Expected expires_in 3600, got %d", response.ExpiresIn)
	}
}

// Test 6: Refresh token with invalid token
func TestRefreshTokenInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock services
	mockAuth := &mockAuthService{
		refreshAccessToken: func(refreshToken string) (string, error) {
			return "", errors.New("invalid refresh token")
		},
	}
	mockPasswordReset := &mockPasswordResetService{}
	mockUserRepo := &mockUserRepository{}

	handler := NewAuthHandler(mockAuth, mockPasswordReset, mockUserRepo, 3600)

	// Create test request
	reqBody := dto.RefreshRequest{
		RefreshToken: "invalid_token",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.RefreshToken(c)

	// Assert response
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Code != "INVALID_REFRESH_TOKEN" {
		t.Errorf("Expected error code 'INVALID_REFRESH_TOKEN', got '%s'", response.Code)
	}
}

// Test 7: Successful logout
func TestLogoutSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock services
	mockAuth := &mockAuthService{
		logoutFunc: func(refreshToken string) error {
			return nil
		},
	}
	mockPasswordReset := &mockPasswordResetService{}
	mockUserRepo := &mockUserRepository{}

	handler := NewAuthHandler(mockAuth, mockPasswordReset, mockUserRepo, 3600)

	// Create test request
	reqBody := dto.LogoutRequest{
		RefreshToken: "valid_refresh_token",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/auth/logout", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.Logout(c)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response dto.MessageResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Message != "Logged out successfully" {
		t.Errorf("Expected success message, got '%s'", response.Message)
	}
}

// Test 8: Password reset request (always returns success)
func TestForgotPasswordAlwaysReturnsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock services
	mockAuth := &mockAuthService{}
	mockPasswordReset := &mockPasswordResetService{
		initiateResetFunc: func(email string) error {
			// Return nil even if user doesn't exist (security)
			return nil
		},
	}
	mockUserRepo := &mockUserRepository{}

	handler := NewAuthHandler(mockAuth, mockPasswordReset, mockUserRepo, 3600)

	// Create test request
	reqBody := dto.ForgotPasswordRequest{
		Email: "nonexistent@example.com",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/auth/forgot-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.ForgotPassword(c)

	// Assert response - should always return 200 to prevent email enumeration
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response dto.MessageResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	expectedMsg := "If an account exists with this email, a password reset link has been sent"
	if response.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, response.Message)
	}
}

// Test 9: Get current user success
func TestGetCurrentUserSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New().String()
	now := time.Now().UTC()

	mockAuth := &mockAuthService{}
	mockPasswordReset := &mockPasswordResetService{}
	mockUserRepo := &mockUserRepository{
		findByIDFunc: func(id string) (*models.User, error) {
			if id == userID {
				return &models.User{
					ID:        uuid.MustParse(userID),
					Email:     "user@example.com",
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			}
			return nil, errors.New("user not found")
		},
	}

	handler := NewAuthHandler(mockAuth, mockPasswordReset, mockUserRepo, 3600)

	// Create test request
	req, _ := http.NewRequest("GET", "/api/auth/me", nil)

	// Record response
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Simulate middleware setting user ID
	c.Set("user_id", userID)

	// Execute handler
	handler.GetCurrentUser(c)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response dto.UserResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Email != "user@example.com" {
		t.Errorf("Expected email 'user@example.com', got '%s'", response.Email)
	}
}

// Test 10: Get current user unauthorized
func TestGetCurrentUserUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuth := &mockAuthService{}
	mockPasswordReset := &mockPasswordResetService{}
	mockUserRepo := &mockUserRepository{}

	handler := NewAuthHandler(mockAuth, mockPasswordReset, mockUserRepo, 3600)

	// Create test request
	req, _ := http.NewRequest("GET", "/api/auth/me", nil)

	// Record response
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user ID to simulate missing auth

	// Execute handler
	handler.GetCurrentUser(c)

	// Assert response
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Code != "NOT_AUTHENTICATED" {
		t.Errorf("Expected error code 'NOT_AUTHENTICATED', got '%s'", response.Code)
	}
}

// Test 11: Get current user not found
func TestGetCurrentUserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuth := &mockAuthService{}
	mockPasswordReset := &mockPasswordResetService{}
	mockUserRepo := &mockUserRepository{
		findByIDFunc: func(id string) (*models.User, error) {
			return nil, errors.New("user not found")
		},
	}

	handler := NewAuthHandler(mockAuth, mockPasswordReset, mockUserRepo, 3600)

	// Create test request
	req, _ := http.NewRequest("GET", "/api/auth/me", nil)

	// Record response
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uuid.New().String())

	// Execute handler
	handler.GetCurrentUser(c)

	// Assert response
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Code != "USER_NOT_FOUND" {
		t.Errorf("Expected error code 'USER_NOT_FOUND', got '%s'", response.Code)
	}
}

// Test 12: Reset password success
func TestResetPasswordSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuth := &mockAuthService{}
	mockPasswordReset := &mockPasswordResetService{
		resetPasswordFunc: func(token, newPassword string) error {
			if token == "valid_token" {
				return nil
			}
			return errors.New("invalid token")
		},
	}
	mockUserRepo := &mockUserRepository{}

	handler := NewAuthHandler(mockAuth, mockPasswordReset, mockUserRepo, 3600)

	// Create test request
	reqBody := dto.ResetPasswordRequest{
		Token:       "valid_token",
		NewPassword: "NewSecurePass123",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/auth/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.ResetPassword(c)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response dto.MessageResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Message != "Password reset successfully" {
		t.Errorf("Expected success message, got '%s'", response.Message)
	}
}

// Test 13: Reset password with invalid token
func TestResetPasswordInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuth := &mockAuthService{}
	mockPasswordReset := &mockPasswordResetService{
		resetPasswordFunc: func(token, newPassword string) error {
			return errors.New("token not found")
		},
	}
	mockUserRepo := &mockUserRepository{}

	handler := NewAuthHandler(mockAuth, mockPasswordReset, mockUserRepo, 3600)

	// Create test request
	reqBody := dto.ResetPasswordRequest{
		Token:       "invalid_token",
		NewPassword: "NewSecurePass123",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/auth/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.ResetPassword(c)

	// Assert response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Code != "INVALID_RESET_TOKEN" {
		t.Errorf("Expected error code 'INVALID_RESET_TOKEN', got '%s'", response.Code)
	}
}

// Test 14: Reset password with invalid password
func TestResetPasswordInvalidPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuth := &mockAuthService{}
	mockPasswordReset := &mockPasswordResetService{
		resetPasswordFunc: func(token, newPassword string) error {
			// Service returns error with "invalid password" in message
			return errors.New("invalid password: must contain uppercase and numbers")
		},
	}
	mockUserRepo := &mockUserRepository{}

	handler := NewAuthHandler(mockAuth, mockPasswordReset, mockUserRepo, 3600)

	// Create test request - password passes gin validation (>=8 chars) but fails service validation
	reqBody := dto.ResetPasswordRequest{
		Token:       "valid_token",
		NewPassword: "weakpassword", // Passes gin min=8 but would fail service validation
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/auth/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.ResetPassword(c)

	// Assert response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Code != "INVALID_PASSWORD" {
		t.Errorf("Expected error code 'INVALID_PASSWORD', got '%s'", response.Code)
	}
}
