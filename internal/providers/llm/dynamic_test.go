package llm

import (
	"context"
	"strings"
	"testing"
)

type testProviderConfig struct {
	model string
}

func (c *testProviderConfig) GetModel() string {
	return c.model
}

func (c *testProviderConfig) SetModel(model string) error {
	c.model = model
	return nil
}

func (c *testProviderConfig) GetProvider() string {
	provider, _ := parseProviderModel(c.model)
	return provider
}

func (c *testProviderConfig) GetAnthropicAPIKey() string {
	return ""
}

func (c *testProviderConfig) GetOpenAIAPIKey() string {
	return ""
}

func (c *testProviderConfig) GetOpenRouterAPIKey() string {
	return ""
}

func (c *testProviderConfig) GetOllamaAPIKey() string {
	return ""
}

func (c *testProviderConfig) GetOllamaBaseURL() string {
	return "http://localhost:11434"
}

func (c *testProviderConfig) GetCustomOpenAIBaseURL() string {
	return ""
}

func (c *testProviderConfig) GetCustomOpenAIAPIKey() string {
	return ""
}

func parseProviderModel(model string) (string, string) {
	parts := strings.SplitN(model, "/", 2)
	if len(parts) != 2 {
		return "", model
	}

	return parts[0], parts[1]
}

func TestDynamicProviderSetModelSwitchesProviderTypes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := &testProviderConfig{model: "openai/gpt-4o-mini"}

	provider, err := NewDynamicProvider(ctx, cfg)
	if err != nil {
		t.Fatalf("NewDynamicProvider() error = %v", err)
	}

	testCases := []struct {
		name  string
		model string
	}{
		{
			name:  "switch to ollama",
			model: "ollama/gemma4:e4",
		},
		{
			name:  "switch back to openai",
			model: "openai/gpt-4o-mini",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			if err := provider.SetModel(ctx, tc.model); err != nil {
				t.Fatalf("SetModel() error = %v", err)
			}

			if got := provider.GetModel(); got != tc.model {
				t.Fatalf("GetModel() = %q, want %q", got, tc.model)
			}
		})
	}
}
