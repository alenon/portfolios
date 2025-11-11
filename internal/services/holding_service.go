package services

import (
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
	"github.com/shopspring/decimal"
)

// HoldingService defines the interface for holding operations
type HoldingService interface {
	GetByPortfolioID(portfolioID, userID string) ([]*models.Holding, error)
	GetByPortfolioIDAndSymbol(portfolioID, symbol, userID string) (*models.Holding, error)
	GetPortfolioValue(portfolioID, userID string, prices map[string]decimal.Decimal) (decimal.Decimal, error)
}

// holdingService implements HoldingService interface
type holdingService struct {
	holdingRepo   repository.HoldingRepository
	portfolioRepo repository.PortfolioRepository
}

// NewHoldingService creates a new HoldingService instance
func NewHoldingService(
	holdingRepo repository.HoldingRepository,
	portfolioRepo repository.PortfolioRepository,
) HoldingService {
	return &holdingService{
		holdingRepo:   holdingRepo,
		portfolioRepo: portfolioRepo,
	}
}

// GetByPortfolioID retrieves all holdings for a portfolio
func (s *holdingService) GetByPortfolioID(portfolioID, userID string) ([]*models.Holding, error) {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return nil, models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	return s.holdingRepo.FindByPortfolioID(portfolioID)
}

// GetByPortfolioIDAndSymbol retrieves a specific holding by portfolio ID and symbol
func (s *holdingService) GetByPortfolioIDAndSymbol(portfolioID, symbol, userID string) (*models.Holding, error) {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return nil, models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	return s.holdingRepo.FindByPortfolioIDAndSymbol(portfolioID, symbol)
}

// GetPortfolioValue calculates the total value of all holdings in a portfolio
// prices is a map of symbol to current market price
func (s *holdingService) GetPortfolioValue(portfolioID, userID string, prices map[string]decimal.Decimal) (decimal.Decimal, error) {
	holdings, err := s.GetByPortfolioID(portfolioID, userID)
	if err != nil {
		return decimal.Zero, err
	}

	totalValue := decimal.Zero
	for _, holding := range holdings {
		price, exists := prices[holding.Symbol]
		if !exists {
			// If price not provided, use cost basis (conservative estimate)
			totalValue = totalValue.Add(holding.CostBasis)
		} else {
			marketValue := holding.Quantity.Mul(price)
			totalValue = totalValue.Add(marketValue)
		}
	}

	return totalValue, nil
}
