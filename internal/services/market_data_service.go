package services

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/lenon/portfolios/internal/dto"
)

// Quote is an alias for dto.Quote for backward compatibility
type Quote = dto.Quote

// HistoricalPrice is an alias for dto.HistoricalPrice for backward compatibility
type HistoricalPrice = dto.HistoricalPrice

// MarketDataProvider defines the interface for market data providers
type MarketDataProvider interface {
	// GetQuote retrieves a real-time or near real-time quote for a symbol
	GetQuote(ctx context.Context, symbol string) (*Quote, error)

	// GetQuotes retrieves quotes for multiple symbols in batch
	GetQuotes(ctx context.Context, symbols []string) (map[string]*Quote, error)

	// GetHistoricalPrices retrieves historical price data for a symbol
	GetHistoricalPrices(ctx context.Context, symbol string, startDate, endDate time.Time) ([]*HistoricalPrice, error)

	// GetExchangeRate retrieves the exchange rate between two currencies
	GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error)

	// IsAvailable checks if the provider is available and configured
	IsAvailable() bool
}

// MarketDataService provides market data operations with caching
type MarketDataService interface {
	// GetQuote retrieves a quote, using cache if available
	GetQuote(symbol string) (*Quote, error)

	// GetQuotes retrieves multiple quotes
	GetQuotes(symbols []string) (map[string]*Quote, error)

	// GetHistoricalPrices retrieves historical prices
	GetHistoricalPrices(symbol string, startDate, endDate time.Time) ([]*HistoricalPrice, error)

	// GetExchangeRate retrieves an exchange rate
	GetExchangeRate(fromCurrency, toCurrency string) (decimal.Decimal, error)

	// RefreshCache forces a refresh of cached data
	RefreshCache(symbol string) error

	// ClearCache clears all cached data
	ClearCache()
}

// marketDataService implements MarketDataService with caching
type marketDataService struct {
	provider   MarketDataProvider
	cache      map[string]*cachedQuote
	cacheTTL   time.Duration
	defaultCtx context.Context
}

// cachedQuote represents a cached quote with expiration
type cachedQuote struct {
	quote     *Quote
	fetchedAt time.Time
}

// NewMarketDataService creates a new MarketDataService with the specified provider
func NewMarketDataService(provider MarketDataProvider, cacheTTL time.Duration) MarketDataService {
	return &marketDataService{
		provider:   provider,
		cache:      make(map[string]*cachedQuote),
		cacheTTL:   cacheTTL,
		defaultCtx: context.Background(),
	}
}

// GetQuote retrieves a quote with caching
func (s *marketDataService) GetQuote(symbol string) (*Quote, error) {
	// Check cache first
	if cached, exists := s.cache[symbol]; exists {
		if time.Since(cached.fetchedAt) < s.cacheTTL {
			return cached.quote, nil
		}
	}

	// Fetch from provider
	ctx, cancel := context.WithTimeout(s.defaultCtx, 10*time.Second)
	defer cancel()

	quote, err := s.provider.GetQuote(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quote for %s: %w", symbol, err)
	}

	// Cache the result
	s.cache[symbol] = &cachedQuote{
		quote:     quote,
		fetchedAt: time.Now(),
	}

	return quote, nil
}

// GetQuotes retrieves multiple quotes
func (s *marketDataService) GetQuotes(symbols []string) (map[string]*Quote, error) {
	result := make(map[string]*Quote)
	uncachedSymbols := []string{}

	// Check cache for each symbol
	for _, symbol := range symbols {
		if cached, exists := s.cache[symbol]; exists {
			if time.Since(cached.fetchedAt) < s.cacheTTL {
				result[symbol] = cached.quote
				continue
			}
		}
		uncachedSymbols = append(uncachedSymbols, symbol)
	}

	// Fetch uncached symbols
	if len(uncachedSymbols) > 0 {
		ctx, cancel := context.WithTimeout(s.defaultCtx, 30*time.Second)
		defer cancel()

		quotes, err := s.provider.GetQuotes(ctx, uncachedSymbols)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch quotes: %w", err)
		}

		// Cache and add to result
		for symbol, quote := range quotes {
			s.cache[symbol] = &cachedQuote{
				quote:     quote,
				fetchedAt: time.Now(),
			}
			result[symbol] = quote
		}
	}

	return result, nil
}

// GetHistoricalPrices retrieves historical prices (no caching for historical data)
func (s *marketDataService) GetHistoricalPrices(symbol string, startDate, endDate time.Time) ([]*HistoricalPrice, error) {
	ctx, cancel := context.WithTimeout(s.defaultCtx, 30*time.Second)
	defer cancel()

	return s.provider.GetHistoricalPrices(ctx, symbol, startDate, endDate)
}

// GetExchangeRate retrieves an exchange rate
func (s *marketDataService) GetExchangeRate(fromCurrency, toCurrency string) (decimal.Decimal, error) {
	if fromCurrency == toCurrency {
		return decimal.NewFromInt(1), nil
	}

	ctx, cancel := context.WithTimeout(s.defaultCtx, 10*time.Second)
	defer cancel()

	return s.provider.GetExchangeRate(ctx, fromCurrency, toCurrency)
}

// RefreshCache forces a refresh of cached data for a symbol
func (s *marketDataService) RefreshCache(symbol string) error {
	delete(s.cache, symbol)
	_, err := s.GetQuote(symbol)
	return err
}

// ClearCache clears all cached data
func (s *marketDataService) ClearCache() {
	s.cache = make(map[string]*cachedQuote)
}
