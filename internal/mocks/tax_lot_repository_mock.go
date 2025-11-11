package mocks

import (
	"github.com/lenon/portfolios/internal/models"
	"github.com/stretchr/testify/mock"
)

// TaxLotRepositoryMock is a mock implementation of TaxLotRepository
type TaxLotRepositoryMock struct {
	mock.Mock
}

func (m *TaxLotRepositoryMock) Create(taxLot *models.TaxLot) error {
	args := m.Called(taxLot)
	return args.Error(0)
}

func (m *TaxLotRepositoryMock) FindByID(id string) (*models.TaxLot, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaxLot), args.Error(1)
}

func (m *TaxLotRepositoryMock) FindByPortfolioID(portfolioID string) ([]*models.TaxLot, error) {
	args := m.Called(portfolioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TaxLot), args.Error(1)
}

func (m *TaxLotRepositoryMock) FindByPortfolioIDAndSymbol(portfolioID, symbol string) ([]*models.TaxLot, error) {
	args := m.Called(portfolioID, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TaxLot), args.Error(1)
}

func (m *TaxLotRepositoryMock) FindByTransactionID(transactionID string) ([]*models.TaxLot, error) {
	args := m.Called(transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TaxLot), args.Error(1)
}

func (m *TaxLotRepositoryMock) Update(taxLot *models.TaxLot) error {
	args := m.Called(taxLot)
	return args.Error(0)
}

func (m *TaxLotRepositoryMock) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *TaxLotRepositoryMock) DeleteByPortfolioIDAndSymbol(portfolioID, symbol string) error {
	args := m.Called(portfolioID, symbol)
	return args.Error(0)
}
