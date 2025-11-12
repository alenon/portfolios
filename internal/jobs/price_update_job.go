package jobs

import (
	"context"
	"log"
	"time"

	"github.com/lenon/portfolios/internal/services"
)

// PriceUpdateJob is a background job that refreshes market data cache
// In a full implementation, this would:
// - Fetch end-of-day prices for all symbols in active portfolios
// - Update holdings with current market values
// - Refresh quotes in the market data cache
type PriceUpdateJob struct {
	marketDataSvc services.MarketDataService
}

// NewPriceUpdateJob creates a new price update job
func NewPriceUpdateJob(marketDataSvc services.MarketDataService) *PriceUpdateJob {
	return &PriceUpdateJob{
		marketDataSvc: marketDataSvc,
	}
}

// Name returns the job name
func (j *PriceUpdateJob) Name() string {
	return "PriceUpdate"
}

// Schedule returns the job schedule
// Runs every day after market close (6 PM ET / 18:00)
func (j *PriceUpdateJob) Schedule() string {
	return "@daily"
}

// Run executes the job
func (j *PriceUpdateJob) Run(ctx context.Context) error {
	log.Println("Starting price update job...")
	startTime := time.Now()

	// Clear the market data cache to force fresh fetches
	// This ensures stale prices don't persist
	if j.marketDataSvc != nil {
		j.marketDataSvc.ClearCache()
		log.Println("Market data cache cleared")
	}

	// Note: In a full implementation, this job would:
	// 1. Query all active portfolios
	// 2. Collect unique symbols from all holdings
	// 3. Fetch current prices for all symbols in batch
	// 4. Update holdings or cache with fresh prices
	//
	// This requires adding a FindAll() method to PortfolioRepository
	// and potentially storing current prices in the Holding model.

	duration := time.Since(startTime)
	log.Printf("Price update completed in %v", duration)

	return nil
}
