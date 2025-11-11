package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/models"
)

// Mock repositories for testing
type mockUserRepository struct {
	users map[string]*models.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*models.User),
	}
}

func (m *mockUserRepository) Create(user *models.User) error {
	if _, exists := m.users[user.Email]; exists {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepository) FindByEmail(email string) (*models.User, error) {
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, fmt.Errorf("user not found with email: %s", email)
}

func (m *mockUserRepository) FindByID(id string) (*models.User, error) {
	for _, user := range m.users {
		if user.ID.String() == id {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found with id: %s", id)
}

func (m *mockUserRepository) UpdateLastLogin(id string) error {
	user, err := m.FindByID(id)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	user.LastLoginAt = &now
	return nil
}

func (m *mockUserRepository) UpdatePassword(id string, passwordHash string) error {
	user, err := m.FindByID(id)
	if err != nil {
		return err
	}
	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now().UTC()
	return nil
}

type mockRefreshTokenRepository struct {
	tokens map[string]*models.RefreshToken
}

func newMockRefreshTokenRepository() *mockRefreshTokenRepository {
	return &mockRefreshTokenRepository{
		tokens: make(map[string]*models.RefreshToken),
	}
}

func (m *mockRefreshTokenRepository) Create(token *models.RefreshToken) error {
	m.tokens[token.TokenHash] = token
	return nil
}

func (m *mockRefreshTokenRepository) FindByTokenHash(hash string) (*models.RefreshToken, error) {
	if token, exists := m.tokens[hash]; exists {
		return token, nil
	}
	return nil, fmt.Errorf("refresh token not found")
}

func (m *mockRefreshTokenRepository) RevokeByUserID(userID string) error {
	uid, _ := uuid.Parse(userID)
	now := time.Now().UTC()
	for _, token := range m.tokens {
		if token.UserID == uid && token.RevokedAt == nil {
			token.RevokedAt = &now
		}
	}
	return nil
}

func (m *mockRefreshTokenRepository) RevokeByTokenHash(hash string) error {
	if token, exists := m.tokens[hash]; exists {
		now := time.Now().UTC()
		token.RevokedAt = &now
		return nil
	}
	return fmt.Errorf("refresh token not found or already revoked")
}

func (m *mockRefreshTokenRepository) DeleteExpired() error {
	now := time.Now().UTC()
	for hash, token := range m.tokens {
		if token.ExpiresAt.Before(now) {
			delete(m.tokens, hash)
		}
	}
	return nil
}

// Test 1: User registration creates user and returns tokens
func TestAuthService_Register_Success(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockRefreshTokenRepository()
	tokenService := NewTokenService("test-secret-key-for-jwt-signing")

	authService := NewAuthService(
		userRepo,
		tokenRepo,
		tokenService,
		30*time.Minute,
		7*24*time.Hour,
		24*time.Hour,
		30*24*time.Hour,
	)

	email := "test@example.com"
	password := "SecurePass123"

	user, accessToken, refreshToken, err := authService.Register(email, password)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if user == nil {
		t.Fatal("Expected user to be returned")
	}

	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}

	if accessToken == "" {
		t.Error("Expected access token to be returned")
	}

	if refreshToken == "" {
		t.Error("Expected refresh token to be returned")
	}
}

// Test 2: Login with correct password succeeds
func TestAuthService_Login_CorrectPassword(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockRefreshTokenRepository()
	tokenService := NewTokenService("test-secret-key-for-jwt-signing")

	authService := NewAuthService(
		userRepo,
		tokenRepo,
		tokenService,
		30*time.Minute,
		7*24*time.Hour,
		24*time.Hour,
		30*24*time.Hour,
	)

	email := "test@example.com"
	password := "SecurePass123"

	// First register a user
	_, _, _, err := authService.Register(email, password)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Now login
	user, accessToken, refreshToken, err := authService.Login(email, password, false)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if user == nil {
		t.Fatal("Expected user to be returned")
	}

	if accessToken == "" {
		t.Error("Expected access token to be returned")
	}

	if refreshToken == "" {
		t.Error("Expected refresh token to be returned")
	}
}

// Test 3: Login with wrong password fails
func TestAuthService_Login_WrongPassword(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockRefreshTokenRepository()
	tokenService := NewTokenService("test-secret-key-for-jwt-signing")

	authService := NewAuthService(
		userRepo,
		tokenRepo,
		tokenService,
		30*time.Minute,
		7*24*time.Hour,
		24*time.Hour,
		30*24*time.Hour,
	)

	email := "test@example.com"
	password := "SecurePass123"
	wrongPassword := "WrongPass123"

	// First register a user
	_, _, _, err := authService.Register(email, password)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Try to login with wrong password
	user, accessToken, refreshToken, err := authService.Login(email, wrongPassword, false)

	if err == nil {
		t.Fatal("Expected error for wrong password, got nil")
	}

	if user != nil {
		t.Error("Expected no user to be returned for wrong password")
	}

	if accessToken != "" {
		t.Error("Expected no access token for wrong password")
	}

	if refreshToken != "" {
		t.Error("Expected no refresh token for wrong password")
	}
}

// Test 4: Token generation and validation
func TestAuthService_TokenGeneration(t *testing.T) {
	tokenService := NewTokenService("test-secret-key-for-jwt-signing")

	userID := uuid.New().String()
	duration := 30 * time.Minute

	// Generate access token
	accessToken, err := tokenService.GenerateAccessToken(userID, duration)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	if accessToken == "" {
		t.Fatal("Expected access token to be generated")
	}

	// Validate token
	token, err := tokenService.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Extract user ID
	extractedUserID, err := tokenService.ExtractUserID(token)
	if err != nil {
		t.Fatalf("Failed to extract user ID: %v", err)
	}

	if extractedUserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, extractedUserID)
	}
}

