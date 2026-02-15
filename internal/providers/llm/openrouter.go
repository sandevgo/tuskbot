package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sandevgo/tuskbot/internal/core"
)

type OpenRouter struct {
	*OpenAICompatible
}

func NewOpenRouter(apiKey, model string) *OpenRouter {
	return &OpenRouter{
		OpenAICompatible: NewOpenAICompatible(OpenAICompatibleConfig{
			BaseURL:    "https://openrouter.ai/api",
			APIKey:     apiKey,
			Model:      model,
			AuthHeader: "Authorization",
			AuthPrefix: "Bearer ",
			ExtraHeaders: map[string]string{
				"HTTP-Referer": core.TuskRepositoryURL,
				"X-Title":      core.TuskName,
			},
		}),
	}
}

func (o *OpenRouter) Models(ctx context.Context) ([]core.Model, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + o.apiKey,
		"HTTP-Referer":  core.TuskRepositoryURL,
		"X-Title":       core.TuskName,
	}

	resp, err := o.doRequest(ctx, http.MethodGet, "/v1/models", nil, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []core.Model `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return result.Data, nil
}
