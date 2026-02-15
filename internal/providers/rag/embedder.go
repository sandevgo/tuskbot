package rag

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/pkg/llamacpp"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type Embedder struct {
	llm *llamacpp.LlamaEmbedder
}

func NewEmbedder(cfg *config.RAGConfig) (*Embedder, error) {
	llm, err := llamacpp.NewLlamaEmbedder(cfg.ModelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize llama embedder: %w", err)
	}
	return &Embedder{llm: llm}, nil
}

func (e *Embedder) Embed(ctx context.Context, text string) ([][]float32, error) {
	chunks := ChunkText(text, E5BaseChunkerConfig())
	embeddings := make([][]float32, 0, len(chunks))

	log.FromCtx(ctx).Debug().Msgf("embedding %d chunk", len(chunks))
	for _, chunk := range chunks {
		emb, err := e.llm.Embed(chunk.Text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed chunk: %w", err)
		}
		embeddings = append(embeddings, emb)
	}
	return embeddings, nil
}

func (e *Embedder) Shutdown() error {
	e.llm.Free()
	return nil
}
