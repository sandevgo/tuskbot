package command

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
)

type HelpCommand struct {
	commands  []core.Command
	formatter *ResponseFormatter
}

func NewHelpCommand(commands []core.Command) core.Command {
	return &HelpCommand{
		commands:  commands,
		formatter: NewResponseFormatter(),
	}
}

func (c *HelpCommand) Name() string {
	return "help"
}

func (c *HelpCommand) Description() string {
	return "List all available commands"
}

func (c *HelpCommand) Execute(ctx context.Context, sessionID string, args []string) (string, error) {
	var list []string
	for _, cmd := range c.commands {
		list = append(list, fmt.Sprintf("/**%s** - %s", cmd.Name(), cmd.Description()))
	}

	return c.formatter.Combine(
		c.formatter.Info("Welcome to TuskBot!"),
		c.formatter.Label("Version", core.TaskVersion),
		c.formatter.List(list),
	), nil
}
