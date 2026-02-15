package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
)

type Ollama struct {
	*OpenAICompatible
}

func NewOllama(baseURL, apiKey, model string) *Ollama {
	return &Ollama{
		OpenAICompatible: NewOpenAICompatible(OpenAICompatibleConfig{
			BaseURL:    baseURL,
			APIKey:     apiKey,
			Model:      model,
			AuthHeader: "Authorization",
			AuthPrefix: "Bearer ",
		}),
	}
}

func (o *Ollama) Models(ctx context.Context) ([]core.Model, error) {
	type ollamaTag struct {
		Name string `json:"name"`
	}
	type ollamaResponse struct {
		Models []ollamaTag `json:"models"`
	}

	req, err := http.NewRequestWithContext(ctx, "GET", o.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	if o.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.apiKey)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama not available: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]core.Model, 0, len(result.Models))
	for _, m := range result.Models {
		models = append(models, core.Model{
			ID:            m.Name,
			Name:          m.Name,
			ContextLength: 32768,
		})
	}
	return models, nil
}
