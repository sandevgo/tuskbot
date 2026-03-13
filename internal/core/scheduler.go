package core

import (
	"context"
	"time"
)

const (
	TriggerTypeCron     = "cron"
	TriggerTypeOnce     = "once"
	TriggerTypeInterval = "interval"
)

type Scheduler interface {
	AddTask(task *Task)
	DelTask(task *Task)
}

type Trigger interface {
	NextFireTime(now time.Time, lastRun time.Time) time.Time
}

type Job func(ctx context.Context) error
