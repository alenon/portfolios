package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewTokenService(t *testing.T) {
	secret := "test-secret"
	service := NewTokenService(secret)

	assert.NotNil(t, service)
	assert.Equal(t, []byte(secret), service.secret)
}

func TestTokenService_GenerateAccessToken_Success(t *testing.T) {
	service := NewTokenService("test-secret")
	userID := uuid.New().String()
	duration := 15 * time.Minute

	token, err := service.GenerateAccessToken(userID, duration)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestTokenService_GenerateAccessToken_EmptyUserID(t *testing.T) {
	service := NewTokenService("test-secret")
	duration := 15 * time.Minute

	token, err := service.GenerateAccessToken("", duration)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "user ID cannot be empty")
}

func TestTokenService_GenerateAccessToken_InvalidUserID(t *testing.T) {
	service := NewTokenService("test-secret")
	duration := 15 * time.Minute

	token, err := service.GenerateAccessToken("invalid-uuid", duration)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "invalid user ID format")
}

func TestTokenService_GenerateRefreshToken_Success(t *testing.T) {
	service := NewTokenService("test-secret")
	userID := uuid.New().String()
	duration := 7 * 24 * time.Hour

	token, err := service.GenerateRefreshToken(userID, duration)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestTokenService_GenerateRefreshToken_EmptyUserID(t *testing.T) {
	service := NewTokenService("test-secret")
	duration := 7 * 24 * time.Hour

	token, err := service.GenerateRefreshToken("", duration)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "user ID cannot be empty")
}

func TestTokenService_GenerateRefreshToken_InvalidUserID(t *testing.T) {
	service := NewTokenService("test-secret")
	duration := 7 * 24 * time.Hour

	token, err := service.GenerateRefreshToken("not-a-uuid", duration)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "invalid user ID format")
}

func TestTokenService_ValidateToken_Success(t *testing.T) {
	service := NewTokenService("test-secret")
	userID := uuid.New().String()
	duration := 15 * time.Minute

	tokenString, err := service.GenerateAccessToken(userID, duration)
	assert.NoError(t, err)

	token, err := service.ValidateToken(tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.True(t, token.Valid)
}

func TestTokenService_ValidateToken_EmptyToken(t *testing.T) {
	service := NewTokenService("test-secret")

	token, err := service.ValidateToken("")

	assert.Error(t, err)
	assert.Nil(t, token)
	assert.Contains(t, err.Error(), "token cannot be empty")
}

func TestTokenService_ValidateToken_InvalidToken(t *testing.T) {
	service := NewTokenService("test-secret")

	token, err := service.ValidateToken("invalid.token.string")

	assert.Error(t, err)
	assert.Nil(t, token)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestTokenService_ValidateToken_ExpiredToken(t *testing.T) {
	service := NewTokenService("test-secret")
	userID := uuid.New().String()
	duration := -1 * time.Hour // Already expired

	tokenString, err := service.GenerateAccessToken(userID, duration)
	assert.NoError(t, err)

	token, err := service.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, token)
}

func TestTokenService_ValidateToken_WrongSecret(t *testing.T) {
	service1 := NewTokenService("secret1")
	service2 := NewTokenService("secret2")

	userID := uuid.New().String()
	tokenString, err := service1.GenerateAccessToken(userID, 15*time.Minute)
	assert.NoError(t, err)

	// Try to validate with different secret
	token, err := service2.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, token)
}

func TestTokenService_ValidateToken_WrongSigningMethod(t *testing.T) {
	service := NewTokenService("test-secret")

	// Create token with wrong signing method
	claims := Claims{
		UserID: uuid.New().String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims) // Using HS512 instead of HS256
	tokenString, _ := token.SignedString([]byte("test-secret"))

	validatedToken, err := service.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, validatedToken)
	assert.Contains(t, err.Error(), "unexpected signing method")
}

