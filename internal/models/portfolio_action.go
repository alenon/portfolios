package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PortfolioActionStatus represents the status of a pending corporate action for a portfolio
type PortfolioActionStatus string

const (
	PortfolioActionStatusPending  PortfolioActionStatus = "PENDING"
	PortfolioActionStatusApproved PortfolioActionStatus = "APPROVED"
	PortfolioActionStatusRejected PortfolioActionStatus = "REJECTED"
	PortfolioActionStatusApplied  PortfolioActionStatus = "APPLIED"
)

// PortfolioAction represents a corporate action that affects a specific portfolio
// and requires user approval before being applied
type PortfolioAction struct {
	ID                uuid.UUID             `gorm:"type:uuid;primaryKey" json:"id"`
	PortfolioID       uuid.UUID             `gorm:"type:uuid;not null;index" json:"portfolio_id" validate:"required"`
	CorporateActionID uuid.UUID             `gorm:"type:uuid;not null;index" json:"corporate_action_id" validate:"required"`
	Status            PortfolioActionStatus `gorm:"type:varchar(20);not null;default:'PENDING';index" json:"status"`
	AffectedSymbol    string                `gorm:"type:varchar(20);not null" json:"affected_symbol" validate:"required"`
	SharesAffected    int64                 `gorm:"not null" json:"shares_affected"`
	DetectedAt        time.Time             `gorm:"not null" json:"detected_at"`
	ReviewedAt        *time.Time            `json:"reviewed_at,omitempty"`
	AppliedAt         *time.Time            `json:"applied_at,omitempty"`
	ReviewedByUserID  *uuid.UUID            `gorm:"type:uuid" json:"reviewed_by_user_id,omitempty"`
	Notes             string                `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
	Portfolio         *Portfolio            `gorm:"foreignKey:PortfolioID" json:"portfolio,omitempty"`
	CorporateAction   *CorporateAction      `gorm:"foreignKey:CorporateActionID" json:"corporate_action,omitempty"`
}

// TableName specifies the table name for the PortfolioAction model
func (PortfolioAction) TableName() string {
	return "portfolio_actions"
}

// BeforeCreate hook to generate UUID before creating a new portfolio action
func (pa *PortfolioAction) BeforeCreate(tx *gorm.DB) error {
	if pa.ID == uuid.Nil {
		pa.ID = uuid.New()
	}
	if pa.CreatedAt.IsZero() {
		pa.CreatedAt = time.Now().UTC()
	}
	if pa.UpdatedAt.IsZero() {
		pa.UpdatedAt = time.Now().UTC()
	}
	if pa.Status == "" {
		pa.Status = PortfolioActionStatusPending
	}
	if pa.DetectedAt.IsZero() {
		pa.DetectedAt = time.Now().UTC()
	}
	return nil
}

// BeforeUpdate hook to update the UpdatedAt timestamp
func (pa *PortfolioAction) BeforeUpdate(tx *gorm.DB) error {
	pa.UpdatedAt = time.Now().UTC()
	return nil
}

// IsPending checks if the action is still pending user review
func (pa *PortfolioAction) IsPending() bool {
	return pa.Status == PortfolioActionStatusPending
}

// IsApplied checks if the action has been applied
func (pa *PortfolioAction) IsApplied() bool {
	return pa.Status == PortfolioActionStatusApplied
}

// Approve marks the action as approved
func (pa *PortfolioAction) Approve(userID uuid.UUID) {
	now := time.Now().UTC()
	pa.Status = PortfolioActionStatusApproved
	pa.ReviewedAt = &now
	pa.ReviewedByUserID = &userID
}

// Reject marks the action as rejected
func (pa *PortfolioAction) Reject(userID uuid.UUID, reason string) {
	now := time.Now().UTC()
	pa.Status = PortfolioActionStatusRejected
	pa.ReviewedAt = &now
	pa.ReviewedByUserID = &userID
	if reason != "" {
		pa.Notes = reason
	}
}

// MarkApplied marks the action as applied
func (pa *PortfolioAction) MarkApplied() {
	now := time.Now().UTC()
	pa.Status = PortfolioActionStatusApplied
	pa.AppliedAt = &now
}
