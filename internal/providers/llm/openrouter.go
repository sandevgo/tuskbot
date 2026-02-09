package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/core"
)

const (
	openRouterURL       = "https://openrouter.ai/api/v1/chat/completions"
	openRouterModelsURL = "https://openrouter.ai/api/v1/models"
)

type OpenRouter struct {
	cfg    *config.AppConfig
	client *http.Client
}

func NewOpenRouter(cfg *config.AppConfig) *OpenRouter {
	return &OpenRouter{
		cfg:    cfg,
		client: &http.Client{},
	}
}

type chatRequest struct {
	Model    string         `json:"model"`
	Messages []core.Message `json:"messages"`
	Tools    []core.Tool    `json:"tools,omitempty"`
}

type choice struct {
	Message core.Message `json:"message"`
}

type chatResponse struct {
	Choices []choice `json:"choices"`
	Error   struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type ModelInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ContextLength int    `json:"context_length"`
}

type modelsResponse struct {
	Data []ModelInfo `json:"data"`
}

func (o *OpenRouter) Chat(ctx context.Context, history []core.Message, tools []core.Tool) (core.Message, error) {
	reqBody := chatRequest{
		Model:    o.cfg.Model,
		Messages: history,
		Tools:    tools,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return core.Message{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return core.Message{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+o.cfg.OpenRouterAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", core.TuskRepositoryURL)
	req.Header.Set("X-Title", core.TuskName)

	resp, err := o.client.Do(req)
	if err != nil {
		return core.Message{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return core.Message{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return core.Message{}, fmt.Errorf("api returned status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return core.Message{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		if chatResp.Error.Message != "" {
			return core.Message{}, fmt.Errorf("api error: %s", chatResp.Error.Message)
		}
		return core.Message{}, fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message, nil
}

func (o *OpenRouter) GetModels(ctx context.Context) ([]ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, openRouterModelsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+o.cfg.OpenRouterAPIKey)
	req.Header.Set("HTTP-Referer", core.TuskRepositoryURL)
	req.Header.Set("X-Title", core.TuskName)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api returned status %d: %s", resp.StatusCode, string(body))
	}

	var res modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return res.Data, nil
}
