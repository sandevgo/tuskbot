//go:build integration
// +build integration

package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/stretchr/testify/assert"
)

// TestScheduler_Integration verifies that the scheduler correctly handles
// tasks over a short wall-clock period.
func TestScheduler_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name     string
		interval time.Duration
		timeout  time.Duration
		minRuns  int32
		maxRuns  int32
	}{
		{
			name:     "1 second interval over 5 seconds",
			interval: 1 * time.Second,
			timeout:  5 * time.Second,
			minRuns:  4, // Execution at 0, 1, 2, 3, 4 -> ~5 times, but allow for timing variance
			maxRuns:  6,
		},
		{
			name:     "500ms interval over 2 seconds",
			interval: 500 * time.Millisecond,
			timeout:  2 * time.Second,
			minRuns:  3, // 0, 0.5, 1.0, 1.5 -> 4 times, allow variance
			maxRuns:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sch := NewScheduler()
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			var counter int32
			job := func(ctx context.Context) error {
				atomic.AddInt32(&counter, 1)
				return nil
			}

			sch.AddTask(&core.Task{
				Name:    "counter",
				Trigger: NewIntervalTrigger(tt.interval),
				Job:     job,
			})

			go func() {
				_ = sch.Start(ctx)
			}()

			<-ctx.Done()

			actual := atomic.LoadInt32(&counter)
			assert.GreaterOrEqual(t, actual, tt.minRuns, "Task should run at least %d times", tt.minRuns)
			assert.LessOrEqual(t, actual, tt.maxRuns, "Task should run at most %d times", tt.maxRuns)
		})
	}
}
