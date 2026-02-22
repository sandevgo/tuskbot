package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

const ChatTimeout = 2 * time.Minute

type Agent struct {
	appCfg   *config.AppConfig
	ai       core.AIProvider
	mcp      core.MCPServer
	memory   core.Memory
	executor *Executor
}

func NewAgent(
	appCfg *config.AppConfig,
	ai core.AIProvider,
	mcp core.MCPServer,
	memory core.Memory,
	executor *Executor,
) *Agent {
	return &Agent{
		appCfg:   appCfg,
		ai:       ai,
		mcp:      mcp,
		memory:   memory,
		executor: executor,
	}
}

func (a *Agent) Run(ctx context.Context, sessionID string, input string, onUpdate func(core.Message)) (string, error) {
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

	var finalContent string

	// 4. ReAct Loop
	for {
		logger.Debug().
			Str("session_id", sessionID).
			Msg("agent sending request to llm")

		chatCtx, cancel := context.WithTimeout(ctx, ChatTimeout)
		responseMsg, err := a.ai.Chat(chatCtx, messages, tools)
		cancel()

		if err != nil {
			return "", fmt.Errorf("ai chat error: %w", err)
		}

		logger.Debug().
			Str("session_id", sessionID).
			Msg("agent received llm response")

		// Save Assistant Response and update local context
		if err := a.memory.SaveMessage(ctx, sessionID, responseMsg); err != nil {
			return "", fmt.Errorf("failed to save assistant message: %w", err)
		}
		messages = append(messages, responseMsg)

		if onUpdate != nil {
			onUpdate(responseMsg)
		}

		if responseMsg.Content != "" {
			finalContent = responseMsg.Content
		}

		// If no tools are called, we are done
		if len(responseMsg.ToolCalls) == 0 {
			break
		}

		// 5. Execute Tool Calls
		logger.Debug().
			Str("session_id", sessionID).
			Msg("agent called mcp tool")

		toolResults := a.executor.Execute(ctx, responseMsg.ToolCalls)

		for _, toolMsg := range toolResults {
			if err := a.memory.SaveMessage(ctx, sessionID, toolMsg); err != nil {
				return "", fmt.Errorf("failed to save tool message: %w", err)
			}
			messages = append(messages, toolMsg)
		}

		// Update tool set (if model added new tools)
		tools, err = a.mcp.GetTools(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get tools: %w", err)
		}
	}

	return finalContent, nil
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
