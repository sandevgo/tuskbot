package state

import (
	"context"
)

type configWriter interface {
	SetModel(string) error
	Pesrist() error
}

type GlobalState struct {
	writer configWriter
}

func NewGlobalState(
	writer configWriter,
) *GlobalState {
	return &GlobalState{
		config: configWriter,
	}
}

func (s *GlobalState) ChangeModel(ctx context.Context, model string) error {
	return s.config.SetModel(model)
}
