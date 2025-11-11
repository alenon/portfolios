package services

import (
	"fmt"
	"sort"
	"time"

	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/repository"
)

// TaxLotService defines the interface for tax lot operations
type TaxLotService interface {
	// Core operations
	GetByID(id, userID string) (*models.TaxLot, error)
	GetByPortfolioID(portfolioID, userID string) ([]*models.TaxLot, error)
	GetByPortfolioIDAndSymbol(portfolioID, symbol, userID string) ([]*models.TaxLot, error)

	// Tax lot allocation for sales
	AllocateSale(portfolioID, symbol, userID string, quantity decimal.Decimal, method models.CostBasisMethod) ([]*LotAllocation, error)

	// Tax optimization
	IdentifyTaxLossOpportunities(portfolioID, userID string, minLossPercent decimal.Decimal) ([]*TaxLossOpportunity, error)

	// Tax reporting
	GenerateTaxReport(portfolioID, userID string, taxYear int) (*TaxReport, error)
}

// LotAllocation represents how a sale is allocated to tax lots
type LotAllocation struct {
	TaxLot       *models.TaxLot  `json:"tax_lot"`
	Quantity     decimal.Decimal `json:"quantity"`
	CostBasis    decimal.Decimal `json:"cost_basis"`
	SaleProceeds decimal.Decimal `json:"sale_proceeds"`
	Gain         decimal.Decimal `json:"gain"`
	IsLongTerm   bool            `json:"is_long_term"`
}

// TaxLossOpportunity represents a potential tax-loss harvesting opportunity
type TaxLossOpportunity struct {
	Symbol          string          `json:"symbol"`
	CurrentQuantity decimal.Decimal `json:"current_quantity"`
	CostBasis       decimal.Decimal `json:"cost_basis"`
	CurrentValue    decimal.Decimal `json:"current_value"`
	UnrealizedLoss  decimal.Decimal `json:"unrealized_loss"`
	LossPercent     decimal.Decimal `json:"loss_percent"`
}

// TaxReport represents a tax report for a given year
type TaxReport struct {
	Year               int             `json:"year"`
	ShortTermGains     []*RealizedGain `json:"short_term_gains"`
	LongTermGains      []*RealizedGain `json:"long_term_gains"`
	TotalShortTermGain decimal.Decimal `json:"total_short_term_gain"`
	TotalLongTermGain  decimal.Decimal `json:"total_long_term_gain"`
	TotalGain          decimal.Decimal `json:"total_gain"`
}

// RealizedGain represents a realized gain or loss
type RealizedGain struct {
	Symbol       string          `json:"symbol"`
	PurchaseDate time.Time       `json:"purchase_date"`
	SaleDate     time.Time       `json:"sale_date"`
	Quantity     decimal.Decimal `json:"quantity"`
	CostBasis    decimal.Decimal `json:"cost_basis"`
	Proceeds     decimal.Decimal `json:"proceeds"`
	Gain         decimal.Decimal `json:"gain"`
	IsLongTerm   bool            `json:"is_long_term"`
}

// taxLotService implements TaxLotService interface
type taxLotService struct {
	taxLotRepo      repository.TaxLotRepository
	portfolioRepo   repository.PortfolioRepository
	holdingRepo     repository.HoldingRepository
	transactionRepo repository.TransactionRepository
}

// NewTaxLotService creates a new TaxLotService instance
func NewTaxLotService(
	taxLotRepo repository.TaxLotRepository,
	portfolioRepo repository.PortfolioRepository,
	holdingRepo repository.HoldingRepository,
	transactionRepo repository.TransactionRepository,
) TaxLotService {
	return &taxLotService{
		taxLotRepo:      taxLotRepo,
		portfolioRepo:   portfolioRepo,
		holdingRepo:     holdingRepo,
		transactionRepo: transactionRepo,
	}
}

