package mocks

import (
	"github.com/lenon/portfolios/internal/models"
	"github.com/stretchr/testify/mock"
)

// HoldingRepositoryMock is a mock implementation of HoldingRepository
type HoldingRepositoryMock struct {
	mock.Mock
}

func (m *HoldingRepositoryMock) Create(holding *models.Holding) error {
	args := m.Called(holding)
	return args.Error(0)
}

func (m *HoldingRepositoryMock) FindByID(id string) (*models.Holding, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Holding), args.Error(1)
}

func (m *HoldingRepositoryMock) FindByPortfolioID(portfolioID string) ([]*models.Holding, error) {
	args := m.Called(portfolioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Holding), args.Error(1)
}

func (m *HoldingRepositoryMock) FindByPortfolioIDAndSymbol(portfolioID, symbol string) (*models.Holding, error) {
	args := m.Called(portfolioID, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Holding), args.Error(1)
}

func (m *HoldingRepositoryMock) FindBySymbol(symbol string) ([]*models.Holding, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Holding), args.Error(1)
}

func (m *HoldingRepositoryMock) Update(holding *models.Holding) error {
	args := m.Called(holding)
	return args.Error(0)
}

func (m *HoldingRepositoryMock) Upsert(holding *models.Holding) error {
	args := m.Called(holding)
	return args.Error(0)
}

func (m *HoldingRepositoryMock) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *HoldingRepositoryMock) DeleteByPortfolioIDAndSymbol(portfolioID, symbol string) error {
	args := m.Called(portfolioID, symbol)
	return args.Error(0)
}
