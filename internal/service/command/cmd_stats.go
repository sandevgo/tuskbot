package command

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
)

type StatsCommand struct {
	memory    core.Memory
	formatter *ResponseFormatter
}

func NewStatsCommand(memory core.Memory) core.Command {
	return &StatsCommand{
		memory:    memory,
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
	messages, err := c.memory.GetFullContext(ctx, sessionID, "")
	if err != nil {
		return "", err
	}

	tokens := c.memory.CountTokens(messages)

	return c.formatter.Combine(
		c.formatter.Info("Session Statistics"),
		c.formatter.Label("Session ID", sessionID),
		c.formatter.Label("Context Size", fmt.Sprintf("%d tokens", tokens)),
		c.formatter.Label("Messages", fmt.Sprintf("%d", len(messages))),
	), nil
}
