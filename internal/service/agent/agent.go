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
	logger := log.FromCtx(ctx)
	logger.Info().Str("task", task.Name).Msg("1. Starting notification process")

	msg := core.Message{
		Role:    core.RoleSystem,
		Content: fmt.Sprintf("Background Task '%s' completed.\n%s", task.Name, result),
	}
	if err := a.memory.SaveMessage(ctx, task.OwnerSessionID, msg); err != nil {
		return err
	}

	logger.Info().Msg("2. Waiting for session lock")
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

	logger.Info().Msg("3. Loading full context")
	messages, err := a.memory.GetFullContext(ctx, task.OwnerSessionID, "")
	if err != nil {
		return err
	}

	messages = append(messages, core.Message{
		Role:    core.RoleUser,
		Content: "Please inform me about the completed background task and take any necessary follow-up actions.",
	})

	logger.Info().Msg("4. Fetching MCP tools")
	tools, err := a.mcp.GetTools(ctx)
	if err != nil {
		return err
	}

	sanitizedMsgs := sanitizeToolCalls(ctx, messages)
	logger.Info().Int("msg_count", len(sanitizedMsgs)).Msg("5. Calling LLM Runner...")

	final, err := a.runner.Run(ctx, sanitizedMsgs, tools, func(m core.Message) error {
		logger.Info().Str("role", m.Role).Msg("6. Runner yielded a message")
		return a.memory.SaveMessage(ctx, task.OwnerSessionID, m)
	})

	if err != nil {
		logger.Error().Err(err).Msg("Runner returned an error")
		return err
	}

	logger.Info().Msg("7. Runner finished, publishing event")
	a.events.Publish(ctx, core.NewChatEvent(core.EventTypeTaskCompleted, task.OwnerSessionID, final))
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
