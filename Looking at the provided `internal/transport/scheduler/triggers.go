// OneOffTrigger At implementation
type OneOffTrigger struct {
	At time.Time
}

func NewOneOffTrigger(at time.Time) *OneOffTrigger {
	return &OneOffTrigger{At: at}
}

func (t *OneOffTrigger) NextFireTime(now time.Time, last time.Time) time.Time {
	if !last.IsZero() {
		return time.Time{}
	}

	if t.At.Before(now) {
		return now
	}

	return t.At
}
