package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/sandevgo/tuskbot/internal/core"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestAnthropicPayloadAddsCacheMarkersAndTools(t *testing.T) {
	req := core.ChatRequest{
		Messages: []core.Message{
			{Role: core.RoleSystem, Content: "system-a"},
			{Role: core.RoleSystem, Content: "system-b"},
			{Role: core.RoleUser, Content: "hello"},
		},
		Tools: []core.Tool{
			{
				Type: "function",
				Function: core.Function{
					Name:        "search",
					Description: "Search things",
					Parameters:  []byte(`{"type":"object"}`),
				},
			},
		},
		PromptCache: &core.PromptCachePolicy{
			Mode:               core.PromptCacheModePrefer,
			MessageBreakpoints: []int{1, 2},
			IncludeTools:       true,
		},
	}

	payload := anthropicPayload("claude-test", req)

	system := payload["system"].([]map[string]any)
	if _, ok := system[0]["cache_control"]; ok {
		t.Fatal("did not expect cache marker on first system block")
	}
	if _, ok := system[1]["cache_control"]; !ok {
		t.Fatal("expected cache marker on second system block")
	}

	messages := payload["messages"].([]map[string]any)
	content := messages[0]["content"].([]map[string]any)
	if _, ok := content[0]["cache_control"]; !ok {
		t.Fatal("expected cache marker on user message block")
	}

	tools := payload["tools"].([]map[string]any)
	if len(tools) != 1 || tools[0]["name"] != "search" {
		t.Fatalf("unexpected tools payload: %#v", tools)
	}
}

func TestAnthropicPayloadBypassSkipsCacheMarkers(t *testing.T) {
	req := core.ChatRequest{
		Messages: []core.Message{
			{Role: core.RoleSystem, Content: "system"},
			{Role: core.RoleUser, Content: "hello"},
		},
		PromptCache: &core.PromptCachePolicy{
			Mode:               core.PromptCacheModeBypass,
			MessageBreakpoints: []int{0, 1},
		},
	}

	payload := anthropicPayload("claude-test", req)

	system := payload["system"].([]map[string]any)
	if _, ok := system[0]["cache_control"]; ok {
		t.Fatal("did not expect cache marker when bypass is set")
	}

	messages := payload["messages"].([]map[string]any)
	content := messages[0]["content"].([]map[string]any)
	if _, ok := content[0]["cache_control"]; ok {
		t.Fatal("did not expect message cache marker when bypass is set")
	}
}

func TestAnthropicPayloadIncludesToolsWithoutPromptCache(t *testing.T) {
	req := core.ChatRequest{
		Messages: []core.Message{
			{Role: core.RoleUser, Content: "hello"},
		},
		Tools: []core.Tool{
			{
				Type: "function",
				Function: core.Function{
					Name:        "search",
					Description: "Search things",
					Parameters:  []byte(`{"type":"object"}`),
				},
			},
		},
	}

	payload := anthropicPayload("claude-test", req)

	tools, ok := payload["tools"]
	if !ok {
		t.Fatal("expected tools payload even when prompt cache is disabled")
	}

	gotTools := tools.([]map[string]any)
	if len(gotTools) != 1 || gotTools[0]["name"] != "search" {
		t.Fatalf("unexpected tools payload: %#v", gotTools)
	}
}

func TestAnthropicChatParsesToolUseBlocks(t *testing.T) {
	provider := NewAnthropic("test-key", "claude-test")
	provider.client = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body: io.NopCloser(strings.NewReader(`{
			"content": [
				{
					"type": "tool_use",
					"id": "toolu_123",
					"name": "search",
					"input": {"query": "golang"}
				}
			]
		}`)),
				Request: r,
			}, nil
		}),
	}

	resp, err := provider.Chat(context.Background(), core.ChatRequest{
		Messages: []core.Message{
			{Role: core.RoleUser, Content: "find docs"},
		},
		Tools: []core.Tool{
			{
				Type: "function",
				Function: core.Function{
					Name:        "search",
					Description: "Search things",
					Parameters:  []byte(`{"type":"object"}`),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	if len(resp.ToolCalls) != 1 {
		t.Fatalf("ToolCalls len = %d, want 1", len(resp.ToolCalls))
	}

	got := resp.ToolCalls[0]
	if got.ID != "toolu_123" {
		t.Fatalf("ToolCall ID = %q, want %q", got.ID, "toolu_123")
	}
	if got.Type != "function" {
		t.Fatalf("ToolCall Type = %q, want %q", got.Type, "function")
	}
	if got.Function.Name != "search" {
		t.Fatalf("ToolCall Function.Name = %q, want %q", got.Function.Name, "search")
	}
	var args map[string]string
	if err := json.Unmarshal([]byte(got.Function.Arguments), &args); err != nil {
		t.Fatalf("ToolCall Function.Arguments is not valid JSON: %v", err)
	}
	if !reflect.DeepEqual(args, map[string]string{"query": "golang"}) {
		t.Fatalf("ToolCall Function.Arguments = %#v, want %#v", args, map[string]string{"query": "golang"})
	}
}

func TestOllamaKeepAliveBody(t *testing.T) {
	provider := NewOllama("http://localhost:11434", "", "llama3")
	body := provider.extraBody(core.ChatRequest{
		ModelReuse: &core.ModelReusePolicy{
			Mode: core.ModelReuseModePrefer,
			TTL:  "5m",
		},
	})

	expected := map[string]any{"keep_alive": "5m"}
	if !reflect.DeepEqual(body, expected) {
		t.Fatalf("extraBody = %#v, want %#v", body, expected)
	}
}
