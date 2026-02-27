package scheduler

import "context"

type Job func(ctx context.Context) error
