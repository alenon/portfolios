package services

import (
	"fmt"
	"time"

	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
	"github.com/shopspring/decimal"
)

// PerformanceSnapshotService defines the interface for performance snapshot operations
type PerformanceSnapshotService interface {
	CreateSnapshot(portfolioID, userID string, prices map[string]decimal.Decimal) (*models.PerformanceSnapshot, error)
	GetByPortfolioID(portfolioID, userID string, limit, offset int) ([]*models.PerformanceSnapshot, error)
	GetByDateRange(portfolioID, userID string, startDate, endDate time.Time) ([]*models.PerformanceSnapshot, error)
	GetLatest(portfolioID, userID string) (*models.PerformanceSnapshot, error)
}

// performanceSnapshotService implements PerformanceSnapshotService interface
type performanceSnapshotService struct {
	snapshotRepo  repository.PerformanceSnapshotRepository
	portfolioRepo repository.PortfolioRepository
	holdingRepo   repository.HoldingRepository
}

// NewPerformanceSnapshotService creates a new PerformanceSnapshotService instance
func NewPerformanceSnapshotService(
	snapshotRepo repository.PerformanceSnapshotRepository,
	portfolioRepo repository.PortfolioRepository,
	holdingRepo repository.HoldingRepository,
) PerformanceSnapshotService {
	return &performanceSnapshotService{
		snapshotRepo:  snapshotRepo,
		portfolioRepo: portfolioRepo,
		holdingRepo:   holdingRepo,
	}
}

// CreateSnapshot creates a performance snapshot for a portfolio at the current time
func (s *performanceSnapshotService) CreateSnapshot(
	portfolioID, userID string,
	prices map[string]decimal.Decimal,
) (*models.PerformanceSnapshot, error) {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return nil, models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	// Get all holdings for the portfolio
	holdings, err := s.holdingRepo.FindByPortfolioID(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve holdings: %w", err)
	}

	// Calculate total value and cost basis
	totalValue := decimal.Zero
	totalCostBasis := decimal.Zero

	for _, holding := range holdings {
		// Add cost basis
		totalCostBasis = totalCostBasis.Add(holding.CostBasis)

		// Calculate market value
		price, exists := prices[holding.Symbol]
		if !exists {
			// If no price provided, use cost basis (conservative estimate)
			totalValue = totalValue.Add(holding.CostBasis)
		} else {
			marketValue := holding.Quantity.Mul(price)
			totalValue = totalValue.Add(marketValue)
		}
	}

	// Create snapshot
	snapshot := &models.PerformanceSnapshot{
		PortfolioID:    portfolio.ID,
		Date:           time.Now().UTC(),
		TotalValue:     totalValue,
		TotalCostBasis: totalCostBasis,
	}

	// Calculate return metrics
	snapshot.CalculateMetrics()

	// Try to get previous day's snapshot for day change calculation
	previousSnapshot, err := s.snapshotRepo.FindLatestByPortfolioID(portfolioID)
	if err == nil && previousSnapshot != nil {
		snapshot.CalculateDayChange(previousSnapshot.TotalValue)
	}

	// Save snapshot
	if err := s.snapshotRepo.Create(snapshot); err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return snapshot, nil
}

// GetByPortfolioID retrieves performance snapshots for a portfolio
func (s *performanceSnapshotService) GetByPortfolioID(
	portfolioID, userID string,
	limit, offset int,
) ([]*models.PerformanceSnapshot, error) {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return nil, models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	return s.snapshotRepo.FindByPortfolioID(portfolioID, limit, offset)
}

// GetByDateRange retrieves performance snapshots for a portfolio within a date range
func (s *performanceSnapshotService) GetByDateRange(
	portfolioID, userID string,
	startDate, endDate time.Time,
) ([]*models.PerformanceSnapshot, error) {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return nil, models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	return s.snapshotRepo.FindByPortfolioIDAndDateRange(portfolioID, startDate, endDate)
}

// GetLatest retrieves the most recent performance snapshot for a portfolio
func (s *performanceSnapshotService) GetLatest(portfolioID, userID string) (*models.PerformanceSnapshot, error) {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return nil, models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	return s.snapshotRepo.FindLatestByPortfolioID(portfolioID)
}