// Test 5: Duplicate email registration fails
func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockRefreshTokenRepository()
	tokenService := NewTokenService("test-secret-key-for-jwt-signing")

	authService := NewAuthService(
		userRepo,
		tokenRepo,
		tokenService,
		30*time.Minute,
		7*24*time.Hour,
		24*time.Hour,
		30*24*time.Hour,
	)

	email := "test@example.com"
	password := "SecurePass123"

	// First registration
	_, _, _, err := authService.Register(email, password)
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	// Second registration with same email
	user, accessToken, refreshToken, err := authService.Register(email, password)

	if err == nil {
		t.Fatal("Expected error for duplicate email, got nil")
	}

	if user != nil {
		t.Error("Expected no user for duplicate registration")
	}

	if accessToken != "" {
		t.Error("Expected no access token for duplicate registration")
	}

	if refreshToken != "" {
		t.Error("Expected no refresh token for duplicate registration")
	}
}

// Test 6: Refresh token generates new access token
func TestAuthService_RefreshAccessToken(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockRefreshTokenRepository()
	tokenService := NewTokenService("test-secret-key-for-jwt-signing")

	authService := NewAuthService(
		userRepo,
		tokenRepo,
		tokenService,
		30*time.Minute,
		7*24*time.Hour,
		24*time.Hour,
		30*24*time.Hour,
	)

	email := "test@example.com"
	password := "SecurePass123"

	// Register and login
	_, _, refreshToken, err := authService.Register(email, password)
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}

	// Refresh access token
	newAccessToken, err := authService.RefreshAccessToken(refreshToken)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if newAccessToken == "" {
		t.Error("Expected new access token to be returned")
	}
}

// Test 7: Logout revokes refresh token
func TestAuthService_Logout(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockRefreshTokenRepository()
	tokenService := NewTokenService("test-secret-key-for-jwt-signing")

	authService := NewAuthService(
		userRepo,
		tokenRepo,
		tokenService,
		30*time.Minute,
		7*24*time.Hour,
		24*time.Hour,
		30*24*time.Hour,
	)

	email := "test@example.com"
	password := "SecurePass123"

	// Register
	_, _, refreshToken, err := authService.Register(email, password)
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}

	// Logout
	err = authService.Logout(refreshToken)
	if err != nil {
		t.Fatalf("Expected no error on logout, got: %v", err)
	}

	// Try to refresh with revoked token
	_, err = authService.RefreshAccessToken(refreshToken)
	if err == nil {
		t.Error("Expected error when using revoked token, got nil")
	}
}

