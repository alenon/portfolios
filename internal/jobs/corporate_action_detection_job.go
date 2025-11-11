package jobs

import (
	"context"

	"github.com/lenon/portfolios/internal/services"
)

// CorporateActionDetectionJob is a background job that detects corporate actions
type CorporateActionDetectionJob struct {
	monitor *services.CorporateActionMonitor
}

// NewCorporateActionDetectionJob creates a new corporate action detection job
func NewCorporateActionDetectionJob(monitor *services.CorporateActionMonitor) *CorporateActionDetectionJob {
	return &CorporateActionDetectionJob{
		monitor: monitor,
	}
}

// Name returns the job name
func (j *CorporateActionDetectionJob) Name() string {
	return "CorporateActionDetection"
}

// Schedule returns the job schedule
// Runs daily to check for new corporate actions
func (j *CorporateActionDetectionJob) Schedule() string {
	return "@daily" // Run once per day
}

// Run executes the job
func (j *CorporateActionDetectionJob) Run(ctx context.Context) error {
	return j.monitor.DetectAndSuggestActions(ctx)
}
