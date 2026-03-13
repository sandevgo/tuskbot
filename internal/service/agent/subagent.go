package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

const (
	TaskExecutionTimeout = 10 * time.Minute
)

type SubAgent struct {
	runner *ReActRunner
	mcp    core.MCPServer
	memory core.Memory
}

func NewSubAgent(ai core.AIProvider, mcp core.MCPServer, memory core.Memory, executor core.ToolExecutor) *SubAgent {
	return &SubAgent{
		runner: NewReActRunner(ai, executor, ChatTimeout),
		mcp:    mcp,
		memory: memory,
	}
}

func (s *SubAgent) Run(ctx context.Context, task *core.Task, onComplete core.UpdateFunc) (string, error) {
	if err := s.validate(task, onComplete); err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, TaskExecutionTimeout)
	defer cancel()

	logger := log.FromCtx(ctx)
	logger.Info().
		Str("session_id", task.SessionID).
		Str("task_name", task.Name).
		Msg("subagent starting task execution")

	messages, err := s.memory.GetTaskContext(ctx, task.SessionID, task.Prompt)
	if err != nil {
		return "", fmt.Errorf("get context: %w", err)
	}

	tools, err := s.mcp.GetTools(ctx)
	if err != nil {
		return "", fmt.Errorf("get tools: %w", err)
	}

	result, err := s.runner.Run(ctx, messages, tools, func(msg core.Message) error {
		if msg.Role == core.RoleAssistant && msg.Content != "" {
			onComplete(msg)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	//s.memory.SaveMessage(ctx, task.OwnerSessionID, core.Message{
	//	Role:    core.RoleSystem,
	//	Content: fmt.Sprintf("Task '%s' completed.\n\nResult:\n%s", task.Name, result),
	//})

	return result, nil
}

func (s *SubAgent) validate(task *core.Task, onComplete core.UpdateFunc) error {
	if task == nil {
		return errors.New("task is nil")
	}
	if task.SessionID == "" {
		return errors.New("task session ID is empty")
	}
	if onComplete == nil {
		return errors.New("onComplete callback is nil")
	}
	return nil
}
