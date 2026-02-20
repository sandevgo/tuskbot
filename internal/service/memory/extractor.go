package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

const (
	defaultBatchSize          = 100
	defaultExtractionInterval = 30 * time.Minute
	defaultSessionGap         = 30 * time.Minute
	defaultCommitTimeout      = 5 * time.Minute
	windowSize                = 20
	windowOverlap             = 5
)

type Extractor struct {
	repo                core.KnowledgeRepository
	ai                  core.AIProvider
	embedder            core.Embedder
	Interval            time.Duration
	BatchSize           int
	ContextGapThreshold time.Duration
}

func NewExtractor(repo core.KnowledgeRepository, ai core.AIProvider, embedder core.Embedder) *Extractor {
	return &Extractor{
		repo:                repo,
		ai:                  ai,
		embedder:            embedder,
		Interval:            defaultExtractionInterval,
		BatchSize:           defaultBatchSize,
		ContextGapThreshold: defaultSessionGap,
	}
}

func (e *Extractor) Start(ctx context.Context) error {
	logger := log.FromCtx(ctx)
	logger.Info().Msg("starting knowledge extractor")

	ticker := time.NewTicker(e.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := e.processBatch(ctx); err != nil {
				logger.Error().Err(err).Msg("batch processing failed")
			}
		}
	}
}

func (e *Extractor) Shutdown(ctx context.Context) error {
	return nil
}

func (e *Extractor) processBatch(ctx context.Context) error {
	unextracted, err := e.repo.GetUnextractedMessages(ctx, e.BatchSize)
	if err != nil {
		return fmt.Errorf("fetch messages: %w", err)
	}
	if len(unextracted) == 0 {
		return nil
	}

	contextMsgs, err := e.repo.GetRecentExtractedMessages(ctx, 5, unextracted[0].CreatedAt, e.ContextGapThreshold)
	if err != nil {
		return fmt.Errorf("fetch context messages: %w", err)
	}

	allMsgs := append(contextMsgs, unextracted...)
	unextractedIDs := make(map[int64]struct{}, len(unextracted))
	for _, m := range unextracted {
		unextractedIDs[m.ID] = struct{}{}
	}

	sessions := splitByContextSessions(allMsgs, e.ContextGapThreshold)

	for _, session := range sessions {
		if err := e.processSessionWindows(ctx, session, unextractedIDs); err != nil {
			return err
		}
	}

	return nil
}

func (e *Extractor) processSessionWindows(ctx context.Context, session []core.StoredMessage, unextractedIDs map[int64]struct{}) error {
	if len(session) == 0 {
		return nil
	}

	windows := createSlidingWindows(session, windowSize, windowOverlap)

	for i, window := range windows {
		isLastWindow := i == len(windows)-1

		if isLastWindow && len(window) < windowSize {
			lastMsg := window[len(window)-1]
			if time.Since(lastMsg.CreatedAt) < defaultCommitTimeout {
				continue
			}
		}

		// zombie sessions
		if !hasUnextracted(window, unextractedIDs) {
			continue
		}

		if err := e.processWindow(ctx, window, unextractedIDs); err != nil {
			return err
		}
	}

	return nil
}

func createSlidingWindows(msgs []core.StoredMessage, size, overlap int) [][]core.StoredMessage {
	if len(msgs) == 0 {
		return nil
	}

	step := size - overlap
	var windows [][]core.StoredMessage

	for i := 0; i < len(msgs); i += step {
		end := i + size
		if end > len(msgs) {
			end = len(msgs)
		}

		window := make([]core.StoredMessage, end-i)
		copy(window, msgs[i:end])
		windows = append(windows, window)

		if end == len(msgs) {
			break
		}
	}

	return windows
}

func (e *Extractor) processWindow(ctx context.Context, window []core.StoredMessage, unextractedIDs map[int64]struct{}) error {
	conversation, allIDs := formatConversation(window)

	logger := log.FromCtx(ctx)
	logger.Debug().Int("count", len(window)).Msg("extracting knowledge from window")

	facts, err := e.extractFacts(ctx, conversation)
	if err != nil {
		logger.Error().Err(err).Str("content", conversation).Msg("extraction failed")
		return fmt.Errorf("extraction failed: %w", err)
	}

	if err = e.persistFacts(ctx, facts); err != nil {
		return err
	}

	return e.markUnextracted(ctx, allIDs, unextractedIDs)
}

