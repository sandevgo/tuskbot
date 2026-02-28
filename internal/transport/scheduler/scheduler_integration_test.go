//go:build integration
// +build integration

package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestScheduler_Integration verifies that the scheduler correctly handles
// tasks over a short wall-clock period.
func TestScheduler_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	sch := NewScheduler()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var counter int
	job := func(ctx context.Context) error {
		counter++
		return nil
	}

	// Run every 1 second
	sch.AddTask("counter", NewIntervalTrigger(1*time.Second), job)

	go func() {
		_ = sch.Start(ctx)
	}()

	<-ctx.Done()

	// 5 seconds -> execution at 0, 1, 2, 3, 4 -> ~5 times
	assert.GreaterOrEqual(t, counter, 4, "Task should run at least 4 times")
	assert.LessOrEqual(t, counter, 6, "Task should run at most 6 times")
}
