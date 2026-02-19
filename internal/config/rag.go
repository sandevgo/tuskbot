package config

import (
	"context"

	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog/log"
)

type RAGConfig struct {
	ModelName string `env:"TUSK_EMBEDDING_MODEL,required"`
}

func NewRAGConfig(ctx context.Context) *RAGConfig {
	cfg := &RAGConfig{}
	if err := env.Parse(cfg); err != nil {
		log.Ctx(ctx).Fatal().Err(err).Msg("failed to parse RAG config")
	}
	return cfg
}
