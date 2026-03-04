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
	// TaskExecutionTimeout is the maximum time allowed for a subagent task to complete
	TaskExecutionTimeout = 10 * time.Minute
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
	if err := s.validate(task, onComplete); err != nil {
		return "", err
	}

	// Add overall timeout for the task execution
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

	return s.runReActLoop(ctx, task.SessionID, messages, tools, onComplete)
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

func (s *SubAgent) runReActLoop(ctx context.Context, sessionID string, messages []core.Message, tools []core.Tool, onComplete core.UpdateFunc) (string, error) {
	var finalContent string

	for {
		responseMsg, err := s.queryAI(ctx, sessionID, messages, tools)
		if err != nil {
			return "", err
		}

		messages = append(messages, responseMsg)

		if responseMsg.Content != "" {
			onComplete(responseMsg)
			finalContent = responseMsg.Content
		}

		if len(responseMsg.ToolCalls) == 0 {
			break
		}

		toolResults := s.executor.Execute(ctx, responseMsg.ToolCalls)
		messages = append(messages, toolResults...)
	}

	return finalContent, nil
}

func (s *SubAgent) queryAI(ctx context.Context, sessionID string, messages []core.Message, tools []core.Tool) (core.Message, error) {
	chatCtx, cancel := context.WithTimeout(ctx, ChatTimeout)
	defer cancel()

	logger := log.FromCtx(ctx)
	logger.Debug().Str("session_id", sessionID).Msg("subagent querying AI")

	responseMsg, err := s.ai.Chat(chatCtx, messages, tools)
	if err != nil {
		return core.Message{}, fmt.Errorf("ai chat: %w", err)
	}

	logger.Debug().Str("session_id", sessionID).Msg("subagent received AI response")
	return responseMsg, nil
}
