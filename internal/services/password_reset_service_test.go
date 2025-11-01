package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/models"
	"github.com/stretchr/testify/assert"
)

// Mock password reset repository
type mockPasswordResetRepository struct {
	tokens map[string]*models.PasswordResetToken
}

func newMockPasswordResetRepository() *mockPasswordResetRepository {
	return &mockPasswordResetRepository{
		tokens: make(map[string]*models.PasswordResetToken),
	}
}

func (m *mockPasswordResetRepository) Create(token *models.PasswordResetToken) error {
	m.tokens[token.TokenHash] = token
	return nil
}

func (m *mockPasswordResetRepository) FindByTokenHash(hash string) (*models.PasswordResetToken, error) {
	if token, exists := m.tokens[hash]; exists {
		return token, nil
	}
	return nil, fmt.Errorf("password reset token not found")
}

func (m *mockPasswordResetRepository) MarkAsUsed(id string) error {
	for _, token := range m.tokens {
		if token.ID.String() == id {
			now := time.Now().UTC()
			token.UsedAt = &now
			return nil
		}
	}
	return fmt.Errorf("password reset token not found or already used")
}

func (m *mockPasswordResetRepository) DeleteExpired() error {
	now := time.Now().UTC()
	for hash, token := range m.tokens {
		if token.ExpiresAt.Before(now) {
			delete(m.tokens, hash)
		}
	}
	return nil
}

// Mock email service
type mockEmailService struct {
	sentEmails []sentEmail
	shouldFail bool
}

type sentEmail struct {
	to    string
	token string
}

func newMockEmailService() *mockEmailService {
	return &mockEmailService{
		sentEmails: []sentEmail{},
		shouldFail: false,
	}
}

func (m *mockEmailService) SendPasswordResetEmail(to, resetToken string) error {
	if m.shouldFail {
		return fmt.Errorf("failed to send email")
	}
	m.sentEmails = append(m.sentEmails, sentEmail{to: to, token: resetToken})
	return nil
}

func TestNewPasswordResetService(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockPasswordResetRepository()
	emailService := newMockEmailService()
	duration := 1 * time.Hour

	service := NewPasswordResetService(userRepo, tokenRepo, emailService, duration)

	assert.NotNil(t, service)
}

func TestPasswordResetService_InitiateReset_Success(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockPasswordResetRepository()
	emailService := newMockEmailService()
	duration := 1 * time.Hour

	// Create a user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}
	err := userRepo.Create(user)
	assert.NoError(t, err)

	service := NewPasswordResetService(userRepo, tokenRepo, emailService, duration)

	err = service.InitiateReset("test@example.com")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(emailService.sentEmails))
	assert.Equal(t, "test@example.com", emailService.sentEmails[0].to)
	assert.NotEmpty(t, emailService.sentEmails[0].token)
	assert.Equal(t, 1, len(tokenRepo.tokens))
}

func TestPasswordResetService_InitiateReset_EmptyEmail(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockPasswordResetRepository()
	emailService := newMockEmailService()
	duration := 1 * time.Hour

	service := NewPasswordResetService(userRepo, tokenRepo, emailService, duration)

	err := service.InitiateReset("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email cannot be empty")
}

func TestPasswordResetService_InitiateReset_UserNotFound(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockPasswordResetRepository()
	emailService := newMockEmailService()
	duration := 1 * time.Hour

	service := NewPasswordResetService(userRepo, tokenRepo, emailService, duration)

	// Try to reset password for non-existent user
	err := service.InitiateReset("nonexistent@example.com")

	// Should not return error to prevent email enumeration
	assert.NoError(t, err)
	assert.Equal(t, 0, len(emailService.sentEmails))
	assert.Equal(t, 0, len(tokenRepo.tokens))
}

func TestPasswordResetService_InitiateReset_EmailServiceFails(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockPasswordResetRepository()
	emailService := newMockEmailService()
	emailService.shouldFail = true
	duration := 1 * time.Hour

	// Create a user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}
	err := userRepo.Create(user)
	assert.NoError(t, err)

	service := NewPasswordResetService(userRepo, tokenRepo, emailService, duration)

	err = service.InitiateReset("test@example.com")

	// Should not return error even if email fails (to prevent enumeration)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(tokenRepo.tokens)) // Token should still be created
}

func TestPasswordResetService_ValidateResetToken_Success(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockPasswordResetRepository()
	emailService := newMockEmailService()
	duration := 1 * time.Hour

	// Create a user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}
	err := userRepo.Create(user)
	assert.NoError(t, err)

	service := NewPasswordResetService(userRepo, tokenRepo, emailService, duration)

	// Initiate reset to get token
	err = service.InitiateReset("test@example.com")
	assert.NoError(t, err)
	token := emailService.sentEmails[0].token

	// Validate token
	resetToken, err := service.ValidateResetToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, resetToken)
	assert.Equal(t, user.ID, resetToken.UserID)
}

func TestPasswordResetService_ValidateResetToken_EmptyToken(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockPasswordResetRepository()
	emailService := newMockEmailService()
	duration := 1 * time.Hour

	service := NewPasswordResetService(userRepo, tokenRepo, emailService, duration)

	resetToken, err := service.ValidateResetToken("")

	assert.Error(t, err)
	assert.Nil(t, resetToken)
	assert.Contains(t, err.Error(), "token cannot be empty")
}

