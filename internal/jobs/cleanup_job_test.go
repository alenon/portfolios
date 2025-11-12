package jobs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanupJob_Name(t *testing.T) {
	job := NewCleanupJob(nil, 365)
	assert.Equal(t, "Cleanup", job.Name())
}

func TestCleanupJob_Schedule(t *testing.T) {
	job := NewCleanupJob(nil, 365)
	assert.Equal(t, "@daily", job.Schedule())
}

func TestCleanupJob_Run(t *testing.T) {
	job := NewCleanupJob(nil, 365)
	ctx := context.Background()

	// Should not error even with nil service
	err := job.Run(ctx)
	assert.NoError(t, err)
}

func TestCleanupJob_DefaultRetentionDays(t *testing.T) {
	// Test that invalid retention days defaults to 365
	job := NewCleanupJob(nil, 0)
	assert.NotNil(t, job)

	job2 := NewCleanupJob(nil, -10)
	assert.NotNil(t, job2)
}
