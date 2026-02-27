package scheduler

import (
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
)

type Task struct {
	ID   string
	Name string

	Trigger core.Trigger
	Job     Job

	LastRun time.Time
	NextRun time.Time
}