// Test 8: Login with non-existent user
func TestAuthService_Login_UserNotFound(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockRefreshTokenRepository()
	tokenService := NewTokenService("test-secret-key-for-jwt-signing")

	authService := NewAuthService(
		userRepo,
		tokenRepo,
		tokenService,
		30*time.Minute,
		7*24*time.Hour,
		24*time.Hour,
		30*24*time.Hour,
	)

	// Try to login with non-existent user
	user, accessToken, refreshToken, err := authService.Login("nonexistent@example.com", "password", false)

	if err == nil {
		t.Fatal("Expected error for non-existent user, got nil")
	}

	if user != nil {
		t.Error("Expected no user to be returned")
	}

	if accessToken != "" {
		t.Error("Expected no access token")
	}

	if refreshToken != "" {
		t.Error("Expected no refresh token")
	}
}

// Test 9: Login with remember me flag
func TestAuthService_Login_RememberMe(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockRefreshTokenRepository()
	tokenService := NewTokenService("test-secret-key-for-jwt-signing")

	authService := NewAuthService(
		userRepo,
		tokenRepo,
		tokenService,
		30*time.Minute,
		7*24*time.Hour,
		24*time.Hour,
		30*24*time.Hour,
	)

	email := "test@example.com"
	password := "SecurePass123"

	// Register user
	_, _, _, err := authService.Register(email, password)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Login with remember me
	user, accessToken, refreshToken, err := authService.Login(email, password, true)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if user == nil {
		t.Fatal("Expected user to be returned")
	}

	if accessToken == "" {
		t.Error("Expected access token")
	}

	if refreshToken == "" {
		t.Error("Expected refresh token")
	}
}

// Test 10: RefreshAccessToken with invalid token
func TestAuthService_RefreshAccessToken_InvalidToken(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockRefreshTokenRepository()
	tokenService := NewTokenService("test-secret-key-for-jwt-signing")

	authService := NewAuthService(
		userRepo,
		tokenRepo,
		tokenService,
		30*time.Minute,
		7*24*time.Hour,
		24*time.Hour,
		30*24*time.Hour,
	)

	// Try to refresh with invalid token
	_, err := authService.RefreshAccessToken("invalid_token")

	if err == nil {
		t.Fatal("Expected error for invalid token, got nil")
	}
}

// Test 11: Register validates password
func TestAuthService_Register_WeakPassword(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockRefreshTokenRepository()
	tokenService := NewTokenService("test-secret-key-for-jwt-signing")

	authService := NewAuthService(
		userRepo,
		tokenRepo,
		tokenService,
		30*time.Minute,
		7*24*time.Hour,
		24*time.Hour,
		30*24*time.Hour,
	)

	// Try to register with weak password
	user, accessToken, refreshToken, err := authService.Register("test@example.com", "weak")

	if err == nil {
		t.Fatal("Expected error for weak password, got nil")
	}

	if user != nil {
		t.Error("Expected no user for weak password")
	}

	if accessToken != "" {
		t.Error("Expected no access token for weak password")
	}

	if refreshToken != "" {
		t.Error("Expected no refresh token for weak password")
	}
}

// Test 12: Logout with invalid token
func TestAuthService_Logout_InvalidToken(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockRefreshTokenRepository()
	tokenService := NewTokenService("test-secret-key-for-jwt-signing")

	authService := NewAuthService(
		userRepo,
		tokenRepo,
		tokenService,
		30*time.Minute,
		7*24*time.Hour,
		24*time.Hour,
		30*24*time.Hour,
	)

	// Try to logout with invalid token - should error
	err := authService.Logout("invalid_token")

	// Logout fails on invalid tokens
	if err == nil {
		t.Fatal("Expected error on logout with invalid token, got nil")
	}
}
