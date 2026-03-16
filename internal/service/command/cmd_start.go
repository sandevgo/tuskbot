package command

import (
	"context"

	"github.com/sandevgo/tuskbot/internal/core"
)

const Version = "0.1.0"

type StartCommand struct {
	formatter *ResponseFormatter
}

func NewStartCommand() core.Command {
	return &StartCommand{
		formatter: NewResponseFormatter(),
	}
}

func (c *StartCommand) Name() string {
	return "start"
}

func (c *StartCommand) Description() string {
	return "Start the bot and show version"
}

func (c *StartCommand) Execute(ctx context.Context, sessionID string, args []string) (string, error) {
	return c.formatter.Combine(
		c.formatter.Success("Welcome to TuskBot!"),
		c.formatter.Label("Version", Version),
		c.formatter.Tip("Use /help to see available commands."),
	), nil
}
