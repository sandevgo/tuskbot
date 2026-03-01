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
	AddTask(name string, trigger Trigger, job Job)
}

type Trigger interface {
	NextFireTime(now time.Time, lastRun time.Time) time.Time
}

type Job func(ctx context.Context) error
