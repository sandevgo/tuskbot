package swarm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofrs/uuid"
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
	subagent  core.SubAgent
	taskRepo  core.TaskRepository
	mu        sync.RWMutex
	tasks     map[string]*core.Task
}

func NewService(
	scheduler core.Scheduler,
	subagent core.SubAgent,
	taskRepo core.TaskRepository,
) *Service {
	return &Service{
		scheduler: scheduler,
		subagent:  subagent,
		taskRepo:  taskRepo,
		tasks:     make(map[string]*core.Task),
	}
}

func (s *Service) ScheduleTask(ctx context.Context, ownerSessionID, name, taskType, timeSpec, instruction string) error {
	taskID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	storedTask := &core.StoredTask{
		ID:             taskID,
		Name:           name,
		Prompt:         instruction,
		TriggerType:    taskType,
		TriggerSpec:    timeSpec,
		OwnerSessionID: ownerSessionID,
	}

	err = s.taskRepo.Create(ctx, storedTask)
	if err != nil {
		return fmt.Errorf("failed to persist task: %w", err)
	}

	task, err := s.toDomain(storedTask)
	if err != nil {
		return fmt.Errorf("failed to convert to domain task: %w", err)
	}

	s.registerTask(task)
	return nil
}

func (s *Service) CancelTask(ctx context.Context, name string) error {
	storedTask, err := s.taskRepo.GetByName(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get task by name %w", err)
	}

	task, err := s.toDomain(storedTask)
	if err != nil {
		return fmt.Errorf("failed to convert to domain task %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from local tracking
	delete(s.tasks, name)
	s.scheduler.DelTask(task)

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

func (s *Service) toDomain(st *core.StoredTask) (*core.Task, error) {
	trigger, err := scheduler.CreateTrigger(st.TriggerType, st.TriggerSpec)
	if err != nil {
		return nil, fmt.Errorf("create trigger: %w", err)
	}

	task := &core.Task{
		ID:             st.ID,
		Name:           st.Name,
		Prompt:         st.Prompt,
		SessionID:      generateSessionID(st.Name),
		OwnerSessionID: st.OwnerSessionID,
		Trigger:        trigger,
	}

	s.assignJob(task)
	return task, nil
}

func generateSessionID(name string) string {
	return fmt.Sprintf("task-%s", name)
}

func (s *Service) assignJob(task *core.Task) {
	task.Job = func(ctx context.Context) error {
		logger := log.FromCtx(ctx)
		logger.Info().Str("session", task.SessionID).Msg("executing scheduled task")

		_, err := s.subagent.Run(ctx, task, func(msg core.Message) {
			if msg.Content != "" {
				logger.Debug().Str("output", msg.Content).Msg("task progress")
			}
		})
		if err != nil {
			return err
		}

		return s.agent.Notify(ctx, task)
	}
}

func (s *Service) registerTask(task *core.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[task.Name] = task
	s.scheduler.AddTask(task)
}
