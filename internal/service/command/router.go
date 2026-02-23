package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/core"
)

type Router struct {
	commands map[string]core.Command
	memory   core.Memory
	cfg      *config.AppConfig
}

func New(cfg *config.AppConfig, memory core.Memory) *Router {
	c := &Router{
		commands: make(map[string]core.Command),
		memory:   memory,
		cfg:      cfg,
	}
	// Register built-in commands
	//c.Register(&ResetCommand{memory: memory})
	//c.Register(&ClearCommand{memory: memory})
	//c.Register(&StatusCommand{cfg: cfg})
	//c.Register(&HelpCommand{commander: c})
	return c
}

func (c *Router) Register(cmd core.Command) {
	c.commands[cmd.Name()] = cmd
}

func (c *Router) Execute(ctx context.Context, sessionID, input string) (string, bool) {
	if !strings.HasPrefix(input, "/") {
		return "", false // Not a command, pass to Agent
	}

	parts := strings.Fields(input)
	name := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	cmd, ok := c.commands[name]
	if !ok {
		return fmt.Sprintf("Unknown command: /%s", name), true
	}

	result, err := cmd.Execute(ctx, sessionID, args)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), true
	}
	return result, true
}
