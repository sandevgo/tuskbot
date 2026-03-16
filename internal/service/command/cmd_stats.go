package command

import (
	"context"

	"github.com/sandevgo/tuskbot/internal/core"
)

type StatsCommand struct {
	formatter *ResponseFormatter
}

func NewStatsCommand() core.Command {
	return &StatsCommand{
		formatter: NewResponseFormatter(),
	}
}

func (c *StatsCommand) Name() string {
	return "stats"
}

func (c *StatsCommand) Description() string {
	return "Show session statistics"
}

func (c *StatsCommand) Execute(ctx context.Context, sessionID string, args []string) (string, error) {
	return "", nil
}
