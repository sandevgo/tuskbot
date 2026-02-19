package memory

import (
	"context"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

const (
	EmbedderBatchSize    = 30
	EmbedderPollInterval = 5 * time.Second
)

type EmbedderWorker struct {
	repo      core.MessagesRepository
	embedder  core.Embedder
	interval  time.Duration
	batchSize int
}

func NewEmbedderWorker(repo core.MessagesRepository, embedder core.Embedder) *EmbedderWorker {
	return &EmbedderWorker{
		repo:      repo,
		embedder:  embedder,
		interval:  EmbedderPollInterval,
		batchSize: EmbedderBatchSize,
	}
}

func (w *EmbedderWorker) Start(ctx context.Context) error {
	logger := log.FromCtx(ctx).With().Str("component", "embedder_worker").Logger()
	logger.Info().Msg("starting embedding worker")

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("shutting down embedding worker")
			return nil
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				logger.Error().Err(err).Msg("embedding batch failed")
			}
		}
	}
}

func (w *EmbedderWorker) Shutdown(ctx context.Context) error {
	return nil
}

func (w *EmbedderWorker) processBatch(ctx context.Context) error {
	logger := log.FromCtx(ctx)

	msgs, err := w.repo.GetUnembeddedMessages(ctx, w.batchSize)
	if err != nil {
		return err
	}

	if len(msgs) == 0 {
		return nil
	}

	for _, msg := range msgs {
		if msg.Content == "" {
			continue
		}

		chunks, err := w.embedder.EncodePassage(ctx, msg.Content)
		if err != nil {
			logger.Warn().
				Err(err).
				Int64("msg_id", msg.ID).
				Msg("failed to embed message")
			continue
		}

		if len(chunks) == 0 {
			logger.Warn().Int64("msg_id", msg.ID).Msg("no embeddings generated for message")
			continue
		}

		if err := w.repo.UpdateMessageEmbedding(ctx, msg.ID, chunks[0]); err != nil {
			logger.Error().
				Err(err).
				Int64("msg_id", msg.ID).
				Msg("failed to save embedding")
		}
	}

	return nil
}
