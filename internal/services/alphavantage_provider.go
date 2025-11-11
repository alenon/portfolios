package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	alphaVantageBaseURL = "https://www.alphavantage.co/query"
)

// AlphaVantageProvider implements MarketDataProvider using Alpha Vantage API
type AlphaVantageProvider struct {
	apiKey     string
	httpClient *http.Client
}

// NewAlphaVantageProvider creates a new Alpha Vantage market data provider
func NewAlphaVantageProvider(apiKey string) *AlphaVantageProvider {
	return &AlphaVantageProvider{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsAvailable checks if the provider is configured with an API key
func (p *AlphaVantageProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// GetQuote retrieves a real-time quote from Alpha Vantage
func (p *AlphaVantageProvider) GetQuote(ctx context.Context, symbol string) (*Quote, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("alpha Vantage API key not configured")
	}

	// Build request URL
	params := url.Values{}
	params.Set("function", "GLOBAL_QUOTE")
	params.Set("symbol", symbol)
	params.Set("apikey", p.apiKey)

	reqURL := fmt.Sprintf("%s?%s", alphaVantageBaseURL, params.Encode())

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quote: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		GlobalQuote struct {
			Symbol           string `json:"01. symbol"`
			Open             string `json:"02. open"`
			High             string `json:"03. high"`
			Low              string `json:"04. low"`
			Price            string `json:"05. price"`
			Volume           string `json:"06. volume"`
			LatestTradingDay string `json:"07. latest trading day"`
			PreviousClose    string `json:"08. previous close"`
			Change           string `json:"09. change"`
			ChangePercent    string `json:"10. change percent"`
		} `json:"Global Quote"`
		ErrorMessage string `json:"Error Message"`
		Note         string `json:"Note"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if result.ErrorMessage != "" {
		return nil, fmt.Errorf("API error: %s", result.ErrorMessage)
	}
	if result.Note != "" {
		return nil, fmt.Errorf("API rate limit exceeded: %s", result.Note)
	}

	// Check if quote data is empty
	if result.GlobalQuote.Symbol == "" {
		return nil, fmt.Errorf("no data found for symbol %s", symbol)
	}

	// Parse quote data
	quote := &Quote{
		Symbol:      result.GlobalQuote.Symbol,
		LastUpdated: time.Now(),
	}

	if price, err := decimal.NewFromString(result.GlobalQuote.Price); err == nil {
		quote.Price = price
	}
	if open, err := decimal.NewFromString(result.GlobalQuote.Open); err == nil {
		quote.Open = open
	}
	if high, err := decimal.NewFromString(result.GlobalQuote.High); err == nil {
		quote.High = high
	}
	if low, err := decimal.NewFromString(result.GlobalQuote.Low); err == nil {
		quote.Low = low
	}
	if prevClose, err := decimal.NewFromString(result.GlobalQuote.PreviousClose); err == nil {
		quote.PreviousClose = prevClose
	}
	if change, err := decimal.NewFromString(result.GlobalQuote.Change); err == nil {
		quote.Change = change
	}
	if changePctStr := strings.TrimSuffix(result.GlobalQuote.ChangePercent, "%"); changePctStr != "" {
		if changePct, err := decimal.NewFromString(changePctStr); err == nil {
			quote.ChangePercent = changePct
		}
	}
	if volume, err := strconv.ParseInt(result.GlobalQuote.Volume, 10, 64); err == nil {
		quote.Volume = volume
	}

	return quote, nil
}

// GetQuotes retrieves multiple quotes (Alpha Vantage doesn't support batch requests natively)
func (p *AlphaVantageProvider) GetQuotes(ctx context.Context, symbols []string) (map[string]*Quote, error) {
	result := make(map[string]*Quote)

	// Alpha Vantage free tier has rate limits (5 requests/minute, 500/day)
	// So we fetch quotes sequentially with a small delay
	for i, symbol := range symbols {
		quote, err := p.GetQuote(ctx, symbol)
		if err != nil {
			// Continue on error, just skip this symbol
			continue
		}
		result[symbol] = quote

		// Add delay between requests to avoid rate limiting (except for last request)
		if i < len(symbols)-1 {
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(12 * time.Second): // 5 requests per minute = 12 seconds between requests
			}
		}
	}

	return result, nil
}

// GetHistoricalPrices retrieves historical price data from Alpha Vantage
func (p *AlphaVantageProvider) GetHistoricalPrices(ctx context.Context, symbol string, startDate, endDate time.Time) ([]*HistoricalPrice, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("alpha Vantage API key not configured")
	}

	// Build request URL for daily time series
	params := url.Values{}
	params.Set("function", "TIME_SERIES_DAILY_ADJUSTED")
	params.Set("symbol", symbol)
	params.Set("outputsize", "full") // Get full history
	params.Set("apikey", p.apiKey)

	reqURL := fmt.Sprintf("%s?%s", alphaVantageBaseURL, params.Encode())

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical data: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		TimeSeriesDaily map[string]struct {
			Open          string `json:"1. open"`
			High          string `json:"2. high"`
			Low           string `json:"3. low"`
			Close         string `json:"4. close"`
			AdjustedClose string `json:"5. adjusted close"`
			Volume        string `json:"6. volume"`
		} `json:"Time Series (Daily)"`
		ErrorMessage string `json:"Error Message"`
		Note         string `json:"Note"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if result.ErrorMessage != "" {
		return nil, fmt.Errorf("API error: %s", result.ErrorMessage)
	}
	if result.Note != "" {
		return nil, fmt.Errorf("API rate limit exceeded: %s", result.Note)
	}

	// Parse historical data
	var prices []*HistoricalPrice
	for dateStr, data := range result.TimeSeriesDaily {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		// Filter by date range
		if date.Before(startDate) || date.After(endDate) {
			continue
		}

		price := &HistoricalPrice{
			Date: date,
		}

		if open, err := decimal.NewFromString(data.Open); err == nil {
			price.Open = open
		}
		if high, err := decimal.NewFromString(data.High); err == nil {
			price.High = high
		}
		if low, err := decimal.NewFromString(data.Low); err == nil {
			price.Low = low
		}
		if close, err := decimal.NewFromString(data.Close); err == nil {
			price.Close = close
		}
		if adjClose, err := decimal.NewFromString(data.AdjustedClose); err == nil {
			price.AdjClose = &adjClose
		}
		if volume, err := strconv.ParseInt(data.Volume, 10, 64); err == nil {
			price.Volume = volume
		}

		prices = append(prices, price)
	}

	return prices, nil
}

// GetExchangeRate retrieves the exchange rate between two currencies
func (p *AlphaVantageProvider) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error) {
	if !p.IsAvailable() {
		return decimal.Zero, fmt.Errorf("alpha Vantage API key not configured")
	}

	if fromCurrency == toCurrency {
		return decimal.NewFromInt(1), nil
	}

	// Build request URL
	params := url.Values{}
	params.Set("function", "CURRENCY_EXCHANGE_RATE")
	params.Set("from_currency", fromCurrency)
	params.Set("to_currency", toCurrency)
	params.Set("apikey", p.apiKey)

	reqURL := fmt.Sprintf("%s?%s", alphaVantageBaseURL, params.Encode())

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to fetch exchange rate: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		RealtimeCurrencyExchangeRate struct {
			ExchangeRate string `json:"5. Exchange Rate"`
		} `json:"Realtime Currency Exchange Rate"`
		ErrorMessage string `json:"Error Message"`
		Note         string `json:"Note"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return decimal.Zero, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if result.ErrorMessage != "" {
		return decimal.Zero, fmt.Errorf("API error: %s", result.ErrorMessage)
	}
	if result.Note != "" {
		return decimal.Zero, fmt.Errorf("API rate limit exceeded: %s", result.Note)
	}

	// Parse exchange rate
	rate, err := decimal.NewFromString(result.RealtimeCurrencyExchangeRate.ExchangeRate)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to parse exchange rate: %w", err)
	}

	return rate, nil
}
