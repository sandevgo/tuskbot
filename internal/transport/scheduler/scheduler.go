package scheduler

import (
	"container/heap"
	"context"
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

func (s *Scheduler) AddTask(task *core.Task) {
	next := task.NextFireTime(time.Now())
	if next.IsZero() {
		return
	}
	task.NextRun = next

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

func (s *Scheduler) DelTask(task *core.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.tasks {
		if t.ID == task.ID {
			heap.Remove(&s.tasks, i)
			// Signal scheduler in case the deleted task was the next scheduled one
			select {
			case s.signal <- struct{}{}:
			default:
			}
			return
		}
	}
}

func (s *Scheduler) ListTasks() []*core.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]*core.Task, len(s.tasks))
	for i, t := range s.tasks {
		result[i] = t
	}
	return result
}

func (s *Scheduler) Start(ctx context.Context) error {
	logger := log.FromCtx(ctx)
	logger.Info().Msg("starting scheduler")

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("scheduler stopped by contexts")
			return nil
		default:
		}

		now := time.Now()

		// Get the earliest task
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
				return nil
			}
		}

		// Wait: If it is not time yet, sleep until NextRun
		delay := nextTask.NextRun.Sub(now)
		if delay > 0 {
			select {
			case <-time.After(delay):
			case <-s.signal:
				continue
			case <-ctx.Done():
				return nil
			}
			now = time.Now()
		}

		// Execute: Pop the task and run it
		s.mu.Lock()
		if s.tasks.Len() == 0 {
			s.mu.Unlock()
			continue
		}
		task := heap.Pop(&s.tasks).(*core.Task)
		s.mu.Unlock()

		// Run job
		go func(t *core.Task) {
			logger.Info().Str("name", t.Name).Msg("executing task")

			err := t.Job(ctx)
			if err != nil {
				logger.Error().Err(err).Msgf("failed to execute task '%s': %v", t.Name, err)
			}
		}(task)

		// Calculate next run time
		task.LastRun = now
		nextRun := task.Trigger.NextFireTime(now, task.LastRun)

		if !nextRun.IsZero() {
			task.NextRun = nextRun
			s.mu.Lock()
			heap.Push(&s.tasks, task)
			s.mu.Unlock()
			logger.Info().Str("name", task.Name).Str("next_run", nextRun.Format(time.DateTime)).Msg("rescheduled task")
		} else {
			logger.Info().Str("name", task.Name).Msg("task completed and removed")
		}
	}
}

func (s *Scheduler) Shutdown(ctx context.Context) error {
	return nil // TODO: Implement stop logic
}
