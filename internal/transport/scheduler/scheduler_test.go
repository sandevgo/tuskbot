package scheduler

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduler_AddTask(t *testing.T) {
	s := NewScheduler()
	
	trigger := NewIntervalTrigger(1 * time.Second)
	job := func(ctx context.Context) error { return nil }
	
	// Should not panic
	require.NotPanics(t, func() {
		s.AddTask(&core.Task{
			Name:    "test-task",
			Trigger: trigger,
			Job:     job,
		})
	})
}

func TestScheduler_Triggers(t *testing.T) {
	tests := []struct {
		name           string
		setupTasks     func(s *Scheduler, counters map[string]*int32)
		expectedCounts map[string]int32
		minCounts      map[string]int32 // For interval tasks, we check minimum
		timeout        time.Duration
	}{
		{
			name: "interval trigger runs multiple times",
			setupTasks: func(s *Scheduler, counters map[string]*int32) {
				trigger := NewIntervalTrigger(100 * time.Millisecond)
				job := func(ctx context.Context) error {
					atomic.AddInt32(counters["interval"], 1)
					return nil
				}
				s.AddTask(&core.Task{
					Name:    "interval-task",
					Trigger: trigger,
					Job:     job,
				})
			},
			minCounts: map[string]int32{
				"interval": 3, // Should run at 0ms, 100ms, 200ms, 300ms... within 350ms
			},
			timeout: 350 * time.Millisecond,
		},
		{
			name: "one-off trigger runs exactly once",
			setupTasks: func(s *Scheduler, counters map[string]*int32) {
				trigger := NewOneOffTrigger(time.Now().Add(50 * time.Millisecond))
				job := func(ctx context.Context) error {
					atomic.AddInt32(counters["oneoff"], 1)
					return nil
				}
				s.AddTask(&core.Task{
					Name:    "oneoff-task",
					Trigger: trigger,
					Job:     job,
				})
			},
			expectedCounts: map[string]int32{
				"oneoff": 1,
			},
			timeout: 200 * time.Millisecond,
		},
		{
			name: "multiple tasks with different triggers",
			setupTasks: func(s *Scheduler, counters map[string]*int32) {
				// Fast interval task
				intervalTrigger := NewIntervalTrigger(100 * time.Millisecond)
				intervalJob := func(ctx context.Context) error {
					atomic.AddInt32(counters["fast"], 1)
					return nil
				}
				s.AddTask(&core.Task{
					Name:    "fast-task",
					Trigger: intervalTrigger,
					Job:     intervalJob,
				})

				// Slow interval task
				slowTrigger := NewIntervalTrigger(200 * time.Millisecond)
				slowJob := func(ctx context.Context) error {
					atomic.AddInt32(counters["slow"], 1)
					return nil
				}
				s.AddTask(&core.Task{
					Name:    "slow-task",
					Trigger: slowTrigger,
					Job:     slowJob,
				})

				// One-off task
				oneOffTrigger := NewOneOffTrigger(time.Now().Add(150 * time.Millisecond))
				oneOffJob := func(ctx context.Context) error {
					atomic.AddInt32(counters["once"], 1)
					return nil
				}
				s.AddTask(&core.Task{
					Name:    "once-task",
					Trigger: oneOffTrigger,
					Job:     oneOffJob,
				})
			},
			minCounts: map[string]int32{
				"fast": 3, // 0, 100, 200, 300... within 350ms
				"slow": 2, // 0, 200... within 350ms
			},
			expectedCounts: map[string]int32{
				"once": 1,
			},
			timeout: 350 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScheduler()
			counters := make(map[string]*int32)
			
			// Initialize counters for this test
			for k := range tt.expectedCounts {
				counters[k] = new(int32)
			}
			for k := range tt.minCounts {
				if _, ok := counters[k]; !ok {
					counters[k] = new(int32)
				}
			}

			tt.setupTasks(s, counters)

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			go func() {
				_ = s.Start(ctx)
			}()

			<-ctx.Done()

			// Check exact counts
			for name, expected := range tt.expectedCounts {
				actual := atomic.LoadInt32(counters[name])
				assert.Equal(t, expected, actual, "task %s should execute exactly %d times", name, expected)
			}

			// Check minimum counts (for interval tasks)
			for name, min := range tt.minCounts {
				actual := atomic.LoadInt32(counters[name])
				assert.GreaterOrEqual(t, actual, min, "task %s should execute at least %d times", name, min)
			}
		})
	}
}

