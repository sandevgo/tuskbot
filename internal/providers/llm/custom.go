package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sandevgo/tuskbot/internal/core"
)

type CustomOpenAI struct {
	*OpenAICompatible
}

func NewCustomOpenAI(baseURL, apiKey, model string) *CustomOpenAI {
	return &CustomOpenAI{
		OpenAICompatible: NewOpenAICompatible(OpenAICompatibleConfig{
			BaseURL:    baseURL,
			APIKey:     apiKey,
			Model:      model,
			AuthHeader: "Authorization",
			AuthPrefix: "Bearer ",
		}),
	}
}

func (c *CustomOpenAI) Models(ctx context.Context) ([]core.Model, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	resp, err := c.doRequest(ctx, http.MethodGet, "/v1/models", nil, headers)
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
			ID string `json:"id"`
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
