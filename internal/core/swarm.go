package core

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type Swarm interface {
	ScheduleTask(ctx context.Context, ownerSessionID, name, taskType, timeSpec, instruction string) error
	CancelTask(ctx context.Context, name string) error
	ListTasks(ctx context.Context) ([]Task, error)
}

type Task struct {
	ID             uuid.UUID
	Name           string
	Prompt         string
	SessionID      string
	OwnerSessionID string
	IsActive       bool

	Trigger Trigger
	Job     Job

	LastRun time.Time
	NextRun time.Time
}

func (t *Task) NextFireTime(now time.Time) time.Time {
	return t.Trigger.NextFireTime(now, t.LastRun)
}
