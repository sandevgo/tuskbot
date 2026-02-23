package state

import (
	"context"

	"github.com/sandevgo/tuskbot/internal/core"
)

type GlobalState struct {
	reader core.ConfigReader
	writer core.ConfigWriter
}

func NewGlobalState(
	reader core.ConfigReader,
	writer core.ConfigWriter,
) *GlobalState {
	return &GlobalState{
		reader: reader,
		writer: writer,
	}
}

func (s *GlobalState) ChangeModel(ctx context.Context, model string) error {
	return s.writer.SetModel(model)
}
