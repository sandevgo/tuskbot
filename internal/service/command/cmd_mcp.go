package command

import (
	"context"
	"fmt"
	"strings"

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
	tools, err := c.mcp.GetTools(ctx)
	if err != nil {
		return "", err
	}

	sb := strings.Builder{}
	sb.WriteString("Connected MCP tools:\n\n")
	for _, tool := range tools {
		sb.WriteString(fmt.Sprintf("- %s\n", tool.Function.Name))
	}

	return sb.String(), nil
}
