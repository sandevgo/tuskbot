package scheduler

import "time"

// IntervalTrigger Heartbeat implementation
type IntervalTrigger struct {
	Every time.Duration
}
