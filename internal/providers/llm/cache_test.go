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

func TestAnthropicPayloadPreservesToolConversationState(t *testing.T) {
	req := core.ChatRequest{
		Messages: []core.Message{
			{Role: core.RoleSystem, Content: "system"},
			{
				Role:    core.RoleAssistant,
				Content: "Let me check that.",
				ToolCalls: []core.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Function: core.FunctionCall{
							Name:      "search",
							Arguments: `{"query":"golang"}`,
						},
					},
				},
			},
			{Role: core.RoleTool, ToolCallID: "call_1", Content: "search results"},
		},
	}

	payload := anthropicPayload("claude-test", req)
	messages := payload["messages"].([]map[string]any)

	if len(messages) != 2 {
		t.Fatalf("messages len = %d, want 2", len(messages))
	}

	assistant := messages[0]
	if assistant["role"] != "assistant" {
		t.Fatalf("assistant role = %#v, want %q", assistant["role"], "assistant")
	}
	assistantContent := assistant["content"].([]map[string]any)
	if assistantContent[0]["type"] != "text" {
		t.Fatalf("assistant first block type = %#v, want %q", assistantContent[0]["type"], "text")
	}
	if assistantContent[1]["type"] != "tool_use" {
		t.Fatalf("assistant second block type = %#v, want %q", assistantContent[1]["type"], "tool_use")
	}
	if assistantContent[1]["id"] != "call_1" {
		t.Fatalf("assistant tool_use id = %#v, want %q", assistantContent[1]["id"], "call_1")
	}
	if assistantContent[1]["name"] != "search" {
		t.Fatalf("assistant tool_use name = %#v, want %q", assistantContent[1]["name"], "search")
	}
	if !reflect.DeepEqual(assistantContent[1]["input"], map[string]any{"query": "golang"}) {
		t.Fatalf("assistant tool_use input = %#v, want %#v", assistantContent[1]["input"], map[string]any{"query": "golang"})
	}

	toolResult := messages[1]
	if toolResult["role"] != "user" {
		t.Fatalf("tool result role = %#v, want %q", toolResult["role"], "user")
	}
	toolResultContent := toolResult["content"].([]map[string]any)
	if toolResultContent[0]["type"] != "tool_result" {
		t.Fatalf("tool result block type = %#v, want %q", toolResultContent[0]["type"], "tool_result")
	}
	if toolResultContent[0]["tool_use_id"] != "call_1" {
		t.Fatalf("tool result tool_use_id = %#v, want %q", toolResultContent[0]["tool_use_id"], "call_1")
	}
	if toolResultContent[0]["content"] != "search results" {
		t.Fatalf("tool result content = %#v, want %q", toolResultContent[0]["content"], "search results")
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
	body := provider.extraBody(core.ChatRequest{})

	expected := map[string]any{"keep_alive": "1h"}
	if !reflect.DeepEqual(body, expected) {
		t.Fatalf("extraBody = %#v, want %#v", body, expected)
	}
}

func TestOpenRouterCacheControlBodyForClaude(t *testing.T) {
	provider := NewOpenRouter("test-key", "anthropic/claude-sonnet-4")
	body := provider.extraBody(core.ChatRequest{
		PromptCache: &core.PromptCachePolicy{
			Mode: core.PromptCacheModePrefer,
		},
	})

	expected := map[string]any{
		"cache_control": map[string]any{
			"type": "ephemeral",
		},
	}
	if !reflect.DeepEqual(body, expected) {
		t.Fatalf("extraBody = %#v, want %#v", body, expected)
	}
}

func TestOpenRouterCacheControlBodySkipsUnsupportedModels(t *testing.T) {
	provider := NewOpenRouter("test-key", "openai/gpt-4o-mini")
	body := provider.extraBody(core.ChatRequest{
		PromptCache: &core.PromptCachePolicy{
			Mode: core.PromptCacheModePrefer,
		},
	})

	if body != nil {
		t.Fatalf("extraBody = %#v, want nil", body)
	}
}

func TestOpenRouterCacheControlBodySkipsBypassMode(t *testing.T) {
	provider := NewOpenRouter("test-key", "anthropic/claude-sonnet-4")
	body := provider.extraBody(core.ChatRequest{
		PromptCache: &core.PromptCachePolicy{
			Mode: core.PromptCacheModeBypass,
		},
	})

	if body != nil {
		t.Fatalf("extraBody = %#v, want nil", body)
	}
}
