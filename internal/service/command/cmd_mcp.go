package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/sandevgo/tuskbot/internal/core"
)

type MCPCommand struct {
	mcp       core.MCPServer
	formatter *ResponseFormatter
}

func NewMCPCommand(mcp core.MCPServer) core.Command {
	return &MCPCommand{
		mcp:       mcp,
		formatter: NewResponseFormatter(),
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

	if len(tools) == 0 {
		return c.formatter.Combine(
			c.formatter.Info("MCP Tools"),
			c.formatter.Label("Status", "No MCP tools are currently connected."),
			c.formatter.Tip("Check your MCP server configuration if tools should be available"),
		), nil
	}

	toolNames := make([]string, len(tools))
	for i, tool := range tools {
		description := strings.ReplaceAll(strings.ReplaceAll(tool.Function.Description, "\n", " "), "\r", " ")
		description = strings.Join(strings.Fields(description), " ")
		if len(description) > 120 {
			description = description[:117] + "..."
		}
		toolNames[i] = fmt.Sprintf("**%s**", tool.Function.Name)
	}

	return c.formatter.Combine(
		c.formatter.Info("MCP Tools"),
		c.formatter.Label("Connected tools", fmt.Sprintf("%d", len(tools))),
		"\n",
		c.formatter.List(toolNames),
	), nil
}
