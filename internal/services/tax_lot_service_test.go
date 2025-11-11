package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lenon/portfolios/internal/mocks"
	"github.com/lenon/portfolios/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewTaxLotService(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	assert.NotNil(t, service)
}

func TestTaxLotService_GetByID_Success(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	userID := uuid.New()
	portfolioID := uuid.New()
	taxLotID := uuid.New()

	taxLot := &models.TaxLot{
		ID:          taxLotID,
		PortfolioID: portfolioID,
		Symbol:      "AAPL",
		Quantity:    decimal.NewFromInt(10),
		CostBasis:   decimal.NewFromInt(1000),
	}

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: userID,
		Name:   "Test Portfolio",
	}

	taxLotRepo.On("FindByID", taxLotID.String()).Return(taxLot, nil)
	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)

	result, err := service.GetByID(taxLotID.String(), userID.String())

	assert.NoError(t, err)
	assert.Equal(t, taxLot, result)
	taxLotRepo.AssertExpectations(t)
	portfolioRepo.AssertExpectations(t)
}

func TestTaxLotService_GetByID_TaxLotNotFound(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	taxLotID := uuid.New().String()
	userID := uuid.New().String()

	taxLotRepo.On("FindByID", taxLotID).Return(nil, fmt.Errorf("not found"))

	_, err := service.GetByID(taxLotID, userID)

	assert.Error(t, err)
	taxLotRepo.AssertExpectations(t)
}

func TestTaxLotService_GetByID_PortfolioNotFound(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	userID := uuid.New()
	portfolioID := uuid.New()
	taxLotID := uuid.New()

	taxLot := &models.TaxLot{
		ID:          taxLotID,
		PortfolioID: portfolioID,
		Symbol:      "AAPL",
	}

	taxLotRepo.On("FindByID", taxLotID.String()).Return(taxLot, nil)
	portfolioRepo.On("FindByID", portfolioID.String()).Return(nil, fmt.Errorf("not found"))

	_, err := service.GetByID(taxLotID.String(), userID.String())

	assert.Error(t, err)
	assert.Equal(t, models.ErrPortfolioNotFound, err)
	taxLotRepo.AssertExpectations(t)
	portfolioRepo.AssertExpectations(t)
}

func TestTaxLotService_GetByID_Unauthorized(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	userID := uuid.New()
	otherUserID := uuid.New()
	portfolioID := uuid.New()
	taxLotID := uuid.New()

	taxLot := &models.TaxLot{
		ID:          taxLotID,
		PortfolioID: portfolioID,
		Symbol:      "AAPL",
	}

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: otherUserID, // Different user
		Name:   "Test Portfolio",
	}

	taxLotRepo.On("FindByID", taxLotID.String()).Return(taxLot, nil)
	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)

	_, err := service.GetByID(taxLotID.String(), userID.String())

	assert.Error(t, err)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)
	taxLotRepo.AssertExpectations(t)
	portfolioRepo.AssertExpectations(t)
}

func TestTaxLotService_GetByPortfolioID_Success(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	userID := uuid.New()
	portfolioID := uuid.New()

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: userID,
		Name:   "Test Portfolio",
	}

	taxLots := []*models.TaxLot{
		{
			ID:          uuid.New(),
			PortfolioID: portfolioID,
			Symbol:      "AAPL",
			Quantity:    decimal.NewFromInt(10),
		},
		{
			ID:          uuid.New(),
			PortfolioID: portfolioID,
			Symbol:      "MSFT",
			Quantity:    decimal.NewFromInt(5),
		},
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)
	taxLotRepo.On("FindByPortfolioID", portfolioID.String()).Return(taxLots, nil)

	result, err := service.GetByPortfolioID(portfolioID.String(), userID.String())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	portfolioRepo.AssertExpectations(t)
	taxLotRepo.AssertExpectations(t)
}

func TestTaxLotService_GetByPortfolioID_Unauthorized(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	userID := uuid.New()
	otherUserID := uuid.New()
	portfolioID := uuid.New()

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: otherUserID, // Different user
		Name:   "Test Portfolio",
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)

	_, err := service.GetByPortfolioID(portfolioID.String(), userID.String())

	assert.Error(t, err)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)
	portfolioRepo.AssertExpectations(t)
}