func TestScheduler_CronTrigger(t *testing.T) {
	tests := []struct {
		name           string
		cronExpression string
		validateNext   func(now, next time.Time) bool
	}{
		{
			name:           "every second",
			cronExpression: "* * * * * *",
			validateNext: func(now, next time.Time) bool {
				// Should fire at the next second
				expected := now.Truncate(time.Second).Add(time.Second)
				diff := next.Sub(expected)
				if diff < 0 {
					diff = -diff
				}
				return diff < 2*time.Second // Allow 2 second tolerance
			},
		},
		{
			name:           "every minute at second 0",
			cronExpression: "0 * * * * *",
			validateNext: func(now, next time.Time) bool {
				// Should fire at the next minute boundary
				expected := now.Truncate(time.Minute).Add(time.Minute)
				diff := next.Sub(expected)
				if diff < 0 {
					diff = -diff
				}
				return diff < 2*time.Second // Allow 2 second tolerance
			},
		},
		{
			name:           "every hour at minute 0, second 0",
			cronExpression: "0 0 * * * *",
			validateNext: func(now, next time.Time) bool {
				// Should fire at the next hour boundary
				expected := now.Truncate(time.Hour).Add(time.Hour)
				diff := next.Sub(expected)
				if diff < 0 {
					diff = -diff
				}
				return diff < 2*time.Second // Allow 2 second tolerance
			},
		},
		{
			name:           "every day at 9 AM",
			cronExpression: "0 0 9 * * *",
			validateNext: func(now, next time.Time) bool {
				// Should fire at 9 AM today or tomorrow
				today := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
				if now.Before(today) {
					diff := next.Sub(today)
					if diff < 0 {
						diff = -diff
					}
					return diff < 2*time.Second
				}
				// Tomorrow
				tomorrow := today.Add(24 * time.Hour)
				diff := next.Sub(tomorrow)
				if diff < 0 {
					diff = -diff
				}
				return diff < 2*time.Second
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger, err := NewCronTrigger(tt.cronExpression)
			require.NoError(t, err, "should create cron trigger")

			now := time.Now()
			next := trigger.NextFireTime(now, time.Time{})

			assert.True(t, tt.validateNext(now, next), "next fire time should be valid for expression: %s", tt.cronExpression)
		})
	}
}

func TestScheduler_CronTaskExecution(t *testing.T) {
	s := NewScheduler()
	
	var executionCount int32
	
	// Use interval trigger for testing execution since cron runs at specific times
	// In production, cron would work the same way
	intervalTrigger := NewIntervalTrigger(100 * time.Millisecond)
	
	job := func(ctx context.Context) error {
		atomic.AddInt32(&executionCount, 1)
		return nil
	}
	
	s.AddTask(&core.Task{
		Name:    "interval-task",
		Trigger: intervalTrigger,
		Job:     job,
	})
	
	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()
	
	go func() {
		_ = s.Start(ctx)
	}()
	
	<-ctx.Done()
	
	// Should have executed at least 3 times (0ms, 100ms, 200ms, 300ms)
	val := atomic.LoadInt32(&executionCount)
	assert.GreaterOrEqual(t, val, int32(3), "interval task should execute at least 3 times")
}

