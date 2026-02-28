package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScheduler_Run(t *testing.T) {
	s := NewScheduler()

	// Task 1: Runs every 1 second
	var count1 int
	trigger1 := NewIntervalTrigger(1 * time.Second)
	job1 := func(ctx context.Context) error {
		count1++
		return nil
	}

	// Task 2: Runs once after 2 seconds
	var count2 int
	trigger2 := NewOneOffTrigger(time.Now().Add(2 * time.Second))
	job2 := func(ctx context.Context) error {
		count2++
		return nil
	}

	s.AddTask("Interval-1s", trigger1, job1)
	s.AddTask("OneOff-2s", trigger2, job2)

	// In a real app we'd block. Here we run for 3.5s and check results.
	ctx, cancel := context.WithTimeout(context.Background(), 3500*time.Millisecond)
	defer cancel()

	go func() {
		_ = s.Start(ctx)
	}()

	<-ctx.Done()

	// Assertions
	assert.GreaterOrEqual(t, count1, 3, "Interval task should run multiple times")
	assert.Equal(t, 1, count2, "OneOff task should run exactly once")
}
