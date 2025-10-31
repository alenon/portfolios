package services

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)

// EmailService defines the interface for email operations
type EmailService interface {
	SendPasswordResetEmail(to, resetToken string) error
}

// emailService implements EmailService interface
type emailService struct {
	host     string
	port     int
	username string
	password string
	from     string
}

// NewEmailService creates a new EmailService instance
func NewEmailService(host string, port int, username, password, from string) EmailService {
	return &emailService{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

// SendPasswordResetEmail sends a password reset email with the reset token
func (s *emailService) SendPasswordResetEmail(to, resetToken string) error {
	if to == "" {
		return fmt.Errorf("recipient email cannot be empty")
	}
	if resetToken == "" {
		return fmt.Errorf("reset token cannot be empty")
	}

	// Create reset link
	resetLink := fmt.Sprintf("https://app.example.com/reset-password?token=%s", resetToken)

	// Create email subject and body
	subject := "Password Reset Request"
	body := fmt.Sprintf(`Hello,

You have requested to reset your password. Please click the link below to reset your password:

%s

This link will expire in 1 hour.

If you did not request a password reset, please ignore this email.

Best regards,
The Portfolios Team`, resetLink)

	// Compose email message
	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", s.from, to, subject, body)

	// Set up authentication
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	// Connect to SMTP server with TLS
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	// Try to send with TLS
	err := s.sendWithTLS(addr, auth, []byte(message), to)
	if err != nil {
		// If TLS fails, try without TLS for development/testing
		err = smtp.SendMail(addr, auth, s.from, []string{to}, []byte(message))
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
	}

	return nil
}

// sendWithTLS sends email with TLS/SSL connection
func (s *emailService) sendWithTLS(addr string, auth smtp.Auth, message []byte, to string) error {
	// Create TLS configuration
	tlsConfig := &tls.Config{
		ServerName: s.host,
	}

	// Connect to the SMTP server
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Start TLS
	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Set sender
	if err = client.Mail(s.from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send message
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(message)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	// Send QUIT command
	return client.Quit()
}
