package dto

import (
	"time"

	"github.com/lenon/portfolios/internal/services"
	"github.com/shopspring/decimal"
)

// GetQuotesRequest represents request for multiple quotes
type GetQuotesRequest struct {
	Symbols []string `json:"symbols" binding:"required"`
}

// HistoricalPricesRequest represents request parameters for historical prices
type HistoricalPricesRequest struct {
	StartDate time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate   time.Time `form:"end_date" time_format:"2006-01-02"`
}

// ExchangeRateRequest represents request parameters for exchange rate
type ExchangeRateRequest struct {
	From string `form:"from" binding:"required,len=3"`
	To   string `form:"to" binding:"required,len=3"`
}

// QuoteResponse represents a stock quote response
type QuoteResponse struct {
	Symbol           string          `json:"symbol"`
	Price            decimal.Decimal `json:"price"`
	Open             decimal.Decimal `json:"open"`
	High             decimal.Decimal `json:"high"`
	Low              decimal.Decimal `json:"low"`
	Volume           int64           `json:"volume"`
	PreviousClose    decimal.Decimal `json:"previous_close"`
	Change           decimal.Decimal `json:"change"`
	ChangePercent    decimal.Decimal `json:"change_percent"`
	LastTradeTime    time.Time       `json:"last_updated"`
}

// QuotesResponse represents multiple quotes response
type QuotesResponse struct {
	Quotes map[string]*QuoteResponse `json:"quotes"`
}

// HistoricalPriceResponse represents a single historical price point
type HistoricalPriceResponse struct {
	Date   time.Time       `json:"date"`
	Open   decimal.Decimal `json:"open"`
	High   decimal.Decimal `json:"high"`
	Low    decimal.Decimal `json:"low"`
	Close  decimal.Decimal `json:"close"`
	Volume int64           `json:"volume"`
}

// HistoricalPricesResponse represents historical prices response
type HistoricalPricesResponse struct {
	Symbol string                         `json:"symbol"`
	Prices []*HistoricalPriceResponse `json:"prices"`
}

// ExchangeRateResponse represents exchange rate response
type ExchangeRateResponse struct {
	From string          `json:"from"`
	To   string          `json:"to"`
	Rate decimal.Decimal `json:"rate"`
	Time time.Time       `json:"time"`
}

// ToQuoteResponse converts service Quote to DTO
func ToQuoteResponse(quote *services.Quote) *QuoteResponse {
	if quote == nil {
		return nil
	}

	return &QuoteResponse{
		Symbol:           quote.Symbol,
		Price:            quote.Price,
		Open:             quote.Open,
		High:             quote.High,
		Low:              quote.Low,
		Volume:           quote.Volume,
		PreviousClose:    quote.PreviousClose,
		Change:           quote.Change,
		ChangePercent:    quote.ChangePercent,
		LastTradeTime:    quote.LastUpdated,
	}
}

// ToQuotesResponse converts service quotes map to DTO
func ToQuotesResponse(quotes map[string]*services.Quote) *QuotesResponse {
	response := &QuotesResponse{
		Quotes: make(map[string]*QuoteResponse),
	}

	for symbol, quote := range quotes {
		response.Quotes[symbol] = ToQuoteResponse(quote)
	}

	return response
}

// ToHistoricalPriceResponse converts service HistoricalPrice to DTO
func ToHistoricalPriceResponse(price *services.HistoricalPrice) *HistoricalPriceResponse {
	if price == nil {
		return nil
	}

	return &HistoricalPriceResponse{
		Date:   price.Date,
		Open:   price.Open,
		High:   price.High,
		Low:    price.Low,
		Close:  price.Close,
		Volume: price.Volume,
	}
}

// ToHistoricalPricesResponse converts service historical prices to DTO
func ToHistoricalPricesResponse(prices []*services.HistoricalPrice) *HistoricalPricesResponse {
	response := &HistoricalPricesResponse{
		Prices: make([]*HistoricalPriceResponse, 0, len(prices)),
	}

	if len(prices) > 0 {
		// Assume all prices are for the same symbol (get from first entry)
		// In the service, you might want to add symbol to the response
		response.Symbol = "" // Will be set from context
	}

	for i := range prices {
		response.Prices = append(response.Prices, ToHistoricalPriceResponse(prices[i]))
	}

	return response
}

// ToExchangeRateResponse creates exchange rate response
func ToExchangeRateResponse(from, to string, rate decimal.Decimal) *ExchangeRateResponse {
	return &ExchangeRateResponse{
		From: from,
		To:   to,
		Rate: rate,
		Time: time.Now(),
	}
}