func TestTaxLotService_GetByPortfolioIDAndSymbol_Success(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	userID := uuid.New()
	portfolioID := uuid.New()
	symbol := "AAPL"

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: userID,
		Name:   "Test Portfolio",
	}

	taxLots := []*models.TaxLot{
		{
			ID:          uuid.New(),
			PortfolioID: portfolioID,
			Symbol:      symbol,
			Quantity:    decimal.NewFromInt(10),
		},
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)
	taxLotRepo.On("FindByPortfolioIDAndSymbol", portfolioID.String(), symbol).Return(taxLots, nil)

	result, err := service.GetByPortfolioIDAndSymbol(portfolioID.String(), symbol, userID.String())

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, symbol, result[0].Symbol)
	portfolioRepo.AssertExpectations(t)
	taxLotRepo.AssertExpectations(t)
}

func TestTaxLotService_AllocateSale_FIFO(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	userID := uuid.New()
	portfolioID := uuid.New()
	symbol := "AAPL"

	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          userID,
		Name:            "Test Portfolio",
		CostBasisMethod: models.CostBasisFIFO,
	}

	// Create tax lots with different purchase dates
	oldDate := time.Now().AddDate(0, 0, -10)
	newDate := time.Now().AddDate(0, 0, -5)

	taxLots := []*models.TaxLot{
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Symbol:       symbol,
			Quantity:     decimal.NewFromInt(10),
			CostBasis:    decimal.NewFromInt(1000),
			PurchaseDate: newDate, // Newer
		},
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Symbol:       symbol,
			Quantity:     decimal.NewFromInt(5),
			CostBasis:    decimal.NewFromInt(400),
			PurchaseDate: oldDate, // Older
		},
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)
	taxLotRepo.On("FindByPortfolioIDAndSymbol", portfolioID.String(), symbol).Return(taxLots, nil)

	// Sell 8 shares - should use FIFO (oldest first)
	allocations, err := service.AllocateSale(
		portfolioID.String(),
		symbol,
		userID.String(),
		decimal.NewFromInt(8),
		models.CostBasisFIFO,
	)

	assert.NoError(t, err)
	assert.Len(t, allocations, 2)
	// First allocation should be from older lot (5 shares)
	assert.Equal(t, decimal.NewFromInt(5), allocations[0].Quantity)
	assert.Equal(t, oldDate, allocations[0].TaxLot.PurchaseDate)
	// Second allocation should be from newer lot (3 shares)
	assert.Equal(t, decimal.NewFromInt(3), allocations[1].Quantity)
	assert.Equal(t, newDate, allocations[1].TaxLot.PurchaseDate)

	portfolioRepo.AssertExpectations(t)
	taxLotRepo.AssertExpectations(t)
}

func TestTaxLotService_AllocateSale_LIFO(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	userID := uuid.New()
	portfolioID := uuid.New()
	symbol := "AAPL"

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: userID,
		Name:   "Test Portfolio",
	}

	oldDate := time.Now().AddDate(0, 0, -10)
	newDate := time.Now().AddDate(0, 0, -5)

	taxLots := []*models.TaxLot{
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Symbol:       symbol,
			Quantity:     decimal.NewFromInt(5),
			CostBasis:    decimal.NewFromInt(400),
			PurchaseDate: oldDate, // Older
		},
		{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			Symbol:       symbol,
			Quantity:     decimal.NewFromInt(10),
			CostBasis:    decimal.NewFromInt(1000),
			PurchaseDate: newDate, // Newer
		},
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)
	taxLotRepo.On("FindByPortfolioIDAndSymbol", portfolioID.String(), symbol).Return(taxLots, nil)

	// Sell 8 shares - should use LIFO (newest first)
	allocations, err := service.AllocateSale(
		portfolioID.String(),
		symbol,
		userID.String(),
		decimal.NewFromInt(8),
		models.CostBasisLIFO,
	)

	assert.NoError(t, err)
	assert.Len(t, allocations, 1)
	// Should use only the newer lot
	assert.Equal(t, decimal.NewFromInt(8), allocations[0].Quantity)
	assert.Equal(t, newDate, allocations[0].TaxLot.PurchaseDate)

	portfolioRepo.AssertExpectations(t)
	taxLotRepo.AssertExpectations(t)
}

