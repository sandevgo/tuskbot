package scheduler

import (
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
)

// IntervalTrigger Heartbeat implementation
type IntervalTrigger struct {
	Interval time.Duration
	Last     time.Time
}

func NewIntervalTrigger(d time.Duration) core.Trigger {
	return &IntervalTrigger{Interval: d}
}

func (t *IntervalTrigger) NextFireTime(now time.Time) time.Time {
	if t.Last.IsZero() {
		return now
	}
	return t.Last.Add(t.Interval)
}

// OneOffTrigger At implementation
type OneOffTrigger struct {
	At    time.Time
	fired bool
}

func NewOneOffTrigger(at time.Time) core.Trigger {
	return &OneOffTrigger{At: at}
}

func (t *OneOffTrigger) NextFireTime(now time.Time) time.Time {
	if t.fired {
		return time.Time{} // Done forever
	}
	t.fired = true
	if t.At.Before(now) {
		return now // Execute immediately if already passed
	}
	return t.At
}

// CronTrigger Cron implementation
// TODO: implement cron trigger with robfig/cron later
type CronTrigger struct {
	Expression string
}
