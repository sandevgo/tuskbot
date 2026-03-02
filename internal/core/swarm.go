package core

import (
	"context"
	"time"
)

type Swarm interface {
	ScheduleTask(ctx context.Context, ownerSessionID, name, taskType, timeSpec, instruction string) error
	CancelTask(ctx context.Context, name string) error
	ListTasks(ctx context.Context) ([]Task, error)
}

type Task struct {
	ID             string
	Name           string
	Prompt         string
	OwnerSessionID string

	Trigger Trigger
	Job     Job

	LastRun time.Time
	NextRun time.Time
}

func (t *Task) NextFireTime(now time.Time) time.Time {
	return t.Trigger.NextFireTime(now, t.LastRun)
}
