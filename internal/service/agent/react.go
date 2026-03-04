package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
)

// ReActRunner executes the ReAct (Reasoning + Acting) loop
type ReActRunner struct {
	ai       core.AIProvider
	executor core.ToolExecutor
	timeout  time.Duration
}

// NewReActRunner creates a new ReActRunner with the given dependencies
func NewReActRunner(ai core.AIProvider, executor core.ToolExecutor, timeout time.Duration) *ReActRunner {
	return &ReActRunner{
		ai:       ai,
		executor: executor,
		timeout:  timeout,
	}
}

// Run executes the ReAct loop until no more tool calls are requested.
// onTurn is called for every message (assistant and tool) to allow persistence and callbacks.
func (r *ReActRunner) Run(
	ctx context.Context,
	messages []core.Message,
	tools []core.Tool,
	onTurn func(core.Message) error,
) (string, error) {
	var finalContent string

	for {
		chatCtx, cancel := context.WithTimeout(ctx, r.timeout)
		response, err := r.ai.Chat(chatCtx, messages, tools)
		cancel()

		if err != nil {
			return "", fmt.Errorf("ai chat: %w", err)
		}

		// Append assistant message and notify
		messages = append(messages, response)
		if err := onTurn(response); err != nil {
			return "", err
		}

		if response.Content != "" {
			finalContent = response.Content
		}

		// If no tools are called, we are done
		if len(response.ToolCalls) == 0 {
			break
		}

		// Execute tools and append results
		toolResults := r.executor.Execute(ctx, response.ToolCalls)
		for _, msg := range toolResults {
			messages = append(messages, msg)
			if err := onTurn(msg); err != nil {
				return "", err
			}
		}
	}

	return finalContent, nil
}
