package services

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
)

// PortfolioService defines the interface for portfolio operations
type PortfolioService interface {
	Create(userID, name, description, baseCurrency string, costBasisMethod models.CostBasisMethod) (*models.Portfolio, error)
	GetByID(id string, userID string) (*models.Portfolio, error)
	GetAllByUserID(userID string) ([]*models.Portfolio, error)
	Update(id, userID, name, description string) (*models.Portfolio, error)
	Delete(id, userID string) error
}

// portfolioService implements PortfolioService interface
type portfolioService struct {
	portfolioRepo repository.PortfolioRepository
	userRepo      repository.UserRepository
}

// NewPortfolioService creates a new PortfolioService instance
func NewPortfolioService(
	portfolioRepo repository.PortfolioRepository,
	userRepo repository.UserRepository,
) PortfolioService {
	return &portfolioService{
		portfolioRepo: portfolioRepo,
		userRepo:      userRepo,
	}
}

// Create creates a new portfolio for a user
func (s *portfolioService) Create(
	userID, name, description, baseCurrency string,
	costBasisMethod models.CostBasisMethod,
) (*models.Portfolio, error) {
	// Validate user exists
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Validate name is not empty
	if name == "" {
		return nil, models.ErrPortfolioNameRequired
	}

	// Check if portfolio with same name already exists for user
	exists, err := s.portfolioRepo.ExistsByUserIDAndName(userID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check portfolio existence: %w", err)
	}
	if exists {
		return nil, models.ErrPortfolioDuplicateName
	}

	// Set defaults
	if baseCurrency == "" {
		baseCurrency = "USD"
	}
	if costBasisMethod == "" {
		costBasisMethod = models.CostBasisFIFO
	}

	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Create portfolio
	portfolio := &models.Portfolio{
		UserID:          uid,
		Name:            name,
		Description:     description,
		BaseCurrency:    baseCurrency,
		CostBasisMethod: costBasisMethod,
	}

	// Validate portfolio
	if err := portfolio.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.portfolioRepo.Create(portfolio); err != nil {
		return nil, fmt.Errorf("failed to create portfolio: %w", err)
	}

	return portfolio, nil
}

// GetByID retrieves a portfolio by ID, ensuring it belongs to the user
func (s *portfolioService) GetByID(id string, userID string) (*models.Portfolio, error) {
	portfolio, err := s.portfolioRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Verify the portfolio belongs to the user
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	return portfolio, nil
}

// GetAllByUserID retrieves all portfolios for a user
func (s *portfolioService) GetAllByUserID(userID string) ([]*models.Portfolio, error) {
	// Validate user exists
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	portfolios, err := s.portfolioRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve portfolios: %w", err)
	}

	return portfolios, nil
}

// Update updates a portfolio's details
func (s *portfolioService) Update(id, userID, name, description string) (*models.Portfolio, error) {
	// Get existing portfolio and verify ownership
	portfolio, err := s.GetByID(id, userID)
	if err != nil {
		return nil, err
	}

	// Check if new name conflicts with another portfolio
	if name != "" && name != portfolio.Name {
		exists, err := s.portfolioRepo.ExistsByUserIDAndName(userID, name)
		if err != nil {
			return nil, fmt.Errorf("failed to check portfolio name: %w", err)
		}
		if exists {
			return nil, models.ErrPortfolioDuplicateName
		}
		portfolio.Name = name
	}

	// Update description if provided
	if description != "" {
		portfolio.Description = description
	}

	// Validate updated portfolio
	if err := portfolio.Validate(); err != nil {
		return nil, err
	}

	// Save changes
	if err := s.portfolioRepo.Update(portfolio); err != nil {
		return nil, fmt.Errorf("failed to update portfolio: %w", err)
	}

	return portfolio, nil
}

// Delete deletes a portfolio, ensuring it belongs to the user
func (s *portfolioService) Delete(id, userID string) error {
	// Get portfolio and verify ownership
	_, err := s.GetByID(id, userID)
	if err != nil {
		return err
	}

	// Delete the portfolio (cascade will delete related records)
	if err := s.portfolioRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete portfolio: %w", err)
	}

	return nil
}
