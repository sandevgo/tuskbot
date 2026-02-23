package llm

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

// NewProvider creates the appropriate AIProvider based on configuration.
func NewProvider(ctx context.Context, cfg core.ProviderConfig) (core.AIProvider, error) {
	provider, model := cfg.GetProvider(), cfg.GetModel()

	log.FromCtx(ctx).Info().
		Str("provider", provider).
		Str("model", model).
		Msg("starting llm provider")

	switch provider {
	case "openai":
		return NewOpenAI(cfg.GetOpenAIAPIKey(), model), nil
	case "anthropic":
		return NewAnthropic(cfg.GetAnthropicAPIKey(), model), nil
	case "openrouter":
		return NewOpenRouter(cfg.GetOpenRouterAPIKey(), model), nil
	case "ollama":
		return NewOllama(cfg.GetOllamaBaseURL(), cfg.GetOllamaAPIKey(), model), nil
	case "custom":
		return NewCustomOpenAI(cfg.GetCustomOpenAIBaseURL(), cfg.GetCustomOpenAIAPIKey(), model), nil
	default:
		return nil, fmt.Errorf("unknown llm provider: %s", provider)
	}
}
