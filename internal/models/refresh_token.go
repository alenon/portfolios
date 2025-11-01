package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RefreshToken represents a refresh token in the system
type RefreshToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id" validate:"required"`
	TokenHash string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"-" validate:"required"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at" validate:"required"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	User      User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName specifies the table name for the RefreshToken model
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// BeforeCreate hook to generate UUID before creating a new refresh token
func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	if rt.CreatedAt.IsZero() {
		rt.CreatedAt = time.Now().UTC()
	}
	return nil
}

// IsExpired checks if the refresh token has expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().UTC().After(rt.ExpiresAt)
}

// IsRevoked checks if the refresh token has been revoked
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}

// IsValid checks if the refresh token is valid (not expired and not revoked)
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked()
}

// Revoke marks the refresh token as revoked
func (rt *RefreshToken) Revoke(tx *gorm.DB) error {
	now := time.Now().UTC()
	rt.RevokedAt = &now
	return tx.Model(rt).Update("revoked_at", now).Error
}
