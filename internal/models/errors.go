package models

import "errors"

// User-related errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// Portfolio-related errors
var (
	ErrPortfolioNotFound        = errors.New("portfolio not found")
	ErrPortfolioNameRequired    = errors.New("portfolio name is required")
	ErrInvalidCurrency          = errors.New("invalid currency code")
	ErrInvalidCostBasisMethod   = errors.New("invalid cost basis method")
	ErrPortfolioDuplicateName   = errors.New("portfolio with this name already exists")
	ErrUnauthorizedAccess       = errors.New("unauthorized access to portfolio")
)

// Transaction-related errors
var (
	ErrTransactionNotFound      = errors.New("transaction not found")
	ErrInvalidTransactionType   = errors.New("invalid transaction type")
	ErrInvalidQuantity          = errors.New("invalid quantity")
	ErrInvalidPrice             = errors.New("invalid price")
	ErrInvalidSymbol            = errors.New("invalid symbol")
	ErrInsufficientShares       = errors.New("insufficient shares for sale")
)

// Holding-related errors
var (
	ErrHoldingNotFound = errors.New("holding not found")
)

// Tax lot-related errors
var (
	ErrTaxLotNotFound = errors.New("tax lot not found")
)

// Corporate action-related errors
var (
	ErrCorporateActionNotFound = errors.New("corporate action not found")
	ErrInvalidCorporateActionType = errors.New("invalid corporate action type")
)
