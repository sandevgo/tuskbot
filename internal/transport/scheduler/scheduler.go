package scheduler

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type Scheduler struct {
	tasks  taskHeap
	mu     sync.Mutex
	signal chan struct{}
}

func NewScheduler() *Scheduler {
	sch := &Scheduler{
		tasks:  make(taskHeap, 0),
		signal: make(chan struct{}, 1),
	}
	heap.Init(&sch.tasks)
	return sch
}

func (s *Scheduler) AddTask(name string, trigger core.Trigger, job core.Job) {
	next := trigger.NextFireTime(time.Now(), time.Time{})
	if next.IsZero() {
		return
	}

	task := &core.Task{
		ID:      name,
		Name:    name,
		Trigger: trigger,
		Job:     job,
		LastRun: time.Time{},
		NextRun: next,
	}

	s.mu.Lock()
	heap.Push(&s.tasks, task)
	isHead := s.tasks.Peek() == task
	s.mu.Unlock()

	if isHead {
		select {
		case s.signal <- struct{}{}:
		default:
		}
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	logger := log.FromCtx(ctx)
	logger.Info().Msg("starting scheduler")

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("scheduler stopped by contexts")
			return ctx.Err()
		default:
		}

		now := time.Now()

		// 1. Peek: Get the earliest task
		s.mu.Lock()
		nextTask := s.tasks.Peek()
		s.mu.Unlock()

		if nextTask == nil {
			select {
			case <-s.signal:
				continue
			case <-time.After(1 * time.Second):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// 2. Wait: If it is not time yet, sleep until NextRun
		delay := nextTask.NextRun.Sub(now)
		if delay > 0 {
			select {
			case <-time.After(delay):
			case <-s.signal:
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
			now = time.Now()
		}

		// 3. Execute: Pop the task and run it
		s.mu.Lock()
		if s.tasks.Len() == 0 {
			s.mu.Unlock()
			continue
		}
		task := heap.Pop(&s.tasks).(*core.Task)
		s.mu.Unlock()

		// Run job
		go func(t *core.Task) {
			fmt.Printf("executing Task '%s'...\n", t.Name)
			_ = t.Job(ctx)
		}(task)

		// 4. Calculate next run time
		task.LastRun = now
		nextRun := task.Trigger.NextFireTime(now, task.LastRun)

		if !nextRun.IsZero() {
			task.NextRun = nextRun
			s.mu.Lock()
			heap.Push(&s.tasks, task)
			s.mu.Unlock()
			fmt.Printf("[Scheduler] Task '%s' done. Rescheduled for %s\n", task.Name, nextRun.Format(time.RFC3339))
		} else {
			fmt.Printf("[Scheduler] Task '%s' completed forever (One-Off).\n", task.Name)
		}
	}
}

func (s *Scheduler) Shutdown(ctx context.Context) error {
	return nil // TODO: Implement stop logic
}
