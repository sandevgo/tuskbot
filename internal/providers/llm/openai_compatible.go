package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sandevgo/tuskbot/internal/core"
)

type OpenAICompatible struct {
	baseProvider
	authHeader   string
	authPrefix   string
	extraHeaders map[string]string
}

type OpenAICompatibleConfig struct {
	BaseURL      string
	APIKey       string
	Model        string
	AuthHeader   string // e.g., "Authorization"
	AuthPrefix   string // e.g., "Bearer "
	ExtraHeaders map[string]string
}

func NewOpenAICompatible(cfg OpenAICompatibleConfig) *OpenAICompatible {
	return &OpenAICompatible{
		baseProvider: newBaseProvider(cfg.BaseURL, cfg.APIKey, cfg.Model),
		authHeader:   cfg.AuthHeader,
		authPrefix:   cfg.AuthPrefix,
		extraHeaders: cfg.ExtraHeaders,
	}
}

func (o *OpenAICompatible) Chat(ctx context.Context, history []core.Message, tools []core.Tool) (core.Message, error) {
	payload := map[string]any{
		"model":    o.model,
		"messages": history,
	}
	if len(tools) > 0 {
		payload["tools"] = tools
	}

	headers := make(map[string]string)
	if o.authHeader != "" && o.apiKey != "" {
		headers[o.authHeader] = o.authPrefix + o.apiKey
	}
	for k, v := range o.extraHeaders {
		headers[k] = v
	}

	resp, err := o.doRequest(ctx, http.MethodPost, "/v1/chat/completions", payload, headers)
	if err != nil {
		return core.Message{}, err
	}
	defer resp.Body.Close()

	return parseOpenAIResponse(resp)
}

func parseOpenAIResponse(resp *http.Response) (core.Message, error) {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return core.Message{}, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return core.Message{}, fmt.Errorf("http %d: %s", resp.StatusCode, string(data))
	}

	var result struct {
		Choices []struct {
			Message core.Message `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return core.Message{}, fmt.Errorf("decode: %w", err)
	}
	if len(result.Choices) == 0 {
		return core.Message{}, fmt.Errorf("empty choices: %s", string(data))
	}
	return result.Choices[0].Message, nil
}
