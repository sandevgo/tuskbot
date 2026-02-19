package rag

import (
	"context"
	"fmt"
	"time"

	"github.com/sandevgo/tuskbot/pkg/log"
)

const embeddingTimeout = 60 * time.Second

type DualEncoder interface {
	EncodeQuery(ctx context.Context, text string) ([]float32, error)
	EncodePassage(ctx context.Context, text string) ([]float32, error)
}

type Embedder struct {
	model DualEncoder
}

func NewEmbedder(model DualEncoder) *Embedder {
	return &Embedder{model: model}
}

// EncodeQuery encodes the beginning of the text and the ending.
func (e *Embedder) EncodeQuery(ctx context.Context, text string) ([][]float32, error) {
	ctx, cancel := context.WithTimeout(ctx, embeddingTimeout)
	defer cancel()

	chunks := ChunkText(text, E5BaseChunkerConfig())

	log.FromCtx(ctx).Debug().
		Int("text_len", len(text)).
		Msg("embedding query")

	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks to encode query")
	}

	embeddings := make([][]float32, 0, 1)

	// Encode first chunk
	chunk, err := e.model.EncodeQuery(ctx, chunks[0].Text)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}
	embeddings = append(embeddings, chunk)

	// Encode last chunk
	if len(chunks) > 1 {
		chunk, err = e.model.EncodeQuery(ctx, chunks[len(chunks)-1].Text)
		if err != nil {
			return nil, fmt.Errorf("failed to encode query: %w", err)
		}
		embeddings = append(embeddings, chunk)
	}

	return embeddings, nil
}

func (e *Embedder) EncodePassage(ctx context.Context, text string) ([][]float32, error) {
	chunks := ChunkText(text, E5BaseChunkerConfig())

	log.FromCtx(ctx).Debug().
		Int("chunks", len(chunks)).
		Int("text_len", len(text)).
		Msg("embedding passage")

	embeddings := make([][]float32, 0, len(chunks))
	for _, chunk := range chunks {
		ctx, cancel := context.WithTimeout(ctx, embeddingTimeout)
		emb, err := e.model.EncodePassage(ctx, chunk.Text)
		cancel()

		if err != nil {
			return nil, fmt.Errorf("failed to embed chunk %d: %w", chunk.Index, err)
		}
		embeddings = append(embeddings, emb)
	}

	return embeddings, nil
}
