package llm

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

// NewProvider creates the appropriate AIProvider based on configuration.
func NewProvider(ctx context.Context, appCfg *config.AppConfig) (core.AIProvider, error) {
	log.FromCtx(ctx).Info().
		Str("provider", appCfg.Provider).
		Str("model", appCfg.Model).
		Msg("starting llm provider")

	switch appCfg.Provider {
	case "openrouter":
		return NewOpenRouter(appCfg), nil
	case "openai":
		return nil, fmt.Errorf("openai provider not yet implemented")
	case "ollama":
		return nil, fmt.Errorf("ollama provider not yet implemented")
	default:
		return nil, fmt.Errorf("unknown llm provider: %s", appCfg.Provider)
	}
}
