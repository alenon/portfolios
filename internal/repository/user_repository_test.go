package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Auto migrate the models
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

func TestNewUserRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	assert.NotNil(t, repo)
}

func TestUserRepository_Create_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	err := repo.Create(user)

	assert.NoError(t, err)
}

func TestUserRepository_Create_NilUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	err := repo.Create(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user cannot be nil")
}

func TestUserRepository_FindByEmail_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	err := repo.Create(user)
	assert.NoError(t, err)

	foundUser, err := repo.FindByEmail("test@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, user.Email, foundUser.Email)
	assert.Equal(t, user.ID, foundUser.ID)
}

func TestUserRepository_FindByEmail_EmptyEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user, err := repo.FindByEmail("")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "email cannot be empty")
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user, err := repo.FindByEmail("nonexistent@example.com")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not found with email")
}

func TestUserRepository_FindByID_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	err := repo.Create(user)
	assert.NoError(t, err)

	foundUser, err := repo.FindByID(user.ID.String())

	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, user.Email, foundUser.Email)
	assert.Equal(t, user.ID, foundUser.ID)
}

func TestUserRepository_FindByID_EmptyID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user, err := repo.FindByID("")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "id cannot be empty")
}

func TestUserRepository_FindByID_InvalidUUID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user, err := repo.FindByID("invalid-uuid")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid user ID format")
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user, err := repo.FindByID(uuid.New().String())

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not found with id")
}

func TestUserRepository_UpdateLastLogin_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	err := repo.Create(user)
	assert.NoError(t, err)

	err = repo.UpdateLastLogin(user.ID.String())

	assert.NoError(t, err)

	// Verify last login was updated
	updatedUser, _ := repo.FindByID(user.ID.String())
	assert.NotNil(t, updatedUser.LastLoginAt)
}

func TestUserRepository_UpdateLastLogin_EmptyID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	err := repo.UpdateLastLogin("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")
}

func TestUserRepository_UpdateLastLogin_InvalidUUID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	err := repo.UpdateLastLogin("invalid-uuid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID format")
}

func TestUserRepository_UpdateLastLogin_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	err := repo.UpdateLastLogin(uuid.New().String())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found with id")
}

func TestUserRepository_UpdatePassword_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "old-hashed-password",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	repo.Create(user)

	newPasswordHash := "new-hashed-password"
	err := repo.UpdatePassword(user.ID.String(), newPasswordHash)

	assert.NoError(t, err)

	// Verify password was updated
	updatedUser, _ := repo.FindByID(user.ID.String())
	assert.Equal(t, newPasswordHash, updatedUser.PasswordHash)
}

func TestUserRepository_UpdatePassword_EmptyID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	err := repo.UpdatePassword("", "new-password-hash")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id cannot be empty")
}

func TestUserRepository_UpdatePassword_EmptyPasswordHash(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	err := repo.UpdatePassword(uuid.New().String(), "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password hash cannot be empty")
}

func TestUserRepository_UpdatePassword_InvalidUUID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	err := repo.UpdatePassword("invalid-uuid", "new-password-hash")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID format")
}

func TestUserRepository_UpdatePassword_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	err := repo.UpdatePassword(uuid.New().String(), "new-password-hash")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found with id")
}
