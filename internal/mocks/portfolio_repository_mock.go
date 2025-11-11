package mocks

import (
	"github.com/lenon/portfolios/internal/models"
	"github.com/stretchr/testify/mock"
)

// PortfolioRepositoryMock is a mock implementation of PortfolioRepository
type PortfolioRepositoryMock struct {
	mock.Mock
}

func (m *PortfolioRepositoryMock) Create(portfolio *models.Portfolio) error {
	args := m.Called(portfolio)
	return args.Error(0)
}

func (m *PortfolioRepositoryMock) FindByID(id string) (*models.Portfolio, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func (m *PortfolioRepositoryMock) FindByUserID(userID string) ([]*models.Portfolio, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Portfolio), args.Error(1)
}

func (m *PortfolioRepositoryMock) FindByUserIDAndName(userID, name string) (*models.Portfolio, error) {
	args := m.Called(userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func (m *PortfolioRepositoryMock) Update(portfolio *models.Portfolio) error {
	args := m.Called(portfolio)
	return args.Error(0)
}

func (m *PortfolioRepositoryMock) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *PortfolioRepositoryMock) ExistsByUserIDAndName(userID, name string) (bool, error) {
	args := m.Called(userID, name)
	return args.Bool(0), args.Error(1)
}