func TestTaxLotService_AllocateSale_InsufficientShares(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	userID := uuid.New()
	portfolioID := uuid.New()
	symbol := "AAPL"

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: userID,
		Name:   "Test Portfolio",
	}

	taxLots := []*models.TaxLot{
		{
			ID:          uuid.New(),
			PortfolioID: portfolioID,
			Symbol:      symbol,
			Quantity:    decimal.NewFromInt(5),
			CostBasis:   decimal.NewFromInt(500),
		},
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)
	taxLotRepo.On("FindByPortfolioIDAndSymbol", portfolioID.String(), symbol).Return(taxLots, nil)

	// Try to sell 10 shares when only 5 available
	_, err := service.AllocateSale(
		portfolioID.String(),
		symbol,
		userID.String(),
		decimal.NewFromInt(10),
		models.CostBasisFIFO,
	)

	assert.Error(t, err)
	assert.Equal(t, models.ErrInsufficientShares, err)

	portfolioRepo.AssertExpectations(t)
	taxLotRepo.AssertExpectations(t)
}

func TestTaxLotService_AllocateSale_NoTaxLots(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	userID := uuid.New()
	portfolioID := uuid.New()
	symbol := "AAPL"

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: userID,
		Name:   "Test Portfolio",
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)
	taxLotRepo.On("FindByPortfolioIDAndSymbol", portfolioID.String(), symbol).Return([]*models.TaxLot{}, nil)

	_, err := service.AllocateSale(
		portfolioID.String(),
		symbol,
		userID.String(),
		decimal.NewFromInt(10),
		models.CostBasisFIFO,
	)

	assert.Error(t, err)
	assert.Equal(t, models.ErrInsufficientShares, err)

	portfolioRepo.AssertExpectations(t)
	taxLotRepo.AssertExpectations(t)
}

func TestTaxLotService_IdentifyTaxLossOpportunities_Success(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	userID := uuid.New()
	portfolioID := uuid.New()

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: userID,
		Name:   "Test Portfolio",
	}

	holdings := []*models.Holding{
		{
			ID:          uuid.New(),
			PortfolioID: portfolioID,
			Symbol:      "AAPL",
			Quantity:    decimal.NewFromInt(10),
			CostBasis:   decimal.NewFromInt(1000),
		},
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)
	holdingRepo.On("FindByPortfolioID", portfolioID.String()).Return(holdings, nil)

	opportunities, err := service.IdentifyTaxLossOpportunities(
		portfolioID.String(),
		userID.String(),
		decimal.NewFromInt(5), // Min 5% loss
	)

	assert.NoError(t, err)
	assert.NotNil(t, opportunities)
	// The service creates a placeholder opportunity with -10% loss
	assert.Len(t, opportunities, 1)

	portfolioRepo.AssertExpectations(t)
	holdingRepo.AssertExpectations(t)
}

func TestTaxLotService_IdentifyTaxLossOpportunities_Unauthorized(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	userID := uuid.New()
	otherUserID := uuid.New()
	portfolioID := uuid.New()

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: otherUserID, // Different user
		Name:   "Test Portfolio",
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)

	_, err := service.IdentifyTaxLossOpportunities(
		portfolioID.String(),
		userID.String(),
		decimal.NewFromInt(5),
	)

	assert.Error(t, err)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)

	portfolioRepo.AssertExpectations(t)
}

func TestTaxLotService_GenerateTaxReport_Success(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	userID := uuid.New()
	portfolioID := uuid.New()

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: userID,
		Name:   "Test Portfolio",
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)

	report, err := service.GenerateTaxReport(
		portfolioID.String(),
		userID.String(),
		2024,
	)

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 2024, report.Year)
	assert.NotNil(t, report.ShortTermGains)
	assert.NotNil(t, report.LongTermGains)

	portfolioRepo.AssertExpectations(t)
}

