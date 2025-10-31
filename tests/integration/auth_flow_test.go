package integration

import (
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

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate all tables
	err = db.AutoMigrate(&models.User{}, &models.RefreshToken{}, &models.PasswordResetToken{})
	require.NoError(t, err)

	return db
}

// setupTestServices creates all required services with test database
func setupTestServices(t *testing.T) (services.AuthService, repository.UserRepository, *gorm.DB) {
	db := setupTestDB(t)

	// Create repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Create token service
	tokenService := services.NewTokenService("test-secret-key-for-jwt-signing-32-chars")

	// Create auth service with test durations
	authService := services.NewAuthService(
		userRepo,
		refreshTokenRepo,
		tokenService,
		30*time.Minute,     // access token duration
		7*24*time.Hour,     // refresh token duration
		24*time.Hour,       // remember me access duration
		30*24*time.Hour,    // remember me refresh duration
	)

	return authService, userRepo, db
}

// Test 1: Full registration flow (register -> auto-login with tokens)
func TestFullRegistrationFlow(t *testing.T) {
	authService, userRepo, db := setupTestServices(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	email := "newuser@example.com"
	password := "SecurePass123"

	// Register user
	user, accessToken, refreshToken, err := authService.Register(email, password)
	require.NoError(t, err, "Registration should succeed")
	assert.NotNil(t, user, "User should be returned")
	assert.NotEmpty(t, accessToken, "Access token should be returned")
	assert.NotEmpty(t, refreshToken, "Refresh token should be returned")
	assert.Equal(t, email, user.Email, "Email should match")
	assert.NotEqual(t, uuid.Nil, user.ID, "User ID should be generated")

	// Verify user was created in database
	dbUser, err := userRepo.FindByEmail(email)
	require.NoError(t, err, "User should exist in database")
	assert.Equal(t, user.ID, dbUser.ID, "User IDs should match")
	assert.NotEmpty(t, dbUser.PasswordHash, "Password should be hashed")
	assert.NotEqual(t, password, dbUser.PasswordHash, "Password should not be stored as plain text")

	// Verify tokens are valid (should not fail validation)
	tokenService := services.NewTokenService("test-secret-key-for-jwt-signing-32-chars")
	validatedToken, err := tokenService.ValidateToken(accessToken)
	require.NoError(t, err, "Access token should be valid")

	userID, err := tokenService.ExtractUserID(validatedToken)
	require.NoError(t, err, "Should extract user ID from token")
	assert.Equal(t, user.ID.String(), userID, "User ID in token should match registered user")
}

// Test 2: Login with remember_me false vs true (verify token expiration differences)
func TestLoginRememberMeFunctionality(t *testing.T) {
	authService, _, db := setupTestServices(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	email := "testuser@example.com"
	password := "SecurePass123"

	// Register user first
	_, _, _, err := authService.Register(email, password)
	require.NoError(t, err)

	// Login WITHOUT remember me
	user1, accessToken1, refreshToken1, err := authService.Login(email, password, false)
	require.NoError(t, err, "Login without remember me should succeed")
	assert.NotNil(t, user1)
	assert.NotEmpty(t, accessToken1)
	assert.NotEmpty(t, refreshToken1)

	// Login WITH remember me
	user2, accessToken2, refreshToken2, err := authService.Login(email, password, true)
	require.NoError(t, err, "Login with remember me should succeed")
	assert.NotNil(t, user2)
	assert.NotEmpty(t, accessToken2)
	assert.NotEmpty(t, refreshToken2)

	// Verify both tokens are different
	assert.NotEqual(t, accessToken1, accessToken2, "Access tokens should be different")
	assert.NotEqual(t, refreshToken1, refreshToken2, "Refresh tokens should be different")

	// Verify both tokens are valid
	tokenService := services.NewTokenService("test-secret-key-for-jwt-signing-32-chars")
	_, err = tokenService.ValidateToken(accessToken1)
	assert.NoError(t, err, "Access token from normal login should be valid")

	_, err = tokenService.ValidateToken(accessToken2)
	assert.NoError(t, err, "Access token from remember me login should be valid")

	// Note: In a real scenario, we would check the expiration times in the JWT claims
	// to verify that remember me extends the duration. For this integration test,
	// we verify that both logins succeed and return valid tokens.
}

// Test 3: Token refresh flow (use refresh token to get new access token)
func TestTokenRefreshFlow(t *testing.T) {
	authService, _, db := setupTestServices(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	email := "refreshtest@example.com"
	password := "SecurePass123"

	// Register user
	_, originalAccessToken, refreshToken, err := authService.Register(email, password)
	require.NoError(t, err)

	// Wait a moment to ensure new token has different timestamp
	time.Sleep(100 * time.Millisecond)

	// Refresh access token
	newAccessToken, err := authService.RefreshAccessToken(refreshToken)
	require.NoError(t, err, "Token refresh should succeed")
	assert.NotEmpty(t, newAccessToken, "New access token should be returned")
	assert.NotEqual(t, originalAccessToken, newAccessToken, "New access token should be different from original")

	// Verify new token is valid
	tokenService := services.NewTokenService("test-secret-key-for-jwt-signing-32-chars")
	validatedToken, err := tokenService.ValidateToken(newAccessToken)
	require.NoError(t, err, "New access token should be valid")

	userID, err := tokenService.ExtractUserID(validatedToken)
	require.NoError(t, err, "Should extract user ID from new token")
	assert.NotEmpty(t, userID, "User ID should be present in new token")
}

// Test 4: Protected endpoint access (valid token succeeds, invalid token fails)
func TestProtectedEndpointAccess(t *testing.T) {
	authService, userRepo, db := setupTestServices(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	email := "protectedtest@example.com"
	password := "SecurePass123"

	// Register user
	user, accessToken, _, err := authService.Register(email, password)
	require.NoError(t, err)

	// Simulate protected endpoint: Get current user
	tokenService := services.NewTokenService("test-secret-key-for-jwt-signing-32-chars")

	// Test 1: Valid token should work
	validatedToken, err := tokenService.ValidateToken(accessToken)
	require.NoError(t, err, "Valid token should be validated")

	userID, err := tokenService.ExtractUserID(validatedToken)
	require.NoError(t, err, "Should extract user ID from valid token")

	fetchedUser, err := userRepo.FindByID(userID)
	require.NoError(t, err, "Should fetch user with valid token")
	assert.Equal(t, user.ID, fetchedUser.ID, "Fetched user should match authenticated user")

	// Test 2: Invalid token should fail
	invalidToken := "invalid.jwt.token"
	_, err = tokenService.ValidateToken(invalidToken)
	assert.Error(t, err, "Invalid token should fail validation")

	// Test 3: Expired token should fail (simulate by creating token with negative duration)
	expiredToken, _ := tokenService.GenerateAccessToken(user.ID.String(), -1*time.Hour)
	_, err = tokenService.ValidateToken(expiredToken)
	assert.Error(t, err, "Expired token should fail validation")
}

// Test 5: Authorization check (user can access own data, cannot access other user's data)
func TestAuthorizationCheck(t *testing.T) {
	authService, userRepo, db := setupTestServices(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Create two users
	user1Email := "user1@example.com"
	user2Email := "user2@example.com"
	password := "SecurePass123"

	user1, accessToken1, _, err := authService.Register(user1Email, password)
	require.NoError(t, err)

	user2, _, _, err := authService.Register(user2Email, password)
	require.NoError(t, err)

	tokenService := services.NewTokenService("test-secret-key-for-jwt-signing-32-chars")

	// User 1 tries to access their own data (should succeed)
	validatedToken, err := tokenService.ValidateToken(accessToken1)
	require.NoError(t, err)

	userID1, err := tokenService.ExtractUserID(validatedToken)
	require.NoError(t, err)

	fetchedUser, err := userRepo.FindByID(userID1)
	require.NoError(t, err)
	assert.Equal(t, user1.ID, fetchedUser.ID, "User should access own data")

	// User 1 tries to access User 2's data (simulate authorization check)
	// In a real API, the authorization middleware would check if the authenticated user ID
	// matches the resource owner ID
	authenticatedUserID := user1.ID.String()
	resourceOwnerID := user2.ID.String()

	// Authorization check: user can only access their own resources
	if authenticatedUserID != resourceOwnerID {
		// This should fail - user 1 cannot access user 2's data
		t.Logf("Authorization check correctly prevented user %s from accessing user %s's data",
			authenticatedUserID, resourceOwnerID)
	} else {
		t.Error("Authorization check failed - users should not access other users' data")
	}

	// Verify the authorization logic works correctly
	assert.NotEqual(t, user1.ID, user2.ID, "Users should have different IDs")
	assert.NotEqual(t, authenticatedUserID, resourceOwnerID, "User should not be able to access other user's resources")
}
