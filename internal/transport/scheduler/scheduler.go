package scheduler

import (
	"container/heap"
	"context"
	"log"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
)

type Scheduler struct {
	tasks taskHeap
}

func NewScheduler() *Scheduler {
	sch := &Scheduler{
		tasks: make(taskHeap, 0),
	}
	heap.Init(&sch.tasks)
	return sch
}

func (s *Scheduler) AddTask(name string, trigger core.Trigger, job core.Job) {
	// Schedule for the very first execution
	next := trigger.NextFireTime(time.Now(), time.Time{})

	if !next.IsZero() {
		task := &core.Task{
			ID:      name,
			Name:    name,
			Trigger: trigger,
			Job:     job,
			LastRun: time.Time{},
			NextRun: next,
		}
		heap.Push(&s.tasks, task)
		log.Printf("[Scheduler] Task '%s' added. Next run: %s\n", name, next.Format(time.RFC3339))
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	log.Println("[Scheduler] Started.")
	for {
		select {
		case <-ctx.Done():
			log.Println("[Scheduler] Stopped by context.")
			return ctx.Err()
		default:
		}

		now := time.Now()

		// 1. Peek: Get the earliest task
		nextTask := s.tasks.Peek()
		if nextTask == nil {
			// No tasks? Just sleep a bit (1s) and check for new ones later.
			time.Sleep(1 * time.Second)
			continue
		}

		// 2. Wait: If it is not time yet, sleep until NextRun
		delay := nextTask.NextRun.Sub(now)
		if delay > 0 {
			// Select on timer OR context cancellation
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
			now = time.Now()
		}

		// 3. Execute: Pop the task and run it
		task := heap.Pop(&s.tasks).(*core.Task) // Pop task from heap

		// Run job in separate goroutine
		log.Printf("[Scheduler] Executing Task '%s'...\n", task.Name)
		go func(t *core.Task) {
			// Execute the job with passed context or a new one? Often we want task context independent of Scheduler cancellation?
			_ = t.Job(context.Background())
		}(task)

		// 4. Reschedule: Calculate next run time
		task.LastRun = now
		nextRun := task.Trigger.NextFireTime(now, task.LastRun)

		if !nextRun.IsZero() {
			task.NextRun = nextRun
			heap.Push(&s.tasks, task)
			log.Printf("[Scheduler] Task '%s' done. Rescheduled for %s\n", task.Name, nextRun.Format(time.RFC3339))
		} else {
			log.Printf("[Scheduler] Task '%s' completed forever (One-Off).\n", task.Name)
		}
	}
}
