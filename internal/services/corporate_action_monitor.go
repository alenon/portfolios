package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
)

// CorporateActionMonitor monitors and detects corporate actions for portfolio holdings
type CorporateActionMonitor struct {
	corporateActionRepo repository.CorporateActionRepository
	portfolioRepo       repository.PortfolioRepository
	holdingRepo         repository.HoldingRepository
	portfolioActionRepo repository.PortfolioActionRepository
}

// NewCorporateActionMonitor creates a new corporate action monitor
func NewCorporateActionMonitor(
	corporateActionRepo repository.CorporateActionRepository,
	portfolioRepo repository.PortfolioRepository,
	holdingRepo repository.HoldingRepository,
	portfolioActionRepo repository.PortfolioActionRepository,
) *CorporateActionMonitor {
	return &CorporateActionMonitor{
		corporateActionRepo: corporateActionRepo,
		portfolioRepo:       portfolioRepo,
		holdingRepo:         holdingRepo,
		portfolioActionRepo: portfolioActionRepo,
	}
}

// DetectAndSuggestActions detects new corporate actions and creates pending actions for affected portfolios
func (m *CorporateActionMonitor) DetectAndSuggestActions(ctx context.Context) error {
	log.Println("Starting corporate action detection...")

	// Step 1: Fetch new/unapplied corporate actions
	actions, err := m.corporateActionRepo.FindUnapplied()
	if err != nil {
		return fmt.Errorf("failed to fetch unapplied corporate actions: %w", err)
	}

	if len(actions) == 0 {
		log.Println("No new corporate actions to process")
		return nil
	}

	log.Printf("Found %d unapplied corporate actions", len(actions))

	// Step 2: For each corporate action, find affected portfolios
	for _, action := range actions {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled: %w", err)
		}

		if err := m.processAction(action); err != nil {
			log.Printf("Error processing corporate action %s for symbol %s: %v",
				action.ID, action.Symbol, err)
			continue
		}
	}

	log.Println("Corporate action detection completed")
	return nil
}

// processAction processes a single corporate action and creates portfolio actions
func (m *CorporateActionMonitor) processAction(action *models.CorporateAction) error {
	log.Printf("Processing %s for symbol %s on %s",
		action.Type, action.Symbol, action.Date.Format("2006-01-02"))

	// Find all portfolios with holdings in this symbol
	portfolios, err := m.findPortfoliosWithSymbol(action.Symbol)
	if err != nil {
		return fmt.Errorf("failed to find portfolios with symbol %s: %w", action.Symbol, err)
	}

	if len(portfolios) == 0 {
		log.Printf("No portfolios hold symbol %s, skipping", action.Symbol)
		return nil
	}

	log.Printf("Found %d portfolios holding %s", len(portfolios), action.Symbol)

	// Create portfolio actions for each affected portfolio
	createdCount := 0
	for _, portfolio := range portfolios {
		// Check if action already exists
		exists, err := m.portfolioActionRepo.ExistsPendingForPortfolioAndAction(
			portfolio.ID.String(),
			action.ID.String(),
		)
		if err != nil {
			log.Printf("Error checking existing action for portfolio %s: %v", portfolio.ID, err)
			continue
		}

		if exists {
			log.Printf("Pending action already exists for portfolio %s, skipping", portfolio.ID)
			continue
		}

		// Get holding to determine shares affected
		holding, err := m.holdingRepo.FindByPortfolioIDAndSymbol(
			portfolio.ID.String(),
			action.Symbol,
		)
		if err != nil {
			log.Printf("Error getting holding for portfolio %s: %v", portfolio.ID, err)
			continue
		}

		// Create portfolio action
		portfolioAction := &models.PortfolioAction{
			PortfolioID:       portfolio.ID,
			CorporateActionID: action.ID,
			Status:            models.PortfolioActionStatusPending,
			AffectedSymbol:    action.Symbol,
			SharesAffected:    holding.Quantity.IntPart(),
			DetectedAt:        time.Now().UTC(),
			Notes:             m.generateActionDescription(action, holding),
		}

		if err := m.portfolioActionRepo.Create(portfolioAction); err != nil {
			log.Printf("Failed to create portfolio action for portfolio %s: %v", portfolio.ID, err)
			continue
		}

		createdCount++
		log.Printf("Created pending action for portfolio %s (%d shares affected)",
			portfolio.ID, portfolioAction.SharesAffected)
	}

	log.Printf("Created %d pending portfolio actions for symbol %s", createdCount, action.Symbol)
	return nil
}

