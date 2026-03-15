package rag

import (
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/llamacpp"
)

func NewEmbeddingModel(cfg core.EmbeddingConfig) (DualEncoder, error) {
	llamaEmb, err := llamacpp.NewLlamaEmbedder(cfg.GetEmbeddingModelPath())
	if err != nil {
		return nil, fmt.Errorf("failed to load embedding model: %w", err)
	}

	switch cfg.GetEmbeddingModelName() {
	case ModelNameE5BaseQ8:
		return NewE5BaseModel(llamaEmb), nil
	default:
		return nil, fmt.Errorf("unknown model name: %s", cfg.GetEmbeddingModelName())
	}
}
