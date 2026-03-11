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
}

func NewAgent(
	ai core.AIProvider,
	mcp core.MCPServer,
	memory core.Memory,
	executor core.ToolExecutor,
) *Agent {
	return &Agent{
		runner: NewReActRunner(ai, executor, ChatTimeout),
		mcp:    mcp,
		memory: memory,
	}
}

func (a *Agent) Run(ctx context.Context, sessionID string, input string, onUpdate core.UpdateFunc) (string, error) {
	logger := log.FromCtx(ctx)

	logger.Debug().
		Str("session_id", sessionID).
		Msg("agent received user request")

	// 1. Record the User Input
	userMsg := core.Message{Role: core.RoleUser, Content: input}
	if err := a.memory.SaveMessage(ctx, sessionID, userMsg); err != nil {
		return "", fmt.Errorf("failed to save user message: %w", err)
	}

	// 2. Recall the "State of the World"
	// Memory returns [System Prompt + RAG Context + Chronological History]
	messages, err := a.memory.GetFullContext(ctx, sessionID, input)
	if err != nil {
		return "", fmt.Errorf("failed to get context: %w", err)
	}

	// Sanitize history to prevent provider errors (orphaned tool calls)
	messages = sanitizeToolCalls(ctx, messages)

	// 3. Prepare Tools
	tools, err := a.mcp.GetTools(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get tools: %w", err)
	}

	// 4. ReAct Loop
	finalContent, err := a.runner.Run(ctx, messages, tools, func(msg core.Message) error {
		// Persist every message to memory
		if err := a.memory.SaveMessage(ctx, sessionID, msg); err != nil {
			return fmt.Errorf("failed to save message: %w", err)
		}

		// Notify callback only for assistant messages
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
	// Notify the user that a task completed
	// Could send a message to the session or trigger a callback
	// Implementation depends on transport layer (Telegram, CLI, etc.)
	// TODO: ???
	return nil
}

// sanitizeToolCalls ensures the message history is valid for LLM consumption.
// It removes Tool messages that do not have a corresponding preceding Assistant tool call.
func sanitizeToolCalls(ctx context.Context, messages []core.Message) []core.Message {
	logger := log.FromCtx(ctx)
	var sanitized []core.Message
	var validToolCallIDs map[string]bool

	for i, msg := range messages {
		switch msg.Role {
		case core.RoleUser, core.RoleSystem:
			// User/System messages reset the tool context
			validToolCallIDs = nil
			sanitized = append(sanitized, msg)

		case core.RoleAssistant:
			// Assistant message establishes new tool context
			validToolCallIDs = make(map[string]bool)
			for _, tc := range msg.ToolCalls {
				validToolCallIDs[tc.ID] = true
			}
			sanitized = append(sanitized, msg)

		case core.RoleTool:
			// Tool message must match a valid ID from the immediate preceding assistant turn
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
			// Keep other message types
			sanitized = append(sanitized, msg)
		}
	}
	return sanitized
}
