package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CostBasisMethod represents the cost basis calculation method
type CostBasisMethod string

const (
	CostBasisFIFO        CostBasisMethod = "FIFO"
	CostBasisLIFO        CostBasisMethod = "LIFO"
	CostBasisSpecificLot CostBasisMethod = "SPECIFIC_LOT"
)

// Portfolio represents a user's investment portfolio
type Portfolio struct {
	ID              uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          uuid.UUID        `gorm:"type:uuid;not null;index" json:"user_id" validate:"required"`
	Name            string           `gorm:"type:varchar(255);not null" json:"name" validate:"required,min=1,max=255"`
	Description     string           `gorm:"type:text" json:"description,omitempty"`
	BaseCurrency    string           `gorm:"type:varchar(3);not null;default:'USD'" json:"base_currency" validate:"required,len=3"`
	CostBasisMethod CostBasisMethod  `gorm:"type:varchar(20);not null;default:'FIFO'" json:"cost_basis_method" validate:"required,oneof=FIFO LIFO SPECIFIC_LOT"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	User            *User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for the Portfolio model
func (Portfolio) TableName() string {
	return "portfolios"
}

// BeforeCreate hook to generate UUID before creating a new portfolio
func (p *Portfolio) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now().UTC()
	}
	if p.BaseCurrency == "" {
		p.BaseCurrency = "USD"
	}
	if p.CostBasisMethod == "" {
		p.CostBasisMethod = CostBasisFIFO
	}
	return nil
}

// BeforeUpdate hook to update the UpdatedAt timestamp
func (p *Portfolio) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// Validate checks if the portfolio has valid data
func (p *Portfolio) Validate() error {
	if p.Name == "" {
		return ErrPortfolioNameRequired
	}
	if len(p.BaseCurrency) != 3 {
		return ErrInvalidCurrency
	}
	if !p.isValidCostBasisMethod() {
		return ErrInvalidCostBasisMethod
	}
	return nil
}

// isValidCostBasisMethod checks if the cost basis method is valid
func (p *Portfolio) isValidCostBasisMethod() bool {
	switch p.CostBasisMethod {
	case CostBasisFIFO, CostBasisLIFO, CostBasisSpecificLot:
		return true
	default:
		return false
	}
}
