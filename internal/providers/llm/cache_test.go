package llm

import (
	"reflect"
	"testing"

	"github.com/sandevgo/tuskbot/internal/core"
)

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
