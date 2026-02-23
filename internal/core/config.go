package core

import "context"

type ConfigReader interface {
	GetModel() string
	GetContextWindowSize() int
}

type ConfigWriter interface {
	SetModel(model string) error
	Persist() error
}

type GlobalState interface {
	ChangeModel(ctx context.Context, model string) error
	// SetProvider(ctx context.Context, provider string) error
}
