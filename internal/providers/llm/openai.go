package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sandevgo/tuskbot/internal/core"
)

// OpenAI provider is implemented using OpenAICompatible.
type OpenAI struct {
	*OpenAICompatible
}

// NewOpenAI creates a new OpenAI provider.
func NewOpenAI(apiKey, model string) *OpenAI {
	return &OpenAI{
		OpenAICompatible: NewOpenAICompatible(OpenAICompatibleConfig{
			BaseURL:    "https://api.openai.com",
			APIKey:     apiKey,
			Model:      model,
			AuthHeader: "Authorization",
			AuthPrefix: "Bearer ",
		}),
	}
}

func (o *OpenAI) Models(ctx context.Context) ([]core.Model, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + o.apiKey,
	}

	resp, err := o.doRequest(ctx, "GET", "/v1/models", nil, headers)
	if err != nil {
		return nil, fmt.Errorf("fetch models: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(data))
	}

	var apiResp struct {
		Data []struct {
			ID      string `json:"id"`
			Created int64  `json:"created"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &apiResp); err != nil {
		return nil, fmt.Errorf("decode models response: %w", err)
	}

	models := make([]core.Model, 0, len(apiResp.Data))
	for _, m := range apiResp.Data {
		models = append(models, core.Model{
			ID:   m.ID,
			Name: m.ID,
		})
	}

	return models, nil
}
