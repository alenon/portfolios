package integration

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
	"github.com/lenon/portfolios/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockEmailService implements EmailService interface for testing
type MockEmailService struct {
	SendPasswordResetEmailCalled bool
	LastEmailRecipient           string
	LastResetToken               string
	SendError                    error
}

func (m *MockEmailService) SendPasswordResetEmail(to, resetToken string) error {
	m.SendPasswordResetEmailCalled = true
	m.LastEmailRecipient = to
	m.LastResetToken = resetToken
	return m.SendError
}

// setupPasswordResetTest creates services for password reset testing
func setupPasswordResetTest(t *testing.T) (services.PasswordResetService, services.AuthService, *MockEmailService, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.User{}, &models.RefreshToken{}, &models.PasswordResetToken{})
	require.NoError(t, err)

	// Create repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)

	// Create mock email service
	mockEmailService := &MockEmailService{}

	// Create services
	tokenService := services.NewTokenService("test-secret-key-for-jwt-signing-32-chars")
	authService := services.NewAuthService(
		userRepo,
		refreshTokenRepo,
		tokenService,
		30*time.Minute,
		7*24*time.Hour,
		24*time.Hour,
		30*24*time.Hour,
	)

	passwordResetService := services.NewPasswordResetService(
		userRepo,
		passwordResetRepo,
		mockEmailService,
		1*time.Hour, // token validity duration
	)

	return passwordResetService, authService, mockEmailService, db
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// hashToken creates a SHA-256 hash of a token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// Test 1: Full password reset flow (request reset -> receive token -> reset password -> login with new password)
func TestFullPasswordResetFlow(t *testing.T) {
	passwordResetService, authService, mockEmailService, db := setupPasswordResetTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	email := "resettest@example.com"
	originalPassword := "OldPassword123"
	newPassword := "NewPassword456"

	// Step 1: Register user with original password
	user, _, _, err := authService.Register(email, originalPassword)
	require.NoError(t, err, "User registration should succeed")

	// Step 2: Request password reset
	err = passwordResetService.InitiateReset(email)
	require.NoError(t, err, "Password reset initiation should succeed")

	// Verify email service was called
	assert.True(t, mockEmailService.SendPasswordResetEmailCalled, "Email service should be called")
	assert.Equal(t, email, mockEmailService.LastEmailRecipient, "Email should be sent to correct recipient")
	assert.NotEmpty(t, mockEmailService.LastResetToken, "Reset token should be generated")

	resetToken := mockEmailService.LastResetToken

	// Step 3: Verify reset token was stored in database
	// We need to validate the token exists
	_, err = passwordResetService.ValidateResetToken(resetToken)
	require.NoError(t, err, "Reset token should be valid")

	// Step 4: Reset password using token
	err = passwordResetService.ResetPassword(resetToken, newPassword)
	require.NoError(t, err, "Password reset should succeed")

	// Step 5: Verify old password no longer works
	_, _, _, err = authService.Login(email, originalPassword, false)
	assert.Error(t, err, "Login with old password should fail")

	// Step 6: Verify new password works
	loggedInUser, accessToken, refreshToken, err := authService.Login(email, newPassword, false)
	require.NoError(t, err, "Login with new password should succeed")
	assert.NotNil(t, loggedInUser, "User should be returned")
	assert.Equal(t, user.ID, loggedInUser.ID, "User ID should match")
	assert.NotEmpty(t, accessToken, "Access token should be returned")
	assert.NotEmpty(t, refreshToken, "Refresh token should be returned")
}

// Test 2: Token expiration and single-use (expired token fails, used token cannot be reused)
func TestPasswordResetTokenExpirationAndSingleUse(t *testing.T) {
	// Setup test with very short token validity (for expiration test)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	err = db.AutoMigrate(&models.User{}, &models.RefreshToken{}, &models.PasswordResetToken{})
	require.NoError(t, err)

	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	mockEmailService := &MockEmailService{}
	tokenService := services.NewTokenService("test-secret-key-for-jwt-signing-32-chars")

	authService := services.NewAuthService(
		userRepo,
		refreshTokenRepo,
		tokenService,
		30*time.Minute,
		7*24*time.Hour,
		24*time.Hour,
		30*24*time.Hour,
	)

	email := "expiretest@example.com"
	password := "SecurePass123"
	newPassword := "NewPassword789"

	// Register user
	_, _, _, err = authService.Register(email, password)
	require.NoError(t, err)

	// Test 1: Expired token should fail
	t.Run("Expired token fails", func(t *testing.T) {
		// Create expired token manually
		expiredToken := generateSecureToken()
		tokenHash := hashToken(expiredToken)

		user, err := userRepo.FindByEmail(email)
		require.NoError(t, err)

		expiredResetToken := &models.PasswordResetToken{
			ID:        uuid.New(),
			UserID:    user.ID,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().UTC().Add(-1 * time.Hour), // Already expired
			CreatedAt: time.Now().UTC(),
		}

		err = passwordResetRepo.Create(expiredResetToken)
		require.NoError(t, err)

		// Try to validate expired token
		_, err = passwordResetRepo.FindByTokenHash(tokenHash)
		require.NoError(t, err) // Token exists

		// Check if token is expired
		storedToken, _ := passwordResetRepo.FindByTokenHash(tokenHash)
		isExpired := time.Now().UTC().After(storedToken.ExpiresAt)
		assert.True(t, isExpired, "Token should be expired")

		// Try to reset password with expired token
		passwordResetService := services.NewPasswordResetService(
			userRepo,
			passwordResetRepo,
			mockEmailService,
			1*time.Hour,
		)
		err = passwordResetService.ResetPassword(expiredToken, newPassword)
		assert.Error(t, err, "Password reset with expired token should fail")
		assert.Contains(t, err.Error(), "expired", "Error should mention token expiration")
	})

	// Test 2: Used token cannot be reused
	t.Run("Used token cannot be reused", func(t *testing.T) {
		// Create new password reset service with normal expiration
		passwordResetService := services.NewPasswordResetService(
			userRepo,
			passwordResetRepo,
			mockEmailService,
			1*time.Hour,
		)

		// Initiate reset
		err := passwordResetService.InitiateReset(email)
		require.NoError(t, err)

		resetToken := mockEmailService.LastResetToken
		assert.NotEmpty(t, resetToken)

		// Reset password (first use)
		err = passwordResetService.ResetPassword(resetToken, newPassword)
		require.NoError(t, err, "First password reset should succeed")

		// Try to reuse the same token (should fail)
		err = passwordResetService.ResetPassword(resetToken, "AnotherPassword123")
		assert.Error(t, err, "Reusing reset token should fail")
		// Token reuse should fail - we accept either "used" or "invalid" in the error message
		// since the IsValid() method checks both conditions
	})
}
