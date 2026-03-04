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
	subagent  core.SubAgent
	taskRepo  core.TaskRepository
	mu        sync.RWMutex
	tasks     map[string]*core.Task
}

func NewService(
	scheduler core.Scheduler,
	agent core.SubAgent,
	taskRepo core.TaskRepository,
) *Service {
	return &Service{
		scheduler: scheduler,
		subagent:  agent,
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

	trigger, err := scheduler.CreateTrigger(taskType, timeSpec)
	if err != nil {
		return fmt.Errorf("create trigger: %w", err)
	}

	task := &core.Task{
		ID:             taskID,
		Name:           name,
		Prompt:         instruction,
		SessionID:      generateSessionID(name),
		OwnerSessionID: ownerSessionID,
		Trigger:        trigger,
	}

	s.assignJob(task)
	s.registerTask(task)
	return nil
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
		return err
	}
}

func (s *Service) registerTask(task *core.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[task.Name] = task
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

func (s *Service) toDomain(st core.StoredTask) (core.Task, error) {
	trigger, err := scheduler.CreateTrigger(st.TriggerType, st.TriggerSpec)
	if err != nil {
		return core.Task{}, err
	}

	return core.Task{
		ID:             st.ID,
		Name:           st.Name,
		OwnerSessionID: st.OwnerSessionID,
		Prompt:         st.Prompt,
		Trigger:        trigger,
		Job:            s.assignJob(st.OwnerSessionID, st.Prompt),
		LastRun:        st.LastRun,
	}, nil
}
