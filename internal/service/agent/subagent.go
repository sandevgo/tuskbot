package agent

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
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
	if onComplete == nil {
		return "", fmt.Errorf("onComplete function is nil")
	}

	logger := log.FromCtx(ctx)

	logger.Debug().
		Str("session_id", task.SessionID).
		Msg("subagent received task")

	messages, err := s.memory.GetTaskContext(ctx, task.SessionID, task.Prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get context: %w", err)
	}

	tools, err := s.mcp.GetTools(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get tools: %w", err)
	}

	var finalContent string

	for {
		logger.Debug().
			Str("session_id", task.SessionID).
			Msg("subagent sending request to llm")

		chatCtx, cancel := context.WithTimeout(ctx, ChatTimeout)
		responseMsg, err := s.ai.Chat(chatCtx, messages, tools)
		cancel()

		if err != nil {
			return "", fmt.Errorf("ai chat error: %w", err)
		}

		logger.Debug().
			Str("session_id", task.SessionID).
			Msg("subagent received llm response")

		messages = append(messages, responseMsg)

		if responseMsg.Content != "" {
			onComplete(responseMsg)
			finalContent = responseMsg.Content
		}

		if len(responseMsg.ToolCalls) == 0 {
			break
		}

		// 5. Execute Tool Calls
		logger.Debug().
			Str("session_id", task.SessionID).
			Msg("subagent called mcp tool")

		toolResults := s.executor.Execute(ctx, responseMsg.ToolCalls)
		messages = append(messages, toolResults...)
	}

	return finalContent, nil
}
