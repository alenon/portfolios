package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewAlphaVantageProvider(t *testing.T) {
	provider := NewAlphaVantageProvider("test-api-key")
	assert.NotNil(t, provider)
	assert.Equal(t, "test-api-key", provider.apiKey)
	assert.NotNil(t, provider.httpClient)
}

func TestAlphaVantageProvider_IsAvailable(t *testing.T) {
	t.Run("available with API key", func(t *testing.T) {
		provider := NewAlphaVantageProvider("test-api-key")
		assert.True(t, provider.IsAvailable())
	})

	t.Run("not available without API key", func(t *testing.T) {
		provider := NewAlphaVantageProvider("")
		assert.False(t, provider.IsAvailable())
	})
}

func TestAlphaVantageProvider_GetQuote(t *testing.T) {
	t.Run("successful quote retrieval", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GLOBAL_QUOTE", r.URL.Query().Get("function"))
			assert.Equal(t, "AAPL", r.URL.Query().Get("symbol"))

			response := `{
				"Global Quote": {
					"01. symbol": "AAPL",
					"02. open": "150.00",
					"03. high": "155.00",
					"04. low": "149.00",
					"05. price": "152.50",
					"06. volume": "1000000",
					"07. latest trading day": "2024-01-15",
					"08. previous close": "150.00",
					"09. change": "2.50",
					"10. change percent": "1.67%"
				}
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey:     "test-api-key",
			httpClient: server.Client(),
		}
		// Override base URL by creating request manually
		originalURL := alphaVantageBaseURL
		defer func() { _ = originalURL }()

		// Use a custom provider that points to our test server
		provider = &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		ctx := context.Background()
		quote, err := provider.GetQuote(ctx, "AAPL")

		assert.NoError(t, err)
		assert.NotNil(t, quote)
		assert.Equal(t, "AAPL", quote.Symbol)
		assert.True(t, quote.Price.Equal(decimal.NewFromFloat(152.50)))
		assert.True(t, quote.Open.Equal(decimal.NewFromFloat(150.00)))
		assert.True(t, quote.High.Equal(decimal.NewFromFloat(155.00)))
		assert.True(t, quote.Low.Equal(decimal.NewFromFloat(149.00)))
		assert.True(t, quote.PreviousClose.Equal(decimal.NewFromFloat(150.00)))
		assert.True(t, quote.Change.Equal(decimal.NewFromFloat(2.50)))
		assert.True(t, quote.ChangePercent.Equal(decimal.NewFromFloat(1.67)))
		assert.Equal(t, int64(1000000), quote.Volume)
	})

	t.Run("API key not configured", func(t *testing.T) {
		provider := &AlphaVantageProvider{
			apiKey:     "",
			httpClient: &http.Client{},
		}

		ctx := context.Background()
		_, err := provider.GetQuote(ctx, "AAPL")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API key not configured")
	})

	t.Run("API error message", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `{
				"Error Message": "Invalid API call"
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		ctx := context.Background()
		_, err := provider.GetQuote(ctx, "INVALID")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
	})

	t.Run("API rate limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `{
				"Note": "Thank you for using Alpha Vantage! Our standard API call frequency is 5 calls per minute and 500 calls per day."
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		ctx := context.Background()
		_, err := provider.GetQuote(ctx, "AAPL")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit")
	})

	t.Run("empty symbol data", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `{
				"Global Quote": {}
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		ctx := context.Background()
		_, err := provider.GetQuote(ctx, "UNKNOWN")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no data found")
	})

	t.Run("non-200 status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		ctx := context.Background()
		_, err := provider.GetQuote(ctx, "AAPL")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status 500")
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		ctx := context.Background()
		_, err := provider.GetQuote(ctx, "AAPL")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse response")
	})
}

func TestAlphaVantageProvider_GetQuotes(t *testing.T) {
	t.Run("successful batch retrieval", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			symbol := r.URL.Query().Get("symbol")

			response := fmt.Sprintf(`{
				"Global Quote": {
					"01. symbol": "%s",
					"05. price": "100.00"
				}
			}`, symbol)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		// Use a timeout that allows for the 12-second delay between requests
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Test with just 2 symbols (requires ~12 seconds with delay)
		quotes, err := provider.GetQuotes(ctx, []string{"AAPL", "GOOGL"})

		assert.NoError(t, err)
		assert.Len(t, quotes, 2)
		assert.NotNil(t, quotes["AAPL"])
		assert.NotNil(t, quotes["GOOGL"])
		assert.Equal(t, 2, callCount)
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(100 * time.Millisecond)
			response := `{
				"Global Quote": {
					"01. symbol": "AAPL",
					"05. price": "100.00"
				}
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		// Create a context that cancels immediately after first request
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		quotes, err := provider.GetQuotes(ctx, []string{"AAPL", "GOOGL", "MSFT"})

		// Should get at least one quote before timeout
		assert.True(t, len(quotes) >= 1 || err != nil)
	})

	t.Run("skip failed symbols", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			symbol := r.URL.Query().Get("symbol")

			if symbol == "INVALID" {
				response := `{"Error Message": "Invalid symbol"}`
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))
				return
			}

			response := fmt.Sprintf(`{
				"Global Quote": {
					"01. symbol": "%s",
					"05. price": "100.00"
				}
			}`, symbol)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		ctx := context.Background()
		quotes, err := provider.GetQuotes(ctx, []string{"AAPL", "INVALID"})

		assert.NoError(t, err)
		assert.Len(t, quotes, 1) // Only AAPL should succeed
		assert.NotNil(t, quotes["AAPL"])
	})
}

