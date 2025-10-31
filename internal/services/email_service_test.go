package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmailService(t *testing.T) {
	service := NewEmailService("smtp.example.com", 587, "user@example.com", "password", "noreply@example.com")

	assert.NotNil(t, service)
}

func TestEmailService_SendPasswordResetEmail_EmptyRecipient(t *testing.T) {
	service := NewEmailService("smtp.example.com", 587, "user@example.com", "password", "noreply@example.com")

	err := service.SendPasswordResetEmail("", "reset-token-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recipient email cannot be empty")
}

func TestEmailService_SendPasswordResetEmail_EmptyToken(t *testing.T) {
	service := NewEmailService("smtp.example.com", 587, "user@example.com", "password", "noreply@example.com")

	err := service.SendPasswordResetEmail("user@example.com", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reset token cannot be empty")
}

func TestEmailService_SendPasswordResetEmail_InvalidSMTP(t *testing.T) {
	// Use invalid SMTP server that will fail to connect
	service := NewEmailService("invalid-smtp-server-that-does-not-exist.local", 587, "user@example.com", "password", "noreply@example.com")

	err := service.SendPasswordResetEmail("user@example.com", "reset-token-123")

	// Should return error since SMTP server doesn't exist
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send email")
}

// Note: Testing actual email sending would require a real SMTP server or mock server
// For comprehensive testing, you would typically use a mock SMTP server
// The tests above cover the validation logic and error handling