// findPortfoliosWithSymbol finds all portfolios that have holdings in the given symbol
func (m *CorporateActionMonitor) findPortfoliosWithSymbol(symbol string) ([]*models.Portfolio, error) {
	// This is a simplified implementation
	// In a production system, you'd want a more efficient query

	// Get all holdings for this symbol
	// Note: We need a new repository method for this
	// For now, we'll need to iterate through portfolios

	// This is a placeholder - in production, add a method to HoldingRepository:
	// FindPortfoliosBySymbol(symbol string) ([]*models.Portfolio, error)

	return m.findPortfoliosBySymbolWorkaround(symbol)
}

// findPortfoliosBySymbolWorkaround uses holdings to find portfolios with a specific symbol
func (m *CorporateActionMonitor) findPortfoliosBySymbolWorkaround(symbol string) ([]*models.Portfolio, error) {
	// Find all holdings for this symbol
	holdings, err := m.holdingRepo.FindBySymbol(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to find holdings: %w", err)
	}

	// Extract unique portfolios
	portfolioMap := make(map[string]*models.Portfolio)
	for _, holding := range holdings {
		if holding.Portfolio != nil {
			portfolioMap[holding.Portfolio.ID.String()] = holding.Portfolio
		}
	}

	// Convert map to slice
	portfolios := make([]*models.Portfolio, 0, len(portfolioMap))
	for _, portfolio := range portfolioMap {
		portfolios = append(portfolios, portfolio)
	}

	return portfolios, nil
}

// generateActionDescription generates a human-readable description of the action
func (m *CorporateActionMonitor) generateActionDescription(
	action *models.CorporateAction,
	holding *models.Holding,
) string {
	switch action.Type {
	case models.CorporateActionTypeSplit:
		if action.Ratio != nil {
			return fmt.Sprintf("Stock split %s for %s. Your %s shares will be adjusted.",
				action.Ratio.String(), action.Symbol, holding.Quantity.String())
		}
		return fmt.Sprintf("Stock split for %s", action.Symbol)

	case models.CorporateActionTypeDividend:
		if action.Amount != nil {
			totalDividend := action.Amount.Mul(holding.Quantity)
			return fmt.Sprintf("Dividend of %s per share (%s total) for %s",
				action.Amount.String(), totalDividend.String(), action.Symbol)
		}
		return fmt.Sprintf("Dividend for %s", action.Symbol)

	case models.CorporateActionTypeMerger:
		if action.NewSymbol != nil {
			return fmt.Sprintf("Merger: %s is being acquired. Shares will be converted to %s",
				action.Symbol, *action.NewSymbol)
		}
		return fmt.Sprintf("Merger for %s", action.Symbol)

	case models.CorporateActionTypeSpinoff:
		if action.NewSymbol != nil && action.Ratio != nil {
			return fmt.Sprintf("Spinoff: You will receive %s shares of %s for your %s holdings",
				action.Ratio.String(), *action.NewSymbol, action.Symbol)
		}
		return fmt.Sprintf("Spinoff for %s", action.Symbol)

	case models.CorporateActionTypeTickerChange:
		if action.NewSymbol != nil {
			return fmt.Sprintf("Ticker change: %s is changing to %s", action.Symbol, *action.NewSymbol)
		}
		return fmt.Sprintf("Ticker change for %s", action.Symbol)

	default:
		return fmt.Sprintf("Corporate action for %s", action.Symbol)
	}
}

// SimulateDetection creates a simulated corporate action for testing
// In production, this would fetch from an external API
func (m *CorporateActionMonitor) SimulateDetection(
	symbol string,
	actionType models.CorporateActionType,
	date time.Time,
	ratio, amount *decimal.Decimal,
	newSymbol *string,
	description string,
) (*models.CorporateAction, error) {
	// Check if action already exists
	existing, err := m.corporateActionRepo.FindBySymbol(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing actions: %w", err)
	}

	// Check if we already have this action
	for _, act := range existing {
		if act.Type == actionType && act.Date.Equal(date) {
			log.Printf("Corporate action already exists for %s on %s", symbol, date.Format("2006-01-02"))
			return act, nil
		}
	}

	// Create new corporate action
	action := &models.CorporateAction{
		Symbol:      symbol,
		Type:        actionType,
		Date:        date,
		Ratio:       ratio,
		Amount:      amount,
		NewSymbol:   newSymbol,
		Description: description,
		Applied:     false,
	}

	if err := action.Validate(); err != nil {
		return nil, fmt.Errorf("invalid corporate action: %w", err)
	}

	if err := m.corporateActionRepo.Create(action); err != nil {
		return nil, fmt.Errorf("failed to create corporate action: %w", err)
	}

	log.Printf("Created simulated corporate action: %s for %s", actionType, symbol)
	return action, nil
}
