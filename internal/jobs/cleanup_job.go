package jobs

import (
	"context"
	"log"
	"time"

	"github.com/lenon/portfolios/internal/services"
)

// CleanupJob is a background job that cleans up stale data
// Currently handles:
// - Clearing market data cache
// Future enhancements:
// - Clean up expired refresh tokens
// - Clean up expired password reset tokens
// - Archive old performance snapshots
type CleanupJob struct {
	marketDataSvc services.MarketDataService
	retentionDays int
}

// NewCleanupJob creates a new cleanup job
func NewCleanupJob(
	marketDataSvc services.MarketDataService,
	retentionDays int,
) *CleanupJob {
	if retentionDays <= 0 {
		retentionDays = 365 // Default to 1 year retention
	}

	return &CleanupJob{
		marketDataSvc: marketDataSvc,
		retentionDays: retentionDays,
	}
}

// Name returns the job name
func (j *CleanupJob) Name() string {
	return "Cleanup"
}

// Schedule returns the job schedule
// Runs daily at midnight
func (j *CleanupJob) Schedule() string {
	return "@daily"
}

// Run executes the job
func (j *CleanupJob) Run(ctx context.Context) error {
	log.Println("Starting cleanup job...")
	startTime := time.Now()

	// Clear market data cache
	if j.marketDataSvc != nil {
		log.Println("Clearing market data cache...")
		j.marketDataSvc.ClearCache()
		log.Println("Market data cache cleared")
	}

	// Note: Additional cleanup tasks would go here:
	//
	// 1. Clean up expired refresh tokens:
	//    - Requires RefreshTokenRepository.FindExpired() or similar
	//    - Delete all tokens where ExpiresAt < now()
	//
	// 2. Clean up expired password reset tokens:
	//    - Requires PasswordResetRepository.FindExpired() or similar
	//    - Delete all resets where ExpiresAt < now() or Used = true
	//
	// 3. Archive old performance snapshots:
	//    - Requires PerformanceSnapshotRepository.FindOlderThan(date)
	//    - Either delete or move to archive table
	//    - Keep recent snapshots (e.g., last 365 days)
	//
	// These require adding batch query methods to the repositories.

	duration := time.Since(startTime)
	log.Printf("Cleanup completed in %v", duration)

	return nil
}
