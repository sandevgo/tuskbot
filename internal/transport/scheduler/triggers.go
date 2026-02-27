package scheduler

import "time"

// CronTrigger Cron implementation
type CronTrigger struct {
	Expression string
}

// IntervalTrigger Heartbeat implementation
type IntervalTrigger struct {
	Interval time.Duration
	Last     time.Time
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
