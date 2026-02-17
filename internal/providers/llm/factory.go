package llm

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

// NewProvider creates the appropriate AIProvider based on configuration.
func NewProvider(ctx context.Context, cfg *config.AppConfig) (core.AIProvider, error) {
	log.FromCtx(ctx).Info().
		Str("provider", cfg.Provider).
		Str("model", cfg.Model).
		Msg("starting llm provider")

	switch cfg.Provider {
	case "openai":
		return NewOpenAI(cfg.OpenAIAPIKey, cfg.Model), nil
	case "anthropic":
		return NewAnthropic(cfg.AnthropicAPIKey, cfg.Model), nil
	case "openrouter":
		return NewOpenRouter(cfg.OpenRouterAPIKey, cfg.Model), nil
	case "ollama":
		return NewOllama(cfg.OllamaBaseURL, cfg.OllamaAPIKey, cfg.Model), nil
	case "custom":
		return NewCustomOpenAI(cfg.CustomOpenAIBaseURL, cfg.CustomOpenAIAPIKey, cfg.Model), nil
	default:
		return nil, fmt.Errorf("unknown llm provider: %s", cfg.Provider)
	}
}
