package rag

import (
	"fmt"
	"path/filepath"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/pkg/llamacpp"
)

func NewEmbeddingModel(cfg *config.RAGConfig) (DualEncoder, error) {
	modelPath := filepath.Join(config.GetRuntimePath(), "models", cfg.ModelName)

	llamaEmb, err := llamacpp.NewLlamaEmbedder(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedding model: %w", err)
	}

	switch cfg.ModelName {
	case ModelNameE5BaseQ8:
		return NewE5BaseModel(llamaEmb), nil
	default:
		return nil, fmt.Errorf("unknown model name: %s", cfg.ModelName)
	}
}
