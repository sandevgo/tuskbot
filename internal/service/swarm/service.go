package swarm

import (
	"context"
	"fmt"
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
}

func NewService(
	scheduler core.Scheduler,
	agent core.Agent,
	subagent core.SubAgent,
	taskRepo core.TaskRepository,
) *Service {
	return &Service{
		scheduler: scheduler,
		agent:     agent,
		subagent:  subagent,
		taskRepo:  taskRepo,
	}
}

func (s *Service) Start(ctx context.Context) error {
	storedTasks, err := s.taskRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tasks on startup: %w", err)
	}

	for _, st := range storedTasks {
		task, err := s.toDomain(&st)
		if err != nil {
			log.FromCtx(ctx).Error().Err(err).Str("task", st.Name).Msg("failed to hydrate task")
			continue
		}
		s.registerTask(task)
	}
	return nil
}

func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}

func (s *Service) ScheduleTask(ctx context.Context, ownerSessionID, name, taskType, timeSpec, instruction string) error {
	taskID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	now := time.Now()
	storedTask := &core.StoredTask{
		ID:             taskID,
		Name:           name,
		Prompt:         instruction,
		TriggerType:    taskType,
		TriggerSpec:    timeSpec,
		OwnerSessionID: ownerSessionID,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      &now,
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

func (s *Service) CancelTask(ctx context.Context, id string) error {
	err := s.taskRepo.Cancel(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to cancel task in repo: %w", err)
	}

	// Map will be more effective, but it's fine for small scale
	for _, task := range s.scheduler.ListTasks() {
		if task.ID.String() == id {
			s.scheduler.DelTask(task)
			return nil
		}
	}

	return fmt.Errorf("task not found in scheduler: %s", id)
}

func (s *Service) ListTasks(ctx context.Context) ([]core.Task, error) {
	tasks := s.scheduler.ListTasks()

	var infos []core.Task
	for _, t := range tasks {
		infos = append(infos, core.Task{
			ID:             t.ID,
			Name:           t.Name,
			Prompt:         t.Prompt,
			OwnerSessionID: t.OwnerSessionID,
			Trigger:        t.Trigger,
			LastRun:        t.LastRun,
			NextRun:        t.NextRun,
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
		IsActive:       st.IsActive,
		Trigger:        trigger,
		LastRun:        st.LastRun,
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

		result, err := s.subagent.Run(ctx, task, func(msg core.Message) {
			if msg.Content != "" {
				logger.Debug().Str("output", msg.Content).Msg("task progress")
			}
		})
		if err != nil {
			return err
		}

		// Update and persist LastRun
		task.LastRun = time.Now()
		if err := s.taskRepo.UpdateExecution(ctx, task.ID, task.LastRun); err != nil {
			logger.Error().Err(err).Str("task", task.Name).Msg("failed to persist last_run")
		}

		return s.agent.Notify(ctx, task, result)
	}
}

func (s *Service) registerTask(task *core.Task) {
	s.scheduler.AddTask(task)
}
