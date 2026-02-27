package scheduler

import "time"

// IntervalTrigger Heartbeat implementation
type IntervalTrigger struct {
	Every time.Duration
	Last  time.Time
}

func (t *IntervalTrigger) NextFireTime(now time.Time) time.Time {
	return t.Last.Add(t.Every)
}