func TestPasswordResetService_ValidateResetToken_InvalidToken(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockPasswordResetRepository()
	emailService := newMockEmailService()
	duration := 1 * time.Hour

	service := NewPasswordResetService(userRepo, tokenRepo, emailService, duration)

	resetToken, err := service.ValidateResetToken("invalid-token")

	assert.Error(t, err)
	assert.Nil(t, resetToken)
	assert.Contains(t, err.Error(), "invalid or expired reset token")
}

func TestPasswordResetService_ValidateResetToken_ExpiredToken(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockPasswordResetRepository()
	emailService := newMockEmailService()
	duration := 1 * time.Hour

	// Create a user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}
	err := userRepo.Create(user)
	assert.NoError(t, err)

	service := NewPasswordResetService(userRepo, tokenRepo, emailService, duration)

	// Initiate reset
	err = service.InitiateReset("test@example.com")
	assert.NoError(t, err)
	token := emailService.sentEmails[0].token

	// Manually expire the token
	for _, resetToken := range tokenRepo.tokens {
		resetToken.ExpiresAt = time.Now().UTC().Add(-1 * time.Hour)
	}

	// Try to validate expired token
	resetToken, err := service.ValidateResetToken(token)

	assert.Error(t, err)
	assert.Nil(t, resetToken)
	assert.Contains(t, err.Error(), "invalid or expired reset token")
}

func TestPasswordResetService_ResetPassword_Success(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockPasswordResetRepository()
	emailService := newMockEmailService()
	duration := 1 * time.Hour

	// Create a user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "oldhashedpassword",
	}
	err := userRepo.Create(user)
	assert.NoError(t, err)

	service := NewPasswordResetService(userRepo, tokenRepo, emailService, duration)

	// Initiate reset
	err = service.InitiateReset("test@example.com")
	assert.NoError(t, err)
	token := emailService.sentEmails[0].token

	// Reset password
	newPassword := "NewSecure123"
	err = service.ResetPassword(token, newPassword)

	assert.NoError(t, err)

	// Verify password was updated
	updatedUser, _ := userRepo.FindByEmail("test@example.com")
	assert.NotEqual(t, "oldhashedpassword", updatedUser.PasswordHash)

	// Verify token was marked as used
	for _, resetToken := range tokenRepo.tokens {
		assert.NotNil(t, resetToken.UsedAt)
	}
}

func TestPasswordResetService_ResetPassword_EmptyToken(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockPasswordResetRepository()
	emailService := newMockEmailService()
	duration := 1 * time.Hour

	service := NewPasswordResetService(userRepo, tokenRepo, emailService, duration)

	err := service.ResetPassword("", "NewSecure123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token cannot be empty")
}

func TestPasswordResetService_ResetPassword_EmptyPassword(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockPasswordResetRepository()
	emailService := newMockEmailService()
	duration := 1 * time.Hour

	service := NewPasswordResetService(userRepo, tokenRepo, emailService, duration)

	err := service.ResetPassword("some-token", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "new password cannot be empty")
}

func TestPasswordResetService_ResetPassword_InvalidPassword(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockPasswordResetRepository()
	emailService := newMockEmailService()
	duration := 1 * time.Hour

	// Create a user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "oldhashedpassword",
	}
	err := userRepo.Create(user)
	assert.NoError(t, err)

	service := NewPasswordResetService(userRepo, tokenRepo, emailService, duration)

	// Initiate reset
	err = service.InitiateReset("test@example.com")
	assert.NoError(t, err)
	token := emailService.sentEmails[0].token

	// Try to reset with weak password
	err = service.ResetPassword(token, "weak")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid password")
}

func TestPasswordResetService_ResetPassword_ExpiredToken(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockPasswordResetRepository()
	emailService := newMockEmailService()
	duration := 1 * time.Hour

	// Create a user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "oldhashedpassword",
	}
	err := userRepo.Create(user)
	assert.NoError(t, err)

	service := NewPasswordResetService(userRepo, tokenRepo, emailService, duration)

	// Initiate reset
	err = service.InitiateReset("test@example.com")
	assert.NoError(t, err)
	token := emailService.sentEmails[0].token

	// Manually expire the token
	for _, resetToken := range tokenRepo.tokens {
		resetToken.ExpiresAt = time.Now().UTC().Add(-1 * time.Hour)
	}

	// Try to reset password with expired token
	err = service.ResetPassword(token, "NewSecure123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired reset token")
}

func TestPasswordResetService_ResetPassword_TokenAlreadyUsed(t *testing.T) {
	userRepo := newMockUserRepository()
	tokenRepo := newMockPasswordResetRepository()
	emailService := newMockEmailService()
	duration := 1 * time.Hour

	// Create a user
	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "oldhashedpassword",
	}
	err := userRepo.Create(user)
	assert.NoError(t, err)

	service := NewPasswordResetService(userRepo, tokenRepo, emailService, duration)

	// Initiate reset
	err = service.InitiateReset("test@example.com")
	assert.NoError(t, err)
	token := emailService.sentEmails[0].token

	// Reset password once
	err = service.ResetPassword(token, "NewSecure123")
	assert.NoError(t, err)

	// Try to use same token again
	err = service.ResetPassword(token, "AnotherPass123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired reset token")
}
