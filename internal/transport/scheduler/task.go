package scheduler

import (
	"context"

	"github.com/sandevgo/tuskbot/internal/core"
)

type Task struct {
	ID      string
	Trigger core.Trigger              // Полиморфизм: тут может быть CronTrigger или IntervalTrigger
	Job     func(ctx context.Context) // Сама работа (спаун агента, проверка URL)
}
