package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
	"github.com/lenon/portfolios/internal/utils"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(email, password string) (*models.User, string, string, error)
	Login(email, password string, rememberMe bool) (*models.User, string, string, error)
	RefreshAccessToken(refreshToken string) (string, error)
	Logout(refreshToken string) error
}

// authService implements AuthService interface
type authService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	tokenService     *TokenService
	accessDuration   time.Duration
	refreshDuration  time.Duration
	rememberMeAccessDuration  time.Duration
	rememberMeRefreshDuration time.Duration
}

// NewAuthService creates a new AuthService instance
func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	tokenService *TokenService,
	accessDuration time.Duration,
	refreshDuration time.Duration,
	rememberMeAccessDuration time.Duration,
	rememberMeRefreshDuration time.Duration,
) AuthService {
	return &authService{
		userRepo:                   userRepo,
		refreshTokenRepo:           refreshTokenRepo,
		tokenService:               tokenService,
		accessDuration:             accessDuration,
		refreshDuration:            refreshDuration,
		rememberMeAccessDuration:   rememberMeAccessDuration,
		rememberMeRefreshDuration:  rememberMeRefreshDuration,
	}
}

// Register creates a new user account and returns the user with tokens
func (s *authService) Register(email, password string) (*models.User, string, string, error) {
	// Validate email
	if email == "" {
		return nil, "", "", fmt.Errorf("email cannot be empty")
	}

	// Validate and hash password
	hashedPassword, err := utils.ValidateAndHashPassword(password)
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid password: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(email)
	if err == nil && existingUser != nil {
		return nil, "", "", fmt.Errorf("user with email %s already exists", email)
	}

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", "", fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens (use default durations, not remember me)
	accessToken, err := s.tokenService.GenerateAccessToken(user.ID.String(), s.accessDuration)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(user.ID.String(), s.refreshDuration)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Hash and store refresh token
	tokenHash := hashToken(refreshToken)
	refreshTokenModel := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(s.refreshDuration),
		CreatedAt: time.Now().UTC(),
	}

	if err := s.refreshTokenRepo.Create(refreshTokenModel); err != nil {
		return nil, "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Update last login
	now := time.Now().UTC()
	user.LastLoginAt = &now

	return user, accessToken, refreshToken, nil
}

// Login authenticates a user and returns the user with tokens
func (s *authService) Login(email, password string, rememberMe bool) (*models.User, string, string, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid email or password")
	}

	// Verify password
	if err := utils.CheckPassword(password, user.PasswordHash); err != nil {
		return nil, "", "", fmt.Errorf("invalid email or password")
	}

	// Determine token durations based on remember me flag
	accessDuration := s.accessDuration
	refreshDuration := s.refreshDuration

	if rememberMe {
		accessDuration = s.rememberMeAccessDuration
		refreshDuration = s.rememberMeRefreshDuration
	}

	// Generate tokens
	accessToken, err := s.tokenService.GenerateAccessToken(user.ID.String(), accessDuration)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(user.ID.String(), refreshDuration)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Hash and store refresh token
	tokenHash := hashToken(refreshToken)
	refreshTokenModel := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(refreshDuration),
		CreatedAt: time.Now().UTC(),
	}

	if err := s.refreshTokenRepo.Create(refreshTokenModel); err != nil {
		return nil, "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Update last login timestamp
	if err := s.userRepo.UpdateLastLogin(user.ID.String()); err != nil {
		// Log error but don't fail login
		fmt.Printf("Warning: failed to update last login: %v\n", err)
	}

	// Refresh user to get updated last login
	user, _ = s.userRepo.FindByID(user.ID.String())

	return user, accessToken, refreshToken, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func (s *authService) RefreshAccessToken(refreshToken string) (string, error) {
	if refreshToken == "" {
		return "", fmt.Errorf("refresh token cannot be empty")
	}

	// Validate refresh token
	token, err := s.tokenService.ValidateToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Extract user ID
	userID, err := s.tokenService.ExtractUserID(token)
	if err != nil {
		return "", fmt.Errorf("failed to extract user ID: %w", err)
	}

	// Verify refresh token exists in database and is not revoked
	tokenHash := hashToken(refreshToken)
	storedToken, err := s.refreshTokenRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return "", fmt.Errorf("refresh token not found: %w", err)
	}

	// Check if token is valid (not expired or revoked)
	if !storedToken.IsValid() {
		return "", fmt.Errorf("refresh token is expired or revoked")
	}

	// Generate new access token (use default duration)
	accessToken, err := s.tokenService.GenerateAccessToken(userID, s.accessDuration)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, nil
}

// Logout revokes the refresh token
func (s *authService) Logout(refreshToken string) error {
	if refreshToken == "" {
		return fmt.Errorf("refresh token cannot be empty")
	}

	// Hash the token
	tokenHash := hashToken(refreshToken)

	// Revoke the token
	if err := s.refreshTokenRepo.RevokeByTokenHash(tokenHash); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// hashToken creates a SHA-256 hash of a token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