func TestAlphaVantageProvider_GetHistoricalPrices(t *testing.T) {
	t.Run("successful historical data retrieval", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "TIME_SERIES_DAILY_ADJUSTED", r.URL.Query().Get("function"))
			assert.Equal(t, "AAPL", r.URL.Query().Get("symbol"))

			response := `{
				"Time Series (Daily)": {
					"2024-01-15": {
						"1. open": "150.00",
						"2. high": "155.00",
						"3. low": "149.00",
						"4. close": "152.50",
						"5. adjusted close": "152.50",
						"6. volume": "1000000"
					},
					"2024-01-14": {
						"1. open": "148.00",
						"2. high": "151.00",
						"3. low": "147.00",
						"4. close": "150.00",
						"5. adjusted close": "150.00",
						"6. volume": "900000"
					}
				}
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		ctx := context.Background()
		startDate := time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

		prices, err := provider.GetHistoricalPrices(ctx, "AAPL", startDate, endDate)

		assert.NoError(t, err)
		assert.Len(t, prices, 2)
	})

	t.Run("API key not configured", func(t *testing.T) {
		provider := &AlphaVantageProvider{
			apiKey:     "",
			httpClient: &http.Client{},
		}

		ctx := context.Background()
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

		_, err := provider.GetHistoricalPrices(ctx, "AAPL", startDate, endDate)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API key not configured")
	})

	t.Run("API error message", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `{
				"Error Message": "Invalid API call"
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		ctx := context.Background()
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

		_, err := provider.GetHistoricalPrices(ctx, "INVALID", startDate, endDate)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
	})

	t.Run("date filtering", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `{
				"Time Series (Daily)": {
					"2024-01-15": {
						"1. open": "150.00",
						"4. close": "152.50",
						"5. adjusted close": "152.50"
					},
					"2024-01-10": {
						"1. open": "145.00",
						"4. close": "147.00",
						"5. adjusted close": "147.00"
					},
					"2024-01-05": {
						"1. open": "140.00",
						"4. close": "142.00",
						"5. adjusted close": "142.00"
					}
				}
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		ctx := context.Background()
		// Only request data from Jan 10-15
		startDate := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

		prices, err := provider.GetHistoricalPrices(ctx, "AAPL", startDate, endDate)

		assert.NoError(t, err)
		// Should only get 2 prices (Jan 15 and Jan 10), not Jan 5
		assert.Len(t, prices, 2)
	})

	t.Run("non-200 status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		ctx := context.Background()
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

		_, err := provider.GetHistoricalPrices(ctx, "AAPL", startDate, endDate)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status 500")
	})
}

func TestAlphaVantageProvider_GetExchangeRate(t *testing.T) {
	t.Run("successful exchange rate retrieval", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "CURRENCY_EXCHANGE_RATE", r.URL.Query().Get("function"))
			assert.Equal(t, "USD", r.URL.Query().Get("from_currency"))
			assert.Equal(t, "EUR", r.URL.Query().Get("to_currency"))

			response := `{
				"Realtime Currency Exchange Rate": {
					"5. Exchange Rate": "0.85"
				}
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		ctx := context.Background()
		rate, err := provider.GetExchangeRate(ctx, "USD", "EUR")

		assert.NoError(t, err)
		assert.True(t, rate.Equal(decimal.NewFromFloat(0.85)))
	})

	t.Run("same currency returns 1", func(t *testing.T) {
		provider := &AlphaVantageProvider{
			apiKey:     "test-api-key",
			httpClient: &http.Client{},
		}

		ctx := context.Background()
		rate, err := provider.GetExchangeRate(ctx, "USD", "USD")

		assert.NoError(t, err)
		assert.True(t, rate.Equal(decimal.NewFromInt(1)))
	})

	t.Run("API key not configured", func(t *testing.T) {
		provider := &AlphaVantageProvider{
			apiKey:     "",
			httpClient: &http.Client{},
		}

		ctx := context.Background()
		_, err := provider.GetExchangeRate(ctx, "USD", "EUR")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API key not configured")
	})

	t.Run("API error message", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `{
				"Error Message": "Invalid currency"
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		ctx := context.Background()
		_, err := provider.GetExchangeRate(ctx, "INVALID", "EUR")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
	})

	t.Run("invalid exchange rate format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `{
				"Realtime Currency Exchange Rate": {
					"5. Exchange Rate": "invalid"
				}
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		provider := &AlphaVantageProvider{
			apiKey: "test-api-key",
			httpClient: &http.Client{
				Transport: &mockTransport{server: server},
			},
		}

		ctx := context.Background()
		_, err := provider.GetExchangeRate(ctx, "USD", "EUR")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse exchange rate")
	})
}

// mockTransport is a custom RoundTripper that redirects all requests to a test server
type mockTransport struct {
	server *httptest.Server
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Redirect the request to our test server
	req.URL.Scheme = "http"
	req.URL.Host = t.server.URL[7:] // Remove "http://" prefix
	return t.server.Client().Do(req)
}
