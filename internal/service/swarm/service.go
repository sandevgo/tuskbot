package swarm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/internal/service/agent"
	"github.com/sandevgo/tuskbot/internal/transport/scheduler"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type TaskInfo struct {
	ID          string
	Name        string
	NextRun     time.Time
	LastRun     time.Time
	Instruction string
}

type Service struct {
	scheduler core.Scheduler
	agent     *agent.Agent
	mu        sync.RWMutex
	tasks     map[string]*core.Task
}

func NewService(scheduler core.Scheduler, agent *agent.Agent) *Service {
	return &Service{
		scheduler: scheduler,
		agent:     agent,
		tasks:     make(map[string]*core.Task),
	}
}

func (s *Service) ScheduleTask(ctx context.Context, name, taskType, timeSpec, instruction string) error {
	sessionID := fmt.Sprintf("task-%s", name)

	var trigger core.Trigger
	var err error

	switch taskType {
	case core.TriggerTypeInterval:
		dur, err := time.ParseDuration(timeSpec)
		if err != nil {
			return fmt.Errorf("invalid interval duration: %w", err)
		}
		trigger = scheduler.NewIntervalTrigger(dur)
	case core.TriggerTypeOnce:
		// Try parsing as duration (delay) first
		if dur, err := time.ParseDuration(timeSpec); err == nil {
			trigger = &scheduler.OneOffTrigger{At: time.Now().Add(dur)}
		} else {
			// Try parsing as ISO8601/RFC3339
			t, err := time.Parse(time.RFC3339, timeSpec)
			if err != nil {
				return fmt.Errorf("invalid time_spec for 'once' (expected duration or RFC3339): %w", err)
			}
			trigger = &scheduler.OneOffTrigger{At: t}
		}
	case core.TriggerTypeCron:
		trigger = &scheduler.CronTrigger{Expression: timeSpec}
	default:
		return fmt.Errorf("unknown task type: %s", taskType)
	}

	// Create the Job that will run when triggered
	job := func(ctx context.Context) error {
		logger := log.FromCtx(ctx)

		logger.Info().
			Str("session", sessionID).
			Msg("executing scheduled task")

		_, err := s.agent.Run(ctx, sessionID, instruction, func(msg core.Message) {
			if msg.Content != "" {
				// save stream state
				logger.Debug().Str("output", msg.Content).Msg("task progress")
			}
		})
		return err
	}

	// Wrap the core.Task to store metadata
	task := &core.Task{
		ID:      name,
		Name:    name,
		Trigger: trigger,
		Job:     job,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[name] = task
	s.scheduler.AddTask(name, trigger, job)

	return nil
}

func (s *Service) CancelTask(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from local tracking
	delete(s.tasks, name)

	// Note: You'll need to add a RemoveTask method to your Scheduler interface
	// or implement cancellation via context cancellation
	return nil
}

func (s *Service) ListTasks(ctx context.Context) ([]core.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var infos []core.Task
	for _, t := range s.tasks {
		infos = append(infos, core.Task{
			ID:      t.ID,
			Name:    t.Name,
			NextRun: t.NextRun,
			LastRun: t.LastRun,
			Prompt:  "...", // You'd need to store this separately
		})
	}
	return infos, nil
}