func (e *Extractor) markUnextracted(ctx context.Context, ids []int64, unextractedIDs map[int64]struct{}) error {
	toMark := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := unextractedIDs[id]; ok {
			toMark = append(toMark, id)
		}
	}

	if len(toMark) == 0 {
		return nil
	}

	if err := e.repo.MarkMessagesExtracted(ctx, toMark); err != nil {
		return fmt.Errorf("mark extracted: %w", err)
	}
	return nil
}

func (e *Extractor) extractFacts(ctx context.Context, conversation string) ([]extractedFact, error) {
	const systemPrompt = "You are a knowledge extraction system. Output only valid JSON."
	userPrompt := buildExtractionPrompt(conversation)

	resp, err := e.ai.Chat(ctx, []core.Message{
		{Role: core.RoleSystem, Content: systemPrompt},
		{Role: core.RoleUser, Content: userPrompt},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("llm chat: %w", err)
	}

	return parseExtractionResponse(resp.Content)
}

func (e *Extractor) persistFacts(ctx context.Context, facts []extractedFact) error {
	logger := log.FromCtx(ctx)

	for _, f := range facts {
		if err := e.saveFact(ctx, f); err != nil {
			if isDuplicateError(err) {
				continue
			}
			return fmt.Errorf("failed to save fact '%s': %w", f.Fact, err)
		}
		logger.Info().Str("category", f.Category).Msg("knowledge extracted")
	}
	return nil
}

func (e *Extractor) saveFact(ctx context.Context, fact extractedFact) error {
	chunks, err := e.embedder.EncodePassage(ctx, fact.Fact)
	if err != nil {
		return fmt.Errorf("embed: %w", err)
	}
	if len(chunks) == 0 {
		return fmt.Errorf("empty embedding")
	}

	stored := core.StoredKnowledge{
		Fact:      fact.Fact,
		Category:  fact.Category,
		Source:    "extracted",
		Embedding: chunks[0],
	}

	if err := e.repo.SaveFact(ctx, stored); err != nil {
		if isDuplicateError(err) {
			return nil
		}
		return fmt.Errorf("save: %w", err)
	}
	return nil
}

func splitByContextSessions(msgs []core.StoredMessage, threshold time.Duration) [][]core.StoredMessage {
	if len(msgs) == 0 {
		return nil
	}

	var groups [][]core.StoredMessage
	currentGroup := []core.StoredMessage{msgs[0]}

	for i := 1; i < len(msgs); i++ {
		if msgs[i].CreatedAt.Sub(msgs[i-1].CreatedAt) > threshold {
			groups = append(groups, currentGroup)
			currentGroup = []core.StoredMessage{}
		}
		currentGroup = append(currentGroup, msgs[i])
	}

	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}

	return groups
}

func hasUnextracted(window []core.StoredMessage, unextractedIDs map[int64]struct{}) bool {
	for _, msg := range window {
		if _, isNew := unextractedIDs[msg.ID]; isNew {
			return true
		}
	}
	return false
}

func formatConversation(msgs []core.StoredMessage) (string, []int64) {
	var b strings.Builder
	ids := make([]int64, 0, len(msgs))

	for _, m := range msgs {
		if m.Role == core.RoleTool || m.Role == core.RoleSystem || strings.HasPrefix(m.Content, "Tool Call") {
			continue
		}

		b.WriteString(strings.ToUpper(m.Role))
		b.WriteString(": ")
		b.WriteString(m.Content)
		b.WriteByte('\n')
		ids = append(ids, m.ID)
	}

	return b.String(), ids
}

type extractedFact struct {
	Fact     string `json:"fact"`
	Category string `json:"category"`
}

func buildExtractionPrompt(conversation string) string {
	return fmt.Sprintf(
		`Extract distinct, permanent facts from the conversation. Output format: JSON list of objects {fact, category}. Categories: [preference, user_fact, project, instruction]. Rules: 1. Ignore greetings and small talk. 2. Facts must be self-contained (replace "he" with "User"). Conversation: %s`,
		conversation,
	)
}

func parseExtractionResponse(content string) ([]extractedFact, error) {
	jsonStr := extractJSONArray(content)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON array found in response")
	}

	var facts []extractedFact
	if err := json.Unmarshal([]byte(jsonStr), &facts); err != nil {
		return nil, fmt.Errorf("unmarshal facts: %w", err)
	}

	return facts, nil
}

func extractJSONArray(content string) string {
	start := strings.Index(content, "[")
	if start == -1 {
		return ""
	}

	end := strings.LastIndex(content[start:], "]")
	if end == -1 {
		return ""
	}

	return content[start : start+end+1]
}

func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint failed") ||
		strings.Contains(msg, "constraint failed")
}
