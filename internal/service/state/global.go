package state

import (
	"context"
)

type provider interface {
	SetModel(ctx context.Context, model string) error
}

type GlobalState struct {
	provider provider
}

func NewGlobalState(
	provider provider,
) *GlobalState {
	return &GlobalState{
		provider: provider,
	}
}

func (s *GlobalState) ChangeModel(ctx context.Context, model string) error {
	return s.provider.SetModel(ctx, model)
}
