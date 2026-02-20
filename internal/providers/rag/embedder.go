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
	Shutdown() error
}

type Embedder struct {
	model     DualEncoder
	timeout   time.Duration
	chunkConf ChunkerConfig
}

func NewEmbedder(model DualEncoder) *Embedder {
	return &Embedder{
		model:     model,
		timeout:   embeddingTimeout,
		chunkConf: E5BaseChunkerConfig(),
	}
}

// EncodeQuery encodes the beginning of the text and the ending.
func (e *Embedder) EncodeQuery(ctx context.Context, text string) ([]float32, error) {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	log.FromCtx(ctx).Debug().
		Int("text_len", len(text)).
		Msg("embedding query")

	chunk, err := e.model.EncodeQuery(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	return chunk, nil
}

func (e *Embedder) EncodePassage(ctx context.Context, text string) ([][]float32, error) {
	chunks := ChunkText(text, e.chunkConf)

	log.FromCtx(ctx).Debug().
		Int("chunks", len(chunks)).
		Int("text_len", len(text)).
		Msg("embedding passage")

	embeddings := make([][]float32, 0, len(chunks))
	for _, chunk := range chunks {
		ctx, cancel := context.WithTimeout(ctx, e.timeout)
		emb, err := e.model.EncodePassage(ctx, chunk.Text)
		cancel()

		if err != nil {
			return nil, fmt.Errorf("failed to embed chunk %d: %w", chunk.Index, err)
		}
		embeddings = append(embeddings, emb)
	}

	return embeddings, nil
}