func TestTokenService_ExtractUserID_Success(t *testing.T) {
	service := NewTokenService("test-secret")
	userID := uuid.New().String()

	tokenString, err := service.GenerateAccessToken(userID, 15*time.Minute)
	assert.NoError(t, err)

	token, err := service.ValidateToken(tokenString)
	assert.NoError(t, err)

	extractedUserID, err := service.ExtractUserID(token)

	assert.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)
}

func TestTokenService_ExtractUserID_NilToken(t *testing.T) {
	service := NewTokenService("test-secret")

	userID, err := service.ExtractUserID(nil)

	assert.Error(t, err)
	assert.Empty(t, userID)
	assert.Contains(t, err.Error(), "token cannot be nil")
}

func TestTokenService_ExtractUserID_InvalidClaimsType(t *testing.T) {
	service := NewTokenService("test-secret")

	// Create token with wrong claims type
	token := &jwt.Token{
		Claims: jwt.MapClaims{
			"user_id": "some-user",
		},
	}

	userID, err := service.ExtractUserID(token)

	assert.Error(t, err)
	assert.Empty(t, userID)
	assert.Contains(t, err.Error(), "invalid token claims type")
}

func TestTokenService_ExtractUserID_EmptyUserID(t *testing.T) {
	service := NewTokenService("test-secret")

	// Create token with empty user ID
	claims := &Claims{
		UserID: "",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	token := &jwt.Token{
		Claims: claims,
	}

	userID, err := service.ExtractUserID(token)

	assert.Error(t, err)
	assert.Empty(t, userID)
	assert.Contains(t, err.Error(), "user ID not found in token claims")
}

func TestTokenService_ExtractUserID_InvalidUserIDFormat(t *testing.T) {
	service := NewTokenService("test-secret")

	// Create token with invalid UUID
	claims := &Claims{
		UserID: "not-a-valid-uuid",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	token := &jwt.Token{
		Claims: claims,
	}

	userID, err := service.ExtractUserID(token)

	assert.Error(t, err)
	assert.Empty(t, userID)
	assert.Contains(t, err.Error(), "invalid user ID format in token")
}

func TestTokenService_ValidateAndExtractUserID_Success(t *testing.T) {
	service := NewTokenService("test-secret")
	userID := uuid.New().String()

	tokenString, err := service.GenerateAccessToken(userID, 15*time.Minute)
	assert.NoError(t, err)

	extractedUserID, err := service.ValidateAndExtractUserID(tokenString)

	assert.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)
}

func TestTokenService_ValidateAndExtractUserID_InvalidToken(t *testing.T) {
	service := NewTokenService("test-secret")

	userID, err := service.ValidateAndExtractUserID("invalid.token")

	assert.Error(t, err)
	assert.Empty(t, userID)
}

func TestTokenService_TokenContainsExpectedClaims(t *testing.T) {
	service := NewTokenService("test-secret")
	userID := uuid.New().String()
	duration := 15 * time.Minute

	tokenString, err := service.GenerateAccessToken(userID, duration)
	assert.NoError(t, err)

	token, err := service.ValidateToken(tokenString)
	assert.NoError(t, err)

	claims, ok := token.Claims.(*Claims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims.UserID)
	assert.NotEmpty(t, claims.ID) // JWT ID should be set
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)
}

func TestTokenService_DifferentTokensHaveDifferentJTI(t *testing.T) {
	service := NewTokenService("test-secret")
	userID := uuid.New().String()
	duration := 15 * time.Minute

	token1String, err := service.GenerateAccessToken(userID, duration)
	assert.NoError(t, err)

	token2String, err := service.GenerateAccessToken(userID, duration)
	assert.NoError(t, err)

	// Tokens should be different even for same user
	assert.NotEqual(t, token1String, token2String)

	// Extract JTI from both tokens
	token1, _ := service.ValidateToken(token1String)
	token2, _ := service.ValidateToken(token2String)

	claims1 := token1.Claims.(*Claims)
	claims2 := token2.Claims.(*Claims)

	// JTI should be different
	assert.NotEqual(t, claims1.ID, claims2.ID)
}
