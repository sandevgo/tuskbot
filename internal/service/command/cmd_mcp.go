package command

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
)

type MCPCommand struct {
	mcp core.MCPServer
}

func NewMCPCommand(mcp core.MCPServer) core.Command {
	return &MCPCommand{
		mcp: mcp,
	}
}

func (c *MCPCommand) Name() string {
	return "mcp"
}

func (c *MCPCommand) Description() string {
	return "Show connected MCP servers"
}

func (c *MCPCommand) Execute(ctx context.Context, sessionID string, args []string) (string, error) {
	return fmt.Sprintf("Current model: %s\n", c.cfg.MainModel), nil
}
