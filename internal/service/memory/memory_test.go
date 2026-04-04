package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
)

type stubAppConfig struct{}

func (stubAppConfig) GetRuntimePath() string    { return "" }
func (stubAppConfig) GetConfigPath() string     { return "" }
func (stubAppConfig) GetModelsPath() string     { return "" }
func (stubAppConfig) GetPromptPath() string     { return "" }
func (stubAppConfig) GetDatabasePath() string   { return "" }
func (stubAppConfig) GetMCPConfigPath() string  { return "" }
func (stubAppConfig) GetContextWindowSize() int { return 10 }
func (stubAppConfig) IsTelegramSelected() bool  { return false }

type stubPromptConfig struct {
	system   string
	identity string
	user     string
	memory   string
	subagent string
}

func (s stubPromptConfig) GetSystemPath() string      { return s.system }
func (s stubPromptConfig) GetIdentityPath() string    { return s.identity }
func (s stubPromptConfig) GetUserProfilePath() string { return s.user }
func (s stubPromptConfig) GetMemoryPath() string      { return s.memory }
func (s stubPromptConfig) GetSubAgentPath() string    { return s.subagent }

type stubMessagesRepo struct {
	history []core.Message
}

func (s stubMessagesRepo) AddMessage(ctx context.Context, sessionID string, msg core.Message) error {
	return nil
}

func (s stubMessagesRepo) GetMessages(ctx context.Context, sessionID string, limit int) ([]core.Message, error) {
	return s.history, nil
}

func (s stubMessagesRepo) GetUnembeddedMessages(ctx context.Context, limit int) ([]core.StoredMessage, error) {
	return nil, nil
}

func (s stubMessagesRepo) UpdateMessageEmbedding(ctx context.Context, id int64, embedding []float32) error {
	return nil
}

type stubKnowledgeRepo struct{}

func (stubKnowledgeRepo) SaveFact(ctx context.Context, fact core.StoredKnowledge) error { return nil }

func (stubKnowledgeRepo) SearchContext(ctx context.Context, vector []float32, limitKnowledge, limitHistory int) ([]core.ContextItem, error) {
	return nil, nil
}

func (stubKnowledgeRepo) MarkMessagesExtracted(ctx context.Context, messageIDs []int64) error {
	return nil
}

func (stubKnowledgeRepo) GetUnextractedMessages(ctx context.Context, limit int) ([]core.StoredMessage, error) {
	return nil, nil
}

func (stubKnowledgeRepo) GetRecentExtractedMessages(ctx context.Context, limit int, before time.Time, threshold time.Duration) ([]core.StoredMessage, error) {
	return nil, nil
}

type stubEmbedder struct{}

func (stubEmbedder) EncodeQuery(ctx context.Context, text string) ([]float32, error) {
	return []float32{1}, nil
}

func (stubEmbedder) EncodePassage(ctx context.Context, text string) ([][]float32, error) {
	return nil, nil
}

type stubCounter struct{}

func (stubCounter) CountTokens(text string) int { return len(text) }

func TestGetFullContextReturnsPromptContextWithStaticPrefix(t *testing.T) {
	dir := t.TempDir()
	cfg := stubPromptConfig{
		system:   writePromptFile(t, dir, "SYSTEM.md", "system %s"),
		identity: writePromptFile(t, dir, "IDENTITY.md", "identity"),
		user:     "",
		memory:   "",
		subagent: "",
	}

	mem := &Memory{
		cfg:          stubAppConfig{},
		msgRepo:      stubMessagesRepo{history: []core.Message{{Role: core.RoleUser, Content: "history"}}},
		knowRepo:     stubKnowledgeRepo{},
		embedder:     stubEmbedder{},
		tokenCounter: stubCounter{},
		prompter:     NewSysPrompt(dir, cfg),
	}

	promptCtx, err := mem.GetFullContext(context.Background(), "s1", "hello")
	if err != nil {
		t.Fatal(err)
	}

	if promptCtx.StaticPrefixCount != 2 {
		t.Fatalf("StaticPrefixCount = %d, want 2", promptCtx.StaticPrefixCount)
	}
	if len(promptCtx.Messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(promptCtx.Messages))
	}
}

func writePromptFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	return path
}
