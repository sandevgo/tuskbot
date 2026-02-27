package scheduler

import "time"

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
