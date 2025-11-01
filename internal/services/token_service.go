package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenService handles JWT token generation and validation
type TokenService struct {
	secret []byte
}

// Claims represents the JWT claims structure
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// NewTokenService creates a new TokenService instance
func NewTokenService(secret string) *TokenService {
	return &TokenService{
		secret: []byte(secret),
	}
}

// GenerateAccessToken generates a new access token for the given user ID
func (ts *TokenService) GenerateAccessToken(userID string, duration time.Duration) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("user ID cannot be empty")
	}

	// Validate userID is a valid UUID
	if _, err := uuid.Parse(userID); err != nil {
		return "", fmt.Errorf("invalid user ID format: %w", err)
	}

	now := time.Now().UTC()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // Unique JWT ID to prevent token collision
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(ts.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a new refresh token for the given user ID
func (ts *TokenService) GenerateRefreshToken(userID string, duration time.Duration) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("user ID cannot be empty")
	}

	// Validate userID is a valid UUID
	if _, err := uuid.Parse(userID); err != nil {
		return "", fmt.Errorf("invalid user ID format: %w", err)
	}

	now := time.Now().UTC()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // Unique JWT ID to prevent token collision
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(ts.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the parsed token
func (ts *TokenService) ValidateToken(tokenString string) (*jwt.Token, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method is specifically HS256
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return ts.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return token, nil
}

// ExtractUserID extracts the user ID from a validated JWT token
func (ts *TokenService) ExtractUserID(token *jwt.Token) (string, error) {
	if token == nil {
		return "", fmt.Errorf("token cannot be nil")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", fmt.Errorf("invalid token claims type")
	}

	if claims.UserID == "" {
		return "", fmt.Errorf("user ID not found in token claims")
	}

	// Validate userID is a valid UUID
	if _, err := uuid.Parse(claims.UserID); err != nil {
		return "", fmt.Errorf("invalid user ID format in token: %w", err)
	}

	return claims.UserID, nil
}

// ValidateAndExtractUserID is a convenience method that validates a token and extracts the user ID
func (ts *TokenService) ValidateAndExtractUserID(tokenString string) (string, error) {
	token, err := ts.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	return ts.ExtractUserID(token)
}
