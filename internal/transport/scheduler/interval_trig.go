package scheduler

import "time"

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
