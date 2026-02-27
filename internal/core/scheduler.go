package core

import "time"

type Trigger interface {
	NextFireTime(now time.Time) time.Time
}
