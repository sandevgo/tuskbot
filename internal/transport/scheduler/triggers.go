package scheduler

import (
	"fmt"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
)

// IntervalTrigger Heartbeat implementation
type IntervalTrigger struct {
	Interval time.Duration
}

func NewIntervalTrigger(d time.Duration) *IntervalTrigger {
	return &IntervalTrigger{Interval: d}
}

func (t *IntervalTrigger) NextFireTime(now time.Time, last time.Time) time.Time {
	if last.IsZero() {
		return now
	}
	return last.Add(t.Interval)
}

// OneOffTrigger At implementation
type OneOffTrigger struct {
	At    time.Time
	fired bool
}

func NewOneOffTrigger(at time.Time) *OneOffTrigger {
	return &OneOffTrigger{At: at}
}

func (t *OneOffTrigger) NextFireTime(now time.Time, last time.Time) time.Time {
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

func (c CronTrigger) NextFireTime(now time.Time, lastRun time.Time) time.Time {
	//TODO implement me
	return time.Time{}
}

func CreateTrigger(taskType, timeSpec string) (core.Trigger, error) {
	switch taskType {
	case core.TriggerTypeInterval:
		return createIntervalTrigger(timeSpec)
	case core.TriggerTypeOnce:
		return createOneOffTrigger(timeSpec)
	case core.TriggerTypeCron:
		return nil, fmt.Errorf("not implemented yet")
	default:
		return nil, fmt.Errorf("unknown task type: %s", taskType)
	}
}

func createIntervalTrigger(timeSpec string) (core.Trigger, error) {
	dur, err := time.ParseDuration(timeSpec)
	if err != nil {
		return nil, fmt.Errorf("invalid interval duration: %w", err)
	}
	return NewIntervalTrigger(dur), nil
}

func createOneOffTrigger(timeSpec string) (core.Trigger, error) {
	if dur, err := time.ParseDuration(timeSpec); err == nil {
		at := time.Now().Add(dur)
		return NewOneOffTrigger(at), nil
	}

	// Try parsing as ISO8601/RFC3339
	t, err := time.Parse(time.RFC3339, timeSpec)
	if err != nil {
		return nil, fmt.Errorf("invalid time_spec for 'once' (expected duration or RFC3339): %w", err)
	}
	return NewOneOffTrigger(t), nil
}
