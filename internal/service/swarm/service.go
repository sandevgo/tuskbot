package swarm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
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
	agent     core.Agent
	taskRepo  core.TaskRepository
	mu        sync.RWMutex
	tasks     map[string]*core.Task
}

func NewService(
	scheduler core.Scheduler,
	agent core.Agent,
	taskRepo core.TaskRepository,
) *Service {
	return &Service{
		scheduler: scheduler,
		agent:     agent,
		taskRepo:  taskRepo,
		tasks:     make(map[string]*core.Task),
	}
}

func (s *Service) ScheduleTask(ctx context.Context, ownerSessionID, name, taskType, timeSpec, instruction string) error {
	sessionID := generateSessionID(name)

	trigger, err := scheduler.CreateTrigger(taskType, timeSpec)
	if err != nil {
		return fmt.Errorf("create trigger: %w", err)
	}

	job := s.createJob(sessionID, instruction)

	task := &core.Task{
		ID:             name,
		Name:           name,
		Prompt:         instruction,
		OwnerSessionID: ownerSessionID,
		Trigger:        trigger,
		Job:            job,
	}

	// TODO: Persist task to database
	storedTask, err := s.taskRepo.Create(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to persist task: %w", err)
	}

	s.registerTask(name, storedTask)
	return nil
}

func generateSessionID(name string) string {
	return fmt.Sprintf("task-%s", name)
}

func (s *Service) createJob(sessionID, instruction string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		logger := log.FromCtx(ctx)
		logger.Info().Str("session", sessionID).Msg("executing scheduled task")

		_, err := s.agent.Run(ctx, sessionID, instruction, func(msg core.Message) {
			if msg.Content != "" {
				logger.Debug().Str("output", msg.Content).Msg("task progress")
			}
		})
		return err
	}
}

func (s *Service) registerTask(name string, task *core.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[name] = task
	s.scheduler.AddTask(task)
}

func (s *Service) CancelTask(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from local tracking
	delete(s.tasks, name)

	// TODO: get task from repo
	//		then s.scheduler.DelTask(tasks)
	return nil
}

func (s *Service) ListTasks(ctx context.Context) ([]core.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var infos []core.Task
	for _, t := range s.tasks {
		infos = append(infos, core.Task{
			ID:             t.ID,
			Name:           t.Name,
			NextRun:        t.NextRun,
			LastRun:        t.LastRun,
			Prompt:         t.Prompt,
			OwnerSessionID: t.OwnerSessionID,
		})
	}
	return infos, nil
}
