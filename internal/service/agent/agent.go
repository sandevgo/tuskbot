package agent

import (
	"context"
	"fmt"
	"strings"
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
}

func NewAgent(
	ai core.AIProvider,
	mcp core.MCPServer,
	memory core.Memory,
	executor core.ToolExecutor,
	events core.EventPublisher,
) *Agent {
	return &Agent{
		runner: NewReActRunner(ai, executor, ChatTimeout),
		mcp:    mcp,
		memory: memory,
		events: events,
	}
}

func (a *Agent) Run(ctx context.Context, sessionID string, input string, onUpdate core.UpdateFunc) (string, error) {
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

	messages = sanitizeToolCalls(ctx, messages)

	tools, err := a.mcp.GetTools(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get tools: %w", err)
	}

	finalContent, err := a.runner.Run(ctx, messages, tools, func(msg core.Message) error {
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

func (a *Agent) Notify(ctx context.Context, task *core.Task) error {
	logger := log.FromCtx(ctx)

	messages, err := a.memory.GetMessages(ctx, task.OwnerSessionID, 5)
	if err != nil {
		return fmt.Errorf("failed to get messages: %w", err)
	}

	var result string
	for _, msg := range messages {
		if msg.Role == core.RoleSystem && strings.Contains(msg.Content, "Task '"+task.Name+"'") {
			result = msg.Content
			break
		}
	}

	if result == "" {
		result = fmt.Sprintf("Task '%s' completed", task.Name)
	}

	a.events.Publish(ctx, core.NewChatEvent(core.EventTypeTaskCompleted, task.OwnerSessionID, result))

	logger.Info().Str("task", task.Name).Str("session", task.OwnerSessionID).Msg("task completion notified")
	return nil
}

func sanitizeToolCalls(ctx context.Context, messages []core.Message) []core.Message {
	logger := log.FromCtx(ctx)
	var sanitized []core.Message
	var validToolCallIDs map[string]bool

	for i, msg := range messages {
		switch msg.Role {
		case core.RoleUser, core.RoleSystem:
			validToolCallIDs = nil
			sanitized = append(sanitized, msg)

		case core.RoleAssistant:
			validToolCallIDs = make(map[string]bool)
			for _, tc := range msg.ToolCalls {
				validToolCallIDs[tc.ID] = true
			}
			sanitized = append(sanitized, msg)

		case core.RoleTool:
			if validToolCallIDs != nil && validToolCallIDs[msg.ToolCallID] {
				sanitized = append(sanitized, msg)
			} else {
				logger.Warn().
					Int("msg_index", i).
					Str("tool_call_id", msg.ToolCallID).
					Interface("valid_ids_in_context", validToolCallIDs).
					Msg("dropping invalid tool message (orphaned or ID mismatch)")
			}

		default:
			sanitized = append(sanitized, msg)
		}
	}
	return sanitized
}
