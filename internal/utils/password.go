package utils

import (
	"fmt"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// MinPasswordLength is the minimum required password length
	MinPasswordLength = 8
	// BcryptCost is the cost factor for bcrypt hashing
	BcryptCost = 12
)

// HashPassword hashes a password using bcrypt with cost factor 12
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedPassword), nil
}

// CheckPassword verifies if the provided password matches the stored hash
func CheckPassword(password, hash string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	if hash == "" {
		return fmt.Errorf("hash cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("password does not match: %w", err)
	}

	return nil
}

// ValidatePassword checks if a password meets the required complexity rules
// Requirements: minimum 8 characters, at least one uppercase, one lowercase, one number
func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", MinPasswordLength)
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	return nil
}

// ValidateAndHashPassword validates a password and returns its hash if valid
func ValidateAndHashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	return HashPassword(password)
}
