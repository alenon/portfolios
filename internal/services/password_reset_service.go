package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
	"github.com/lenon/portfolios/internal/utils"
)

const (
	// TokenLength is the length of the reset token in bytes (will be hex encoded to 64 chars)
	TokenLength = 32
	// TokenExpirationDuration is how long a password reset token is valid
	TokenExpirationDuration = 1 * time.Hour
)

// PasswordResetService defines the interface for password reset operations
type PasswordResetService interface {
	InitiateReset(email string) error
	ValidateResetToken(token string) (*models.PasswordResetToken, error)
	ResetPassword(token, newPassword string) error
}

// passwordResetService implements PasswordResetService interface
type passwordResetService struct {
	userRepo              repository.UserRepository
	tokenRepo             repository.PasswordResetRepository
	emailService          EmailService
	tokenValidityDuration time.Duration
}

// NewPasswordResetService creates a new PasswordResetService instance
func NewPasswordResetService(
	userRepo repository.UserRepository,
	tokenRepo repository.PasswordResetRepository,
	emailService EmailService,
	tokenValidityDuration time.Duration,
) PasswordResetService {
	return &passwordResetService{
		userRepo:              userRepo,
		tokenRepo:             tokenRepo,
		emailService:          emailService,
		tokenValidityDuration: tokenValidityDuration,
	}
}

// InitiateReset generates a password reset token and sends an email to the user
func (s *passwordResetService) InitiateReset(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Find user by email (don't reveal if user exists for security)
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		// Return success even if user doesn't exist to prevent email enumeration
		return nil
	}

	// Generate cryptographically secure random token
	tokenBytes := make([]byte, TokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// Hash token before storing in database
	tokenHash := hashResetToken(token)

	// Create password reset token record
	resetToken := &models.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(s.tokenValidityDuration),
		CreatedAt: time.Now().UTC(),
	}

	if err := s.tokenRepo.Create(resetToken); err != nil {
		return fmt.Errorf("failed to create reset token: %w", err)
	}

	// Send password reset email
	if err := s.emailService.SendPasswordResetEmail(user.Email, token); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to send password reset email: %v\n", err)
		// Still return nil to prevent email enumeration
	}

	return nil
}

// ValidateResetToken validates a password reset token and returns it if valid
func (s *passwordResetService) ValidateResetToken(token string) (*models.PasswordResetToken, error) {
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	// Hash the token
	tokenHash := hashResetToken(token)

	// Find token in database
	resetToken, err := s.tokenRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired reset token")
	}

	// Check if token is valid (not expired and not used)
	if !resetToken.IsValid() {
		return nil, fmt.Errorf("invalid or expired reset token")
	}

	return resetToken, nil
}

// ResetPassword validates the token, updates the password, and marks the token as used
func (s *passwordResetService) ResetPassword(token, newPassword string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	if newPassword == "" {
		return fmt.Errorf("new password cannot be empty")
	}

	// Validate the reset token
	resetToken, err := s.ValidateResetToken(token)
	if err != nil {
		return err
	}

	// Check if token is expired
	if time.Now().UTC().After(resetToken.ExpiresAt) {
		return fmt.Errorf("reset token has expired")
	}

	// Check if token was already used
	if resetToken.UsedAt != nil {
		return fmt.Errorf("reset token has already been used")
	}

	// Validate and hash new password
	hashedPassword, err := utils.ValidateAndHashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}

	// Update user's password
	if err := s.userRepo.UpdatePassword(resetToken.UserID.String(), hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	if err := s.tokenRepo.MarkAsUsed(resetToken.ID.String()); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// hashResetToken creates a SHA-256 hash of a reset token
func hashResetToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
