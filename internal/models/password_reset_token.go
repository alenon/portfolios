package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PasswordResetToken represents a password reset token in the system
type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id" validate:"required"`
	TokenHash string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"-" validate:"required"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at" validate:"required"`
	CreatedAt time.Time  `json:"created_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	User      User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName specifies the table name for the PasswordResetToken model
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// BeforeCreate hook to generate UUID before creating a new password reset token
func (prt *PasswordResetToken) BeforeCreate(tx *gorm.DB) error {
	if prt.ID == uuid.Nil {
		prt.ID = uuid.New()
	}
	if prt.CreatedAt.IsZero() {
		prt.CreatedAt = time.Now().UTC()
	}
	return nil
}

// IsExpired checks if the password reset token has expired
func (prt *PasswordResetToken) IsExpired() bool {
	return time.Now().UTC().After(prt.ExpiresAt)
}

// IsUsed checks if the password reset token has been used
func (prt *PasswordResetToken) IsUsed() bool {
	return prt.UsedAt != nil
}

// IsValid checks if the password reset token is valid (not expired and not used)
func (prt *PasswordResetToken) IsValid() bool {
	return !prt.IsExpired() && !prt.IsUsed()
}

// MarkAsUsed marks the password reset token as used
func (prt *PasswordResetToken) MarkAsUsed(tx *gorm.DB) error {
	now := time.Now().UTC()
	prt.UsedAt = &now
	return tx.Model(prt).Update("used_at", now).Error
}
