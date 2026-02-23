package state

import (
	"context"
)

type persister interface {
	SetModel(string) error
}

type GlobalState struct {
	config persister
}

func NewGlobalState(
	writer persister,
) *GlobalState {
	return &GlobalState{
		config: writer,
	}
}

func (s *GlobalState) ChangeModel(ctx context.Context, model string) error {
	return s.config.SetModel(model)
}
