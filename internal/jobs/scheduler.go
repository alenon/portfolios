package jobs

import (
	"context"
	"log"
	"sync"
	"time"
)

// Job represents a scheduled job
type Job interface {
	Name() string
	Run(ctx context.Context) error
	Schedule() string // Cron-like schedule (e.g., "0 18 * * *" for 6 PM daily)
}

// Scheduler manages and runs scheduled jobs
type Scheduler struct {
	jobs   []Job
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex
}

// NewScheduler creates a new job scheduler
func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		jobs:   make([]Job, 0),
		ctx:    ctx,
		cancel: cancel,
	}
}

// AddJob adds a job to the scheduler
func (s *Scheduler) AddJob(job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = append(s.jobs, job)
	log.Printf("Added job: %s with schedule: %s", job.Name(), job.Schedule())
}

// Start begins running all scheduled jobs
func (s *Scheduler) Start() {
	s.mu.RLock()
	jobs := make([]Job, len(s.jobs))
	copy(jobs, s.jobs)
	s.mu.RUnlock()

	for _, job := range jobs {
		s.wg.Add(1)
		go s.runJob(job)
	}

	log.Printf("Scheduler started with %d jobs", len(jobs))
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	s.cancel()
	s.wg.Wait()
	log.Println("Scheduler stopped")
}

// runJob runs a single job on its schedule
func (s *Scheduler) runJob(job Job) {
	defer s.wg.Done()

	// Parse schedule to determine interval
	interval := s.parseSchedule(job.Schedule())
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Job %s scheduled to run every %v", job.Name(), interval)

	// Run immediately on start (optional - can be disabled if needed)
	// s.executeJob(job)

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("Job %s stopped", job.Name())
			return
		case <-ticker.C:
			s.executeJob(job)
		}
	}
}

// executeJob executes a single job run
func (s *Scheduler) executeJob(job Job) {
	log.Printf("Running job: %s", job.Name())
	startTime := time.Now()

	if err := job.Run(s.ctx); err != nil {
		log.Printf("Job %s failed: %v", job.Name(), err)
	} else {
		log.Printf("Job %s completed successfully in %v", job.Name(), time.Since(startTime))
	}
}

// parseSchedule converts a cron-like schedule string to a duration
// For simplicity, we support a few common patterns:
// "@daily" or "0 0 * * *" -> 24 hours
// "@hourly" or "0 * * * *" -> 1 hour
// "@every Xm" -> X minutes
// "@every Xh" -> X hours
// Default: 24 hours
func (s *Scheduler) parseSchedule(schedule string) time.Duration {
	switch schedule {
	case "@daily", "0 0 * * *":
		return 24 * time.Hour
	case "@hourly", "0 * * * *":
		return time.Hour
	case "@every 30m":
		return 30 * time.Minute
	case "@every 1h":
		return time.Hour
	case "@every 6h":
		return 6 * time.Hour
	case "@every 12h":
		return 12 * time.Hour
	default:
		// Default to daily
		return 24 * time.Hour
	}
}

// RunOnce runs a job immediately (useful for testing or manual triggers)
func (s *Scheduler) RunOnce(jobName string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, job := range s.jobs {
		if job.Name() == jobName {
			return job.Run(s.ctx)
		}
	}

	return nil
}
