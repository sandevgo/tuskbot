package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/sandevgo/tuskbot/internal/core"
)

type Router struct {
	commands  map[string]core.Command
	formatter *ResponseFormatter
}

func New(commands []core.Command) *Router {
	c := &Router{
		commands:  make(map[string]core.Command),
		formatter: NewResponseFormatter(),
	}

	for _, cmd := range commands {
		c.commands[cmd.Name()] = cmd
	}
	return c
}

func (c *Router) Execute(ctx context.Context, sessionID, input string) (string, bool) {
	if !strings.HasPrefix(input, "/") {
		return "", false
	}

	parts := strings.Fields(input)
	name := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	cmd, ok := c.commands[name]
	if !ok {
		return c.formatter.Combine(
			c.formatter.Info("Unknown Command"),
			fmt.Sprintf("**Command**: /%s", name),
			c.formatter.Usage("/help"),
			c.formatter.Tip("Use /help to see all available commands"),
		), true
	}

	result, err := cmd.Execute(ctx, sessionID, args)
	if err != nil {
		return c.formatter.Error(cmd.Name(), err), true
	}
	return result, true
}

func (c *Router) ListCommands() []core.Command {
	res := make([]core.Command, 0, len(c.commands))
	for _, cmd := range c.commands {
		res = append(res, cmd)
	}
	return res
}
