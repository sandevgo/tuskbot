package agent

import (
	"context"

	"github.com/sandevgo/tuskbot/internal/core"
)

type SubAgent struct {
	ai       core.AIProvider
	mcp      core.MCPServer
	memory   core.Memory
	executor *Executor
}

func NewSubAgent(ai core.AIProvider, mcp core.MCPServer, memory core.Memory, executor *Executor) *SubAgent {
	return &SubAgent{
		ai:       ai,
		mcp:      mcp,
		memory:   memory,
		executor: executor,
	}
}

func (s *SubAgent) Run(ctx context.Context, task *core.Task, onComplete core.UpdateFunc) (string, error) {
	return "", nil
}