func TestTaxLotService_GenerateTaxReport_PortfolioNotFound(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	userID := uuid.New()
	portfolioID := uuid.New()

	portfolioRepo.On("FindByID", portfolioID.String()).Return(nil, fmt.Errorf("not found"))

	_, err := service.GenerateTaxReport(
		portfolioID.String(),
		userID.String(),
		2024,
	)

	assert.Error(t, err)
	assert.Equal(t, models.ErrPortfolioNotFound, err)

	portfolioRepo.AssertExpectations(t)
}

func TestTaxLotService_GenerateTaxReport_Unauthorized(t *testing.T) {
	taxLotRepo := new(mocks.TaxLotRepositoryMock)
	portfolioRepo := new(mocks.PortfolioRepositoryMock)
	holdingRepo := new(mocks.HoldingRepositoryMock)

	service := NewTaxLotService(taxLotRepo, portfolioRepo, holdingRepo)

	userID := uuid.New()
	otherUserID := uuid.New()
	portfolioID := uuid.New()

	portfolio := &models.Portfolio{
		ID:     portfolioID,
		UserID: otherUserID, // Different user
		Name:   "Test Portfolio",
	}

	portfolioRepo.On("FindByID", portfolioID.String()).Return(portfolio, nil)

	_, err := service.GenerateTaxReport(
		portfolioID.String(),
		userID.String(),
		2024,
	)

	assert.Error(t, err)
	assert.Equal(t, models.ErrUnauthorizedAccess, err)

	portfolioRepo.AssertExpectations(t)
}

func TestSortTaxLots_FIFO(t *testing.T) {
	now := time.Now()
	taxLots := []*models.TaxLot{
		{PurchaseDate: now.AddDate(0, 0, -1)}, // Yesterday
		{PurchaseDate: now.AddDate(0, 0, -5)}, // 5 days ago
		{PurchaseDate: now.AddDate(0, 0, -3)}, // 3 days ago
	}

	sortTaxLots(taxLots, models.CostBasisFIFO)

	// Should be sorted oldest first
	assert.True(t, taxLots[0].PurchaseDate.Before(taxLots[1].PurchaseDate))
	assert.True(t, taxLots[1].PurchaseDate.Before(taxLots[2].PurchaseDate))
}

func TestSortTaxLots_LIFO(t *testing.T) {
	now := time.Now()
	taxLots := []*models.TaxLot{
		{PurchaseDate: now.AddDate(0, 0, -5)}, // 5 days ago
		{PurchaseDate: now.AddDate(0, 0, -1)}, // Yesterday
		{PurchaseDate: now.AddDate(0, 0, -3)}, // 3 days ago
	}

	sortTaxLots(taxLots, models.CostBasisLIFO)

	// Should be sorted newest first
	assert.True(t, taxLots[0].PurchaseDate.After(taxLots[1].PurchaseDate))
	assert.True(t, taxLots[1].PurchaseDate.After(taxLots[2].PurchaseDate))
}

func TestSortTaxLots_SpecificLot(t *testing.T) {
	now := time.Now()
	taxLots := []*models.TaxLot{
		{PurchaseDate: now.AddDate(0, 0, -1)},
		{PurchaseDate: now.AddDate(0, 0, -5)},
		{PurchaseDate: now.AddDate(0, 0, -3)},
	}

	originalOrder := []time.Time{
		taxLots[0].PurchaseDate,
		taxLots[1].PurchaseDate,
		taxLots[2].PurchaseDate,
	}

	sortTaxLots(taxLots, models.CostBasisSpecificLot)

	// Should maintain original order for specific lot
	assert.Equal(t, originalOrder[0], taxLots[0].PurchaseDate)
	assert.Equal(t, originalOrder[1], taxLots[1].PurchaseDate)
	assert.Equal(t, originalOrder[2], taxLots[2].PurchaseDate)
}
