package scheduler

import "time"

// OneOffTrigger At implementation
type OneOffTrigger struct {
	At    time.Time
	fired bool
}

func (t *OneOffTrigger) NextFireTime(now time.Time) time.Time {
	if t.fired {
		return time.Time{}
	}
	t.fired = true
	return t.At
}
