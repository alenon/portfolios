package jobs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSnapshotGenerationJob_Name(t *testing.T) {
	job := NewSnapshotGenerationJob()
	assert.Equal(t, "SnapshotGeneration", job.Name())
}

func TestSnapshotGenerationJob_Schedule(t *testing.T) {
	job := NewSnapshotGenerationJob()
	assert.Equal(t, "@daily", job.Schedule())
}

func TestSnapshotGenerationJob_Run(t *testing.T) {
	job := NewSnapshotGenerationJob()
	ctx := context.Background()

	// Should not error (it's currently a placeholder)
	err := job.Run(ctx)
	assert.NoError(t, err)
}
