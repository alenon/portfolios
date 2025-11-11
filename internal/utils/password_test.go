package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	t.Run("successful hashing", func(t *testing.T) {
		password := "TestPassword123"
		hash, err := HashPassword(password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.True(t, strings.HasPrefix(hash, "$2a$"))

		// Verify it's a valid bcrypt hash
		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		assert.NoError(t, err)
	})

	t.Run("empty password error", func(t *testing.T) {
		_, err := HashPassword("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password cannot be empty")
	})

	t.Run("different passwords produce different hashes", func(t *testing.T) {
		hash1, err1 := HashPassword("Password123")
		hash2, err2 := HashPassword("Password456")
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("same password produces different hashes (salt)", func(t *testing.T) {
		password := "SamePassword123"
		hash1, err1 := HashPassword(password)
		hash2, err2 := HashPassword(password)
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2) // Due to salt
	})
}

func TestCheckPassword(t *testing.T) {
	password := "TestPassword123"
	hash, _ := HashPassword(password)

	t.Run("correct password matches", func(t *testing.T) {
		err := CheckPassword(password, hash)
		assert.NoError(t, err)
	})

	t.Run("incorrect password fails", func(t *testing.T) {
		err := CheckPassword("WrongPassword123", hash)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password does not match")
	})

	t.Run("empty password error", func(t *testing.T) {
		err := CheckPassword("", hash)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password cannot be empty")
	})

	t.Run("empty hash error", func(t *testing.T) {
		err := CheckPassword(password, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "hash cannot be empty")
	})

	t.Run("invalid hash format", func(t *testing.T) {
		err := CheckPassword(password, "invalid-hash")
		assert.Error(t, err)
	})
}

func TestValidatePassword(t *testing.T) {
	t.Run("valid password", func(t *testing.T) {
		err := ValidatePassword("ValidPass123")
		assert.NoError(t, err)
	})

	t.Run("too short password", func(t *testing.T) {
		err := ValidatePassword("Short1A")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 8 characters")
	})

	t.Run("minimum length password", func(t *testing.T) {
		err := ValidatePassword("Valid1Aa")
		assert.NoError(t, err)
	})

	t.Run("missing uppercase letter", func(t *testing.T) {
		err := ValidatePassword("lowercase123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "uppercase letter")
	})

	t.Run("missing lowercase letter", func(t *testing.T) {
		err := ValidatePassword("UPPERCASE123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lowercase letter")
	})

	t.Run("missing number", func(t *testing.T) {
		err := ValidatePassword("NoNumbers")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "number")
	})

	t.Run("long complex password", func(t *testing.T) {
		err := ValidatePassword("VeryLongAndComplexPassword123WithManyCharacters")
		assert.NoError(t, err)
	})

	t.Run("password with special characters", func(t *testing.T) {
		err := ValidatePassword("Pass123!@#$%")
		assert.NoError(t, err)
	})

	t.Run("password with spaces", func(t *testing.T) {
		err := ValidatePassword("Pass Word 123")
		assert.NoError(t, err)
	})

	t.Run("empty password", func(t *testing.T) {
		err := ValidatePassword("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 8 characters")
	})
}

func TestValidateAndHashPassword(t *testing.T) {
	t.Run("valid password is hashed", func(t *testing.T) {
		password := "ValidPass123"
		hash, err := ValidateAndHashPassword(password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)

		// Verify the hash works
		err = CheckPassword(password, hash)
		assert.NoError(t, err)
	})

	t.Run("invalid password returns validation error", func(t *testing.T) {
		_, err := ValidateAndHashPassword("short")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 8 characters")
	})

	t.Run("password without uppercase fails", func(t *testing.T) {
		_, err := ValidateAndHashPassword("lowercase123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "uppercase letter")
	})

	t.Run("password without lowercase fails", func(t *testing.T) {
		_, err := ValidateAndHashPassword("UPPERCASE123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lowercase letter")
	})

	t.Run("password without number fails", func(t *testing.T) {
		_, err := ValidateAndHashPassword("NoNumbers")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "number")
	})

	t.Run("empty password fails", func(t *testing.T) {
		_, err := ValidateAndHashPassword("")
		assert.Error(t, err)
	})
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 8, MinPasswordLength)
	assert.Equal(t, 12, BcryptCost)
}
