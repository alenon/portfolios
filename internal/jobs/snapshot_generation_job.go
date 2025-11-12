package jobs

import (
	"context"
	"log"
	"time"
)

// SnapshotGenerationJob is a background job that generates daily performance snapshots
// In a full implementation, this would:
// - Query all active portfolios
// - Generate daily performance snapshots for each
// - Calculate daily returns and metrics
type SnapshotGenerationJob struct{}

// NewSnapshotGenerationJob creates a new snapshot generation job
func NewSnapshotGenerationJob() *SnapshotGenerationJob {
	return &SnapshotGenerationJob{}
}

// Name returns the job name
func (j *SnapshotGenerationJob) Name() string {
	return "SnapshotGeneration"
}

// Schedule returns the job schedule
// Runs daily after market close and price updates (7 PM ET / 19:00)
func (j *SnapshotGenerationJob) Schedule() string {
	return "@daily"
}

// Run executes the job
func (j *SnapshotGenerationJob) Run(ctx context.Context) error {
	log.Println("Starting snapshot generation job...")
	startTime := time.Now()

	// Note: In a full implementation, this job would:
	// 1. Query all active portfolios via PortfolioRepository.FindAll()
	// 2. For each portfolio:
	//    a. Get current holdings
	//    b. Fetch current market prices
	//    c. Calculate total value, returns, etc.
	//    d. Create a performance snapshot via PerformanceSnapshotService
	// 3. Handle errors gracefully, logging failed portfolios
	//
	// This requires:
	// - Adding FindAll() method to PortfolioRepository
	// - Potentially simplifying CreateSnapshot to auto-fetch prices
	// - Or passing a userID context (perhaps a system user for batch jobs)

	duration := time.Since(startTime)
	log.Printf("Snapshot generation completed in %v", duration)

	return nil
}
