package repository

import (
	"fmt"
	"time"

	"github.com/lenon/portfolios/internal/models"
	"gorm.io/gorm"
)

// PerformanceSnapshotRepository defines the interface for performance snapshot data operations
type PerformanceSnapshotRepository interface {
	Create(snapshot *models.PerformanceSnapshot) error
	FindByID(id string) (*models.PerformanceSnapshot, error)
	FindByPortfolioID(portfolioID string, limit, offset int) ([]*models.PerformanceSnapshot, error)
	FindByPortfolioIDAndDateRange(portfolioID string, startDate, endDate time.Time) ([]*models.PerformanceSnapshot, error)
	FindLatestByPortfolioID(portfolioID string) (*models.PerformanceSnapshot, error)
	FindByPortfolioIDAndDate(portfolioID string, date time.Time) (*models.PerformanceSnapshot, error)
	Delete(id string) error
	DeleteByPortfolioID(portfolioID string) error
}

// performanceSnapshotRepository implements PerformanceSnapshotRepository interface
type performanceSnapshotRepository struct {
	db *gorm.DB
}

// NewPerformanceSnapshotRepository creates a new PerformanceSnapshotRepository instance
func NewPerformanceSnapshotRepository(db *gorm.DB) PerformanceSnapshotRepository {
	return &performanceSnapshotRepository{db: db}
}

// Create creates a new performance snapshot
func (r *performanceSnapshotRepository) Create(snapshot *models.PerformanceSnapshot) error {
	if snapshot == nil {
		return fmt.Errorf("snapshot cannot be nil")
	}

	if err := snapshot.Validate(); err != nil {
		return err
	}

	return r.db.Create(snapshot).Error
}

// FindByID finds a performance snapshot by ID
func (r *performanceSnapshotRepository) FindByID(id string) (*models.PerformanceSnapshot, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	var snapshot models.PerformanceSnapshot
	if err := r.db.Where("id = ?", id).First(&snapshot).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrPerformanceSnapshotNotFound
		}
		return nil, err
	}

	return &snapshot, nil
}

// FindByPortfolioID finds all performance snapshots for a portfolio, ordered by date descending
func (r *performanceSnapshotRepository) FindByPortfolioID(portfolioID string, limit, offset int) ([]*models.PerformanceSnapshot, error) {
	if portfolioID == "" {
		return nil, fmt.Errorf("portfolio ID cannot be empty")
	}

	var snapshots []*models.PerformanceSnapshot
	query := r.db.Where("portfolio_id = ?", portfolioID).
		Order("date DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&snapshots).Error; err != nil {
		return nil, err
	}

	return snapshots, nil
}

// FindByPortfolioIDAndDateRange finds performance snapshots for a portfolio within a date range
func (r *performanceSnapshotRepository) FindByPortfolioIDAndDateRange(
	portfolioID string,
	startDate, endDate time.Time,
) ([]*models.PerformanceSnapshot, error) {
	if portfolioID == "" {
		return nil, fmt.Errorf("portfolio ID cannot be empty")
	}

	var snapshots []*models.PerformanceSnapshot
	if err := r.db.Where("portfolio_id = ? AND date >= ? AND date <= ?", portfolioID, startDate, endDate).
		Order("date ASC").
		Find(&snapshots).Error; err != nil {
		return nil, err
	}

	return snapshots, nil
}

// FindLatestByPortfolioID finds the most recent performance snapshot for a portfolio
func (r *performanceSnapshotRepository) FindLatestByPortfolioID(portfolioID string) (*models.PerformanceSnapshot, error) {
	if portfolioID == "" {
		return nil, fmt.Errorf("portfolio ID cannot be empty")
	}

	var snapshot models.PerformanceSnapshot
	if err := r.db.Where("portfolio_id = ?", portfolioID).
		Order("date DESC").
		First(&snapshot).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrPerformanceSnapshotNotFound
		}
		return nil, err
	}

	return &snapshot, nil
}

// FindByPortfolioIDAndDate finds a performance snapshot for a specific portfolio and date
func (r *performanceSnapshotRepository) FindByPortfolioIDAndDate(portfolioID string, date time.Time) (*models.PerformanceSnapshot, error) {
	if portfolioID == "" {
		return nil, fmt.Errorf("portfolio ID cannot be empty")
	}

	var snapshot models.PerformanceSnapshot
	// Normalize date to start of day for comparison
	normalizedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	if err := r.db.Where("portfolio_id = ? AND DATE(date) = DATE(?)", portfolioID, normalizedDate).
		First(&snapshot).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrPerformanceSnapshotNotFound
		}
		return nil, err
	}

	return &snapshot, nil
}

// Delete deletes a performance snapshot by ID
func (r *performanceSnapshotRepository) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	result := r.db.Where("id = ?", id).Delete(&models.PerformanceSnapshot{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return models.ErrPerformanceSnapshotNotFound
	}

	return nil
}

// DeleteByPortfolioID deletes all performance snapshots for a portfolio
func (r *performanceSnapshotRepository) DeleteByPortfolioID(portfolioID string) error {
	if portfolioID == "" {
		return fmt.Errorf("portfolio ID cannot be empty")
	}

	return r.db.Where("portfolio_id = ?", portfolioID).Delete(&models.PerformanceSnapshot{}).Error
}
