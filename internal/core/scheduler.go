package core

import (
	"context"
	"time"
)

type Scheduler interface {
	AddTask(name string, trigger Trigger, job Job)
}

type Trigger interface {
	NextFireTime(now time.Time, lastRun time.Time) time.Time
}

type Job func(ctx context.Context) error

type Task struct {
	ID   string
	Name string

	Trigger Trigger
	Job     Job

	LastRun time.Time
	NextRun time.Time
}
