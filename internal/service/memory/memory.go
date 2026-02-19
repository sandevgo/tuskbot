package memory

import (
	"context"
	"fmt"
	"strings"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type Memory struct {
	cfg      *config.AppConfig
	msgRepo  core.MessagesRepository
	knowRepo core.KnowledgeRepository
	embedder core.Embedder
	prompter *SysPrompt
}

func NewMemory(
	cfg *config.AppConfig,
	msgRepo core.MessagesRepository,
	knowRepo core.KnowledgeRepository,
	embedder core.Embedder,
	prompter *SysPrompt,
) *Memory {
	return &Memory{
		cfg:      cfg,
		msgRepo:  msgRepo,
		knowRepo: knowRepo,
		embedder: embedder,
		prompter: prompter,
	}
}

func (s *Memory) GetFullContext(ctx context.Context, sessionID, userQuery string) ([]core.Message, error) {
	messages := s.prompter.Build()

	if rag := s.getContext(ctx, sessionID, userQuery); rag != "" {
		messages = append(messages, core.Message{
			Role:    core.RoleSystem,
			Content: rag,
		})
	}

	history, err := s.msgRepo.GetMessages(ctx, sessionID, s.cfg.ContextWindowSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	messages = append(messages, history...)

	return messages, nil
}

// GetContext retrieves relevant knowledge and messages.
func (s *Memory) getContext(ctx context.Context, sessionID, userQuery string) string {
	logger := log.FromCtx(ctx)

	// 1. Generate embedding for the query
	queryVec, err := s.embedder.EncodeQuery(ctx, userQuery)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to embed query for RAG")
		return ""
	}

	// 2. Search Knowledge and Semantic History
	items, err := s.knowRepo.SearchContext(ctx, queryVec, 5, 3)
	if err != nil {
		logger.Error().Err(err).Msg("RAG search failed")
		return ""
	}

	if len(items) == 0 {
		return ""
	}

	// 3. Format the results
	var facts []string
	var semanticHistory []string

	for _, item := range items {
		if item.Type == "fact" {
			facts = append(facts, "- "+item.Content)
		} else {
			semanticHistory = append(semanticHistory, "- "+item.Content)
		}
	}

	var sb strings.Builder

	if len(facts) > 0 {
		sb.WriteString("\n### Relevant Knowledge\n")
		sb.WriteString(strings.Join(facts, "\n"))
		sb.WriteString("\n")
	}

	if len(semanticHistory) > 0 {
		sb.WriteString("\n### Related Past Conversations\n")
		sb.WriteString(strings.Join(semanticHistory, "\n"))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (s *Memory) SaveMessage(ctx context.Context, sessionID string, msg core.Message) error {
	return s.msgRepo.AddMessage(ctx, sessionID, msg)
}
