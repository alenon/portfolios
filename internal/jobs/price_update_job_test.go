package jobs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriceUpdateJob_Name(t *testing.T) {
	job := NewPriceUpdateJob(nil)
	assert.Equal(t, "PriceUpdate", job.Name())
}

func TestPriceUpdateJob_Schedule(t *testing.T) {
	job := NewPriceUpdateJob(nil)
	assert.Equal(t, "@daily", job.Schedule())
}

func TestPriceUpdateJob_Run(t *testing.T) {
	job := NewPriceUpdateJob(nil)
	ctx := context.Background()

	// Should not error even with nil service
	err := job.Run(ctx)
	assert.NoError(t, err)
}
