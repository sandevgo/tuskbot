package state

import (
	"context"
)

type configWriter interface {
	SetModel(string) error
}

type GlobalState struct {
	config configWriter
}

func NewGlobalState(
	writer configWriter,
) *GlobalState {
	return &GlobalState{
		config: writer,
	}
}

func (s *GlobalState) ChangeModel(ctx context.Context, model string) error {
	return s.config.SetModel(model)
}
