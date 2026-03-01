package scheduler

import (
	"time"
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
	panic("implement me")
}
