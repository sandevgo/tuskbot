package swarm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/internal/service/agent"
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

// AddTask schedules a new autonomous task with its own session
func (s *Service) AddTask(name string, trigger core.Trigger, instruction string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a dedicated session ID for this task
	sessionID := fmt.Sprintf("task-%s", name)

	// Create the Job that will run when triggered
	job := func(ctx context.Context) error {
		logger := log.FromCtx(ctx).With().Str("task", name).Str("session", sessionID).Logger()
		ctx = logger.WithContext(ctx)

		logger.Info().Str("instruction", instruction).Msg("executing scheduled task")

		// Execute via Agent in dedicated session
		// The onUpdate callback can be used to send notifications back to user if needed
		_, err := s.agent.Run(ctx, sessionID, instruction, func(msg core.Message) {
			// Optional: Handle streaming updates from the agent
			// For async tasks, you might want to send Telegram notifications here
			// or just log the completion
			if msg.Content != "" {
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

	s.tasks[name] = task
	s.scheduler.AddTask(name, trigger, job)

	return nil
}

func (s *Service) CancelTask(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from local tracking
	delete(s.tasks, name)

	// Note: You'll need to add a RemoveTask method to your Scheduler interface
	// or implement cancellation via context cancellation
	return nil
}

func (s *Service) ListTasks() []TaskInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var infos []TaskInfo
	for _, t := range s.tasks {
		infos = append(infos, TaskInfo{
			ID:          t.ID,
			Name:        t.Name,
			NextRun:     t.NextRun,
			LastRun:     t.LastRun,
			Instruction: "...", // You'd need to store this separately
		})
	}
	return infos
}