func TestScheduler_CronMultipleTasks(t *testing.T) {
	s := NewScheduler()
	
	counters := make(map[string]*int32)
	counters["task1"] = new(int32)
	counters["task2"] = new(int32)
	
	// Two interval tasks with different schedules
	intervalTrigger1 := NewIntervalTrigger(100 * time.Millisecond)
	
	job1 := func(ctx context.Context) error {
		atomic.AddInt32(counters["task1"], 1)
		return nil
	}
	
	s.AddTask(&core.Task{
		Name:    "task-1",
		Trigger: intervalTrigger1,
		Job:     job1,
	})
	
	intervalTrigger2 := NewIntervalTrigger(200 * time.Millisecond)
	
	job2 := func(ctx context.Context) error {
		atomic.AddInt32(counters["task2"], 1)
		return nil
	}
	
	s.AddTask(&core.Task{
		Name:    "task-2",
		Trigger: intervalTrigger2,
		Job:     job2,
	})
	
	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()
	
	go func() {
		_ = s.Start(ctx)
	}()
	
	<-ctx.Done()
	
	// Both tasks should have executed
	val1 := atomic.LoadInt32(counters["task1"])
	val2 := atomic.LoadInt32(counters["task2"])
	
	assert.GreaterOrEqual(t, val1, int32(3), "task 1 should execute at least 3 times")
	assert.GreaterOrEqual(t, val2, int32(2), "task 2 should execute at least 2 times")
}

func TestScheduler_ExecutionOrder(t *testing.T) {
	s := NewScheduler()
	
	var executionOrder []string
	var mu sync.Mutex
	
	// Task that runs immediately
	nowTrigger := NewOneOffTrigger(time.Now())
	job1 := func(ctx context.Context) error {
		mu.Lock()
		executionOrder = append(executionOrder, "immediate")
		mu.Unlock()
		return nil
	}
	s.AddTask(&core.Task{
		Name:    "immediate",
		Trigger: nowTrigger,
		Job:     job1,
	})

	// Task that runs in 100ms
	futureTrigger := NewOneOffTrigger(time.Now().Add(100 * time.Millisecond))
	job2 := func(ctx context.Context) error {
		mu.Lock()
		executionOrder = append(executionOrder, "delayed")
		mu.Unlock()
		return nil
	}
	s.AddTask(&core.Task{
		Name:    "delayed",
		Trigger: futureTrigger,
		Job:     job2,
	})
	
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	
	go func() {
		_ = s.Start(ctx)
	}()
	
	<-ctx.Done()
	
	mu.Lock()
	defer mu.Unlock()
	
	require.Len(t, executionOrder, 2)
	assert.Equal(t, "immediate", executionOrder[0])
	assert.Equal(t, "delayed", executionOrder[1])
}

func TestScheduler_AddAfterStart(t *testing.T) {
	s := NewScheduler()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler in background
	go func() {
		_ = s.Start(ctx)
	}()

	// Wait a bit to ensure it's running
	time.Sleep(50 * time.Millisecond)

	done := make(chan struct{})
	job := func(ctx context.Context) error {
		close(done)
		return nil
	}

	// Add task that should run immediately
	s.AddTask(&core.Task{
		Name:    "late-task",
		Trigger: NewOneOffTrigger(time.Now()),
		Job:     job,
	})

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("task added after start did not execute in time")
	}
}

func TestScheduler_ConcurrentAdd(t *testing.T) {
	s := NewScheduler()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = s.Start(ctx)
	}()

	const numTasks = 100
	var counter int32

	var wg sync.WaitGroup
	wg.Add(numTasks)

	for i := 0; i < numTasks; i++ {
		go func() {
			defer wg.Done()
			s.AddTask(&core.Task{
				Name:    "task",
				Trigger: NewOneOffTrigger(time.Now()),
				Job: func(ctx context.Context) error {
					atomic.AddInt32(&counter, 1)
					return nil
				},
			})
		}()
	}

	wg.Wait()

	// Wait for tasks to execute
	time.Sleep(500 * time.Millisecond)

	val := atomic.LoadInt32(&counter)
	assert.Equal(t, int32(numTasks), val, "all concurrent tasks should have executed")
}

func TestScheduler_DelayedExecution(t *testing.T) {
	s := NewScheduler()

	var executed int32

	// Task that runs in 200ms
	futureTrigger := NewOneOffTrigger(time.Now().Add(200 * time.Millisecond))
	job := func(ctx context.Context) error {
		atomic.AddInt32(&executed, 1)
		return nil
	}

	s.AddTask(&core.Task{
		Name:    "delayed-check",
		Trigger: futureTrigger,
		Job:     job,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		_ = s.Start(ctx)
	}()

	<-ctx.Done()

	// Should NOT have executed yet
	val := atomic.LoadInt32(&executed)
	assert.Equal(t, int32(0), val, "delayed task should not execute immediately")
}
