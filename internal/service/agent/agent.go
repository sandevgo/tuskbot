package agent

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

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

	// 3. Prepare Tools
	tools, err := a.mcp.GetTools(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get tools: %w", err)
	}

	var finalContent string

	// 4. ReAct Loop
	for {
		responseMsg, err := a.ai.Chat(ctx, messages, tools)
		if err != nil {
			return "", fmt.Errorf("ai chat error: %w", err)
		}

		// Save Assistant Response and update local context
		if err := a.memory.SaveMessage(ctx, sessionID, responseMsg); err != nil {
			logger.Error().Err(err).Msg("failed to save assistant message")
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
		toolResults := a.executor.Execute(ctx, responseMsg.ToolCalls)

		for _, toolMsg := range toolResults {
			if err := a.memory.SaveMessage(ctx, sessionID, toolMsg); err != nil {
				logger.Error().Err(err).Msg("failed to save tool message")
			}
			messages = append(messages, toolMsg)
		}
	}

	return finalContent, nil
}
