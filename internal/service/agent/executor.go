package agent

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
)

type Executor struct {
	mcp core.MCPServer
}

func NewExecutor(mcp core.MCPServer) *Executor {
	return &Executor{
		mcp: mcp,
	}
}

func (e *Executor) Execute(ctx context.Context, toolCalls []core.ToolCall) []core.Message {
	var results []core.Message
	for _, tc := range toolCalls {
		res, err := e.mcp.CallTool(ctx, tc.Function.Name, tc.Function.Arguments)
		if err != nil {
			res = fmt.Sprintf("Error: %v", err)
		}

		results = append(results, core.Message{
			Role:       core.RoleTool,
			Content:    res, // TODO: enable truncate only for specific tools
			ToolCallID: tc.ID,
		})
	}
	return results
}

func (e *Executor) truncate(input string) string {
	const maxLen = 2000
	if len(input) <= maxLen {
		return input
	}

	head := input[:500]
	tail := input[len(input)-(maxLen-500):]
	return fmt.Sprintf("%s\n\n... [TRUNCATED %d bytes] ...\n\n%s", head, len(input)-maxLen, tail)
}