// GetByID retrieves a tax lot by ID, ensuring it belongs to the user
func (s *taxLotService) GetByID(id, userID string) (*models.TaxLot, error) {
	taxLot, err := s.taxLotRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Verify the portfolio belongs to the user
	portfolio, err := s.portfolioRepo.FindByID(taxLot.PortfolioID.String())
	if err != nil {
		return nil, models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	return taxLot, nil
}

// GetByPortfolioID retrieves all tax lots for a portfolio
func (s *taxLotService) GetByPortfolioID(portfolioID, userID string) ([]*models.TaxLot, error) {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return nil, models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	taxLots, err := s.taxLotRepo.FindByPortfolioID(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tax lots: %w", err)
	}

	return taxLots, nil
}

// GetByPortfolioIDAndSymbol retrieves all tax lots for a portfolio and symbol
func (s *taxLotService) GetByPortfolioIDAndSymbol(portfolioID, symbol, userID string) ([]*models.TaxLot, error) {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return nil, models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	taxLots, err := s.taxLotRepo.FindByPortfolioIDAndSymbol(portfolioID, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tax lots: %w", err)
	}

	return taxLots, nil
}

// AllocateSale allocates a sale to tax lots based on the specified cost basis method
func (s *taxLotService) AllocateSale(
	portfolioID, symbol, userID string,
	quantity decimal.Decimal,
	method models.CostBasisMethod,
) ([]*LotAllocation, error) {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return nil, models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	// Get all tax lots for this symbol
	taxLots, err := s.taxLotRepo.FindByPortfolioIDAndSymbol(portfolioID, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tax lots: %w", err)
	}

	if len(taxLots) == 0 {
		return nil, models.ErrInsufficientShares
	}

	// Sort tax lots based on cost basis method
	sortTaxLots(taxLots, method)

	// Allocate the sale across tax lots
	allocations := make([]*LotAllocation, 0)
	remainingQuantity := quantity

	for _, lot := range taxLots {
		if remainingQuantity.IsZero() {
			break
		}

		// Determine how much to take from this lot
		quantityFromLot := lot.Quantity
		if quantityFromLot.GreaterThan(remainingQuantity) {
			quantityFromLot = remainingQuantity
		}

		// Calculate cost basis for this allocation
		costBasisPerShare := lot.GetCostPerShare()
		costBasis := costBasisPerShare.Mul(quantityFromLot)

		allocation := &LotAllocation{
			TaxLot:     lot,
			Quantity:   quantityFromLot,
			CostBasis:  costBasis,
			IsLongTerm: lot.IsLongTerm(time.Now().UTC()),
		}

		allocations = append(allocations, allocation)
		remainingQuantity = remainingQuantity.Sub(quantityFromLot)
	}

	// Check if we have enough shares
	if remainingQuantity.GreaterThan(decimal.Zero) {
		return nil, models.ErrInsufficientShares
	}

	return allocations, nil
}

// sortTaxLots sorts tax lots based on the cost basis method
func sortTaxLots(taxLots []*models.TaxLot, method models.CostBasisMethod) {
	switch method {
	case models.CostBasisFIFO:
		// Sort by purchase date (oldest first)
		sort.Slice(taxLots, func(i, j int) bool {
			return taxLots[i].PurchaseDate.Before(taxLots[j].PurchaseDate)
		})
	case models.CostBasisLIFO:
		// Sort by purchase date (newest first)
		sort.Slice(taxLots, func(i, j int) bool {
			return taxLots[i].PurchaseDate.After(taxLots[j].PurchaseDate)
		})
	case models.CostBasisSpecificLot:
		// For specific lot, no sorting needed - user will specify
		// This would be handled differently in a real implementation
	}
}

// IdentifyTaxLossOpportunities identifies holdings with unrealized losses
func (s *taxLotService) IdentifyTaxLossOpportunities(
	portfolioID, userID string,
	minLossPercent decimal.Decimal,
) ([]*TaxLossOpportunity, error) {
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

	opportunities := make([]*TaxLossOpportunity, 0)

	// For each holding, calculate unrealized loss
	// Note: In a real implementation, we would fetch current market prices
	// For now, we'll use a placeholder approach
	for _, holding := range holdings {
		// Skip holdings without losses
		// This is a placeholder - in production we'd compare with current market price
		// For now, we'll just create opportunities for demonstration
		if holding.CostBasis.IsPositive() {
			// Placeholder: assume current value is 10% less than cost basis for demonstration
			currentValue := holding.CostBasis.Mul(decimal.NewFromFloat(0.9))
			unrealizedLoss := currentValue.Sub(holding.CostBasis)
			lossPercent := unrealizedLoss.Div(holding.CostBasis).Mul(decimal.NewFromInt(100))

			if lossPercent.LessThan(minLossPercent.Neg()) {
				opportunity := &TaxLossOpportunity{
					Symbol:          holding.Symbol,
					CurrentQuantity: holding.Quantity,
					CostBasis:       holding.CostBasis,
					CurrentValue:    currentValue,
					UnrealizedLoss:  unrealizedLoss,
					LossPercent:     lossPercent,
				}
				opportunities = append(opportunities, opportunity)
			}
		}
	}

	return opportunities, nil
}

// GenerateTaxReport generates a tax report for a given year
func (s *taxLotService) GenerateTaxReport(portfolioID, userID string, taxYear int) (*TaxReport, error) {
	// Verify portfolio exists and belongs to user
	portfolio, err := s.portfolioRepo.FindByID(portfolioID)
	if err != nil {
		return nil, models.ErrPortfolioNotFound
	}
	if portfolio.UserID.String() != userID {
		return nil, models.ErrUnauthorizedAccess
	}

	// Initialize report
	report := &TaxReport{
		Year:               taxYear,
		ShortTermGains:     make([]*RealizedGain, 0),
		LongTermGains:      make([]*RealizedGain, 0),
		TotalShortTermGain: decimal.Zero,
		TotalLongTermGain:  decimal.Zero,
		TotalGain:          decimal.Zero,
	}

	// Get the start and end dates for the tax year
	startDate := time.Date(taxYear, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(taxYear, 12, 31, 23, 59, 59, 999999999, time.UTC)

	// Get all transactions for this portfolio in the tax year
	allTransactions, err := s.transactionRepo.FindByPortfolioIDWithFilters(portfolioID, nil, &startDate, &endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}

	// Filter for sell transactions only
	var sellTransactions []*models.Transaction
	for _, tx := range allTransactions {
		if tx.Type == models.TransactionTypeSell {
			sellTransactions = append(sellTransactions, tx)
		}
	}

	// For each sell transaction, allocate to tax lots and calculate gain/loss
	for _, sellTx := range sellTransactions {
		if sellTx.Price == nil {
			continue // Skip sells without a price
		}

		salePrice := *sellTx.Price

		// Get tax lots at the time of the sale to allocate
		taxLots, err := s.taxLotRepo.FindByPortfolioIDAndSymbol(portfolioID, sellTx.Symbol)
		if err != nil {
			continue // Skip if we can't find tax lots
		}

		// Sort lots by purchase date (FIFO)
		sortTaxLots(taxLots, portfolio.CostBasisMethod)

		// Allocate the sale to lots
		remainingQuantity := sellTx.Quantity
		totalCostBasis := decimal.Zero

		for _, lot := range taxLots {
			if remainingQuantity.IsZero() {
				break
			}

			// Skip lots purchased after this sale
			if lot.PurchaseDate.After(sellTx.Date) {
				continue
			}

			// Determine how much to take from this lot
			quantityFromLot := lot.Quantity
			if quantityFromLot.GreaterThan(remainingQuantity) {
				quantityFromLot = remainingQuantity
			}

			// Calculate cost basis for this portion
			costBasisPerShare := lot.GetCostPerShare()
			costBasisForPortion := costBasisPerShare.Mul(quantityFromLot)
			totalCostBasis = totalCostBasis.Add(costBasisForPortion)

			// Calculate proceeds for this portion
			proceedsForPortion := salePrice.Mul(quantityFromLot)

			// Calculate gain/loss
			gain := proceedsForPortion.Sub(costBasisForPortion)

			// Determine if long-term or short-term
			isLongTerm := lot.IsLongTerm(sellTx.Date)

			// Create realized gain entry
			realizedGain := &RealizedGain{
				Symbol:       sellTx.Symbol,
				PurchaseDate: lot.PurchaseDate,
				SaleDate:     sellTx.Date,
				Quantity:     quantityFromLot,
				CostBasis:    costBasisForPortion,
				Proceeds:     proceedsForPortion,
				Gain:         gain,
				IsLongTerm:   isLongTerm,
			}

			// Add to appropriate category
			if isLongTerm {
				report.LongTermGains = append(report.LongTermGains, realizedGain)
				report.TotalLongTermGain = report.TotalLongTermGain.Add(gain)
			} else {
				report.ShortTermGains = append(report.ShortTermGains, realizedGain)
				report.TotalShortTermGain = report.TotalShortTermGain.Add(gain)
			}

			remainingQuantity = remainingQuantity.Sub(quantityFromLot)
		}
	}

	// Calculate total gain
	report.TotalGain = report.TotalShortTermGain.Add(report.TotalLongTermGain)

	return report, nil
}
