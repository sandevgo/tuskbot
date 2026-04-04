package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

const ChatTimeout = 2 * time.Minute

type Agent struct {
	runner *ReActRunner
	mcp    core.MCPServer
	memory core.Memory
	events core.EventPublisher
	lock   chan struct{}
}

func NewAgent(
	ai core.AIProvider,
	mcp core.MCPServer,
	memory core.Memory,
	executor core.ToolExecutor,
	events core.EventPublisher,
) *Agent {
	lock := make(chan struct{}, 1)
	lock <- struct{}{}
	return &Agent{
		runner: NewReActRunner(ai, executor, ChatTimeout),
		mcp:    mcp,
		memory: memory,
		events: events,
		lock:   lock,
	}
}

func (a *Agent) Run(ctx context.Context, sessionID string, input string, onUpdate core.UpdateFunc) (string, error) {
	select {
	case <-ctx.Done():
		return "", nil
	case <-a.lock:
		defer func() { a.lock <- struct{}{} }()
	}

	logger := log.FromCtx(ctx)

	logger.Debug().
		Str("session_id", sessionID).
		Msg("agent received user request")

	userMsg := core.Message{Role: core.RoleUser, Content: input}
	if err := a.memory.SaveMessage(ctx, sessionID, userMsg); err != nil {
		return "", fmt.Errorf("failed to save user message: %w", err)
	}

	promptCtx, err := a.memory.GetFullContext(ctx, sessionID, input)
	if err != nil {
		return "", fmt.Errorf("failed to get context: %w", err)
	}

	tools, err := a.mcp.GetTools(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get tools: %w", err)
	}

	promptCtx.Messages = sanitizeToolCalls(ctx, promptCtx.Messages)
	req := core.NewChatRequest(promptCtx, tools)

	finalContent, err := a.runner.Run(ctx, req, func(msg core.Message) error {
		if err := a.memory.SaveMessage(ctx, sessionID, msg); err != nil {
			return fmt.Errorf("failed to save message: %w", err)
		}

		if msg.Role == core.RoleAssistant && onUpdate != nil {
			onUpdate(msg)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return finalContent, nil
}

func (a *Agent) Notify(ctx context.Context, task *core.Task, result string) error {
	logger := log.FromCtx(ctx)

	msg := core.Message{
		Role:    core.RoleSystem,
		Content: fmt.Sprintf("Background Task '%s' completed.\n%s", task.Name, result),
	}
	if err := a.memory.SaveMessage(ctx, task.OwnerSessionID, msg); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-a.lock:
		defer func() { a.lock <- struct{}{} }()
	}

	promptCtx, err := a.memory.GetFullContext(ctx, task.OwnerSessionID, "")
	if err != nil {
		return err
	}

	promptCtx.Messages = append(promptCtx.Messages, core.Message{
		Role:    core.RoleUser,
		Content: "Please inform me about the completed background task and take any necessary follow-up actions.",
	})

	tools, err := a.mcp.GetTools(ctx)
	if err != nil {
		return err
	}

	promptCtx.Messages = sanitizeToolCalls(ctx, promptCtx.Messages)
	req := core.NewChatRequest(promptCtx, tools)

	final, err := a.runner.Run(ctx, req, func(m core.Message) error {
		return a.memory.SaveMessage(ctx, task.OwnerSessionID, m)
	})

	if err != nil {
		logger.Error().Err(err).Msg("Runner returned an error")
		return err
	}

	event := core.NewChatEvent(core.EventTypeTaskCompleted, task.OwnerSessionID, final)
	a.events.Publish(ctx, event)
	return nil
}

func sanitizeToolCalls(ctx context.Context, messages []core.Message) []core.Message {
	var sanitized []core.Message
	var validIDs map[string]bool
	for _, msg := range messages {
		switch msg.Role {
		case core.RoleAssistant:
			validIDs = make(map[string]bool)
			for _, tc := range msg.ToolCalls {
				validIDs[tc.ID] = true
			}
			sanitized = append(sanitized, msg)
		case core.RoleTool:
			if validIDs != nil && validIDs[msg.ToolCallID] {
				sanitized = append(sanitized, msg)
			}
		default:
			sanitized = append(sanitized, msg)
			validIDs = nil
		}
	}
	return sanitized
}
