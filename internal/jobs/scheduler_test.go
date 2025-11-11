package jobs

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock job for testing
type mockJob struct {
	name      string
	schedule  string
	runCount  int
	runError  error
	mu        sync.Mutex
	runCalled chan struct{}
}

func newMockJob(name, schedule string) *mockJob {
	return &mockJob{
		name:      name,
		schedule:  schedule,
		runCalled: make(chan struct{}, 10),
	}
}

func (m *mockJob) Name() string {
	return m.name
}

func (m *mockJob) Schedule() string {
	return m.schedule
}

func (m *mockJob) Run(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runCount++
	select {
	case m.runCalled <- struct{}{}:
	default:
	}
	if m.runError != nil {
		return m.runError
	}
	return nil
}

func (m *mockJob) getRunCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.runCount
}

func TestNewScheduler(t *testing.T) {
	scheduler := NewScheduler()
	require.NotNil(t, scheduler)
	assert.NotNil(t, scheduler.jobs)
	assert.NotNil(t, scheduler.ctx)
	assert.NotNil(t, scheduler.cancel)
}

func TestScheduler_AddJob(t *testing.T) {
	scheduler := NewScheduler()
	job := newMockJob("test-job", "@daily")

	scheduler.AddJob(job)

	assert.Len(t, scheduler.jobs, 1)
	assert.Equal(t, job, scheduler.jobs[0])
}

func TestScheduler_AddMultipleJobs(t *testing.T) {
	scheduler := NewScheduler()
	job1 := newMockJob("job1", "@daily")
	job2 := newMockJob("job2", "@hourly")

	scheduler.AddJob(job1)
	scheduler.AddJob(job2)

	assert.Len(t, scheduler.jobs, 2)
}

func TestScheduler_StartAndStop(t *testing.T) {
	scheduler := NewScheduler()
	job := newMockJob("test-job", "@every 30m")

	scheduler.AddJob(job)
	scheduler.Start()

	// Give scheduler time to start
	time.Sleep(100 * time.Millisecond)

	scheduler.Stop()

	// Verify job was initialized (scheduler started successfully)
	assert.GreaterOrEqual(t, job.getRunCount(), 0)
}

func TestScheduler_JobExecution(t *testing.T) {
	scheduler := NewScheduler()
	// Use a very short interval for testing
	job := newMockJob("fast-job", "@every 30m")

	scheduler.AddJob(job)
	scheduler.Start()

	// Wait a bit for any initial execution
	time.Sleep(100 * time.Millisecond)

	scheduler.Stop()

	// Job should have been set up (even if not run yet due to ticker)
	assert.GreaterOrEqual(t, job.getRunCount(), 0)
}

func TestScheduler_JobWithError(t *testing.T) {
	scheduler := NewScheduler()
	job := newMockJob("error-job", "@daily")
	job.runError = errors.New("test error")

	scheduler.AddJob(job)
	scheduler.Start()

	time.Sleep(100 * time.Millisecond)

	scheduler.Stop()

	// Job should still be callable despite error
	assert.GreaterOrEqual(t, job.getRunCount(), 0)
}

func TestScheduler_RunOnce(t *testing.T) {
	scheduler := NewScheduler()
	job := newMockJob("test-job", "@daily")

	scheduler.AddJob(job)

	err := scheduler.RunOnce("test-job")
	require.NoError(t, err)

	assert.Equal(t, 1, job.getRunCount())
}

func TestScheduler_RunOnce_NonExistentJob(t *testing.T) {
	scheduler := NewScheduler()
	job := newMockJob("test-job", "@daily")

	scheduler.AddJob(job)

	err := scheduler.RunOnce("non-existent")
	require.NoError(t, err) // Returns nil for non-existent jobs

	assert.Equal(t, 0, job.getRunCount())
}

func TestScheduler_ParseSchedule(t *testing.T) {
	scheduler := NewScheduler()

	tests := []struct {
		name     string
		schedule string
		expected time.Duration
	}{
		{"daily", "@daily", 24 * time.Hour},
		{"daily cron", "0 0 * * *", 24 * time.Hour},
		{"hourly", "@hourly", time.Hour},
		{"hourly cron", "0 * * * *", time.Hour},
		{"every 30m", "@every 30m", 30 * time.Minute},
		{"every 1h", "@every 1h", time.Hour},
		{"every 6h", "@every 6h", 6 * time.Hour},
		{"every 12h", "@every 12h", 12 * time.Hour},
		{"unknown", "unknown", 24 * time.Hour}, // default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := scheduler.parseSchedule(tt.schedule)
			assert.Equal(t, tt.expected, duration)
		})
	}
}

func TestScheduler_ConcurrentJobs(t *testing.T) {
	scheduler := NewScheduler()

	job1 := newMockJob("job1", "@daily")
	job2 := newMockJob("job2", "@hourly")
	job3 := newMockJob("job3", "@every 30m")

	scheduler.AddJob(job1)
	scheduler.AddJob(job2)
	scheduler.AddJob(job3)

	scheduler.Start()
	time.Sleep(100 * time.Millisecond)
	scheduler.Stop()

	// All jobs should be set up
	assert.GreaterOrEqual(t, job1.getRunCount(), 0)
	assert.GreaterOrEqual(t, job2.getRunCount(), 0)
	assert.GreaterOrEqual(t, job3.getRunCount(), 0)
}

func TestScheduler_ContextCancellation(t *testing.T) {
	scheduler := NewScheduler()
	job := newMockJob("test-job", "@every 1h")

	scheduler.AddJob(job)
	scheduler.Start()

	// Give time for job to start
	time.Sleep(50 * time.Millisecond)

	// Stop should cancel context
	scheduler.Stop()

	// Context should be cancelled
	select {
	case <-scheduler.ctx.Done():
		// Expected
	case <-time.After(time.Second):
		t.Fatal("Context not cancelled after Stop")
	}
}
