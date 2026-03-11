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
	runner   *ReActRunner
	mcp      core.MCPServer
	memory   core.Memory
	events   core.EventPublisher
	sessions core.SessionManager
}

func NewAgent(
	ai core.AIProvider,
	mcp core.MCPServer,
	memory core.Memory,
	executor core.ToolExecutor,
	events core.EventPublisher,
	sessions core.SessionManager,
) *Agent {
	return &Agent{
		runner:   NewReActRunner(ai, executor, ChatTimeout),
		mcp:      mcp,
		memory:   memory,
		events:   events,
		sessions: sessions,
	}
}

func (a *Agent) Run(ctx context.Context, sessionID string, input string, onUpdate core.UpdateFunc) (string, error) {
	if !a.sessions.TryLock(sessionID) {
		return "", fmt.Errorf("agent is busy")
	}
	defer a.sessions.Unlock(sessionID)

	logger := log.FromCtx(ctx)

	logger.Debug().
		Str("session_id", sessionID).
		Msg("agent received user request")

	userMsg := core.Message{Role: core.RoleUser, Content: input}
	if err := a.memory.SaveMessage(ctx, sessionID, userMsg); err != nil {
		return "", fmt.Errorf("failed to save user message: %w", err)
	}

	messages, err := a.memory.GetFullContext(ctx, sessionID, input)
	if err != nil {
		return "", fmt.Errorf("failed to get context: %w", err)
	}

	tools, err := a.mcp.GetTools(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get tools: %w", err)
	}

	finalContent, err := a.runner.Run(ctx, sanitizeToolCalls(ctx, messages), tools, func(msg core.Message) error {
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
	msg := core.Message{
		Role:    core.RoleSystem,
		Content: fmt.Sprintf("Background Task '%s' completed. Send\n %s", task.Name, result),
	}
	if err := a.memory.SaveMessage(ctx, task.OwnerSessionID, msg); err != nil {
		return err
	}

	for {
		if a.sessions.TryLock(task.OwnerSessionID) {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
			continue
		}
	}
	defer a.sessions.Unlock(task.OwnerSessionID)

	messages, err := a.memory.GetFullContext(ctx, task.OwnerSessionID, "")
	if err != nil {
		return err
	}

	tools, err := a.mcp.GetTools(ctx)
	if err != nil {
		return err
	}

	final, err := a.runner.Run(ctx, sanitizeToolCalls(ctx, messages), tools, func(m core.Message) error {
		return a.memory.SaveMessage(ctx, task.OwnerSessionID, m)
	})
	if err != nil {
		return err
	}

	a.events.Publish(ctx, core.NewChatEvent(core.EventTypeTaskCompleted, task.OwnerSessionID, final))
	return nil
}

func sanitizeToolCalls(ctx context.Context, messages []core.Message) []core.Message {
	var sanitized []core.Message
	var validIDs map[string]bool
	for _, msg := range messages {
		switch msg.Role {
		case core.RoleSystem:
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
			validIDs = nil
			sanitized = append(sanitized, msg)
		}
	}
	return sanitized
}
