package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sandevgo/tuskbot/internal/core"
)

type Anthropic struct {
	baseProvider
}

func NewAnthropic(apiKey, model string) *Anthropic {
	return &Anthropic{
		baseProvider: newBaseProvider("https://api.anthropic.com", apiKey, model),
	}
}

func (a *Anthropic) Chat(ctx context.Context, history []core.Message, tools []core.Tool) (core.Message, error) {
	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	var messages []msg
	for _, m := range history {
		if m.Role == core.RoleSystem {
			continue
		}
		messages = append(messages, msg{Role: m.Role, Content: m.Content})
	}

	payload := map[string]any{
		"model":      a.model,
		"max_tokens": 4096,
		"messages":   messages,
	}

	headers := map[string]string{
		"x-api-key":         a.apiKey,
		"anthropic-version": "2023-06-01",
	}

	resp, err := a.doRequest(ctx, http.MethodPost, "/v1/messages", payload, headers)
	if err != nil {
		return core.Message{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return core.Message{}, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return core.Message{}, fmt.Errorf("http %d: %s", resp.StatusCode, string(data))
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return core.Message{}, fmt.Errorf("decode: %w", err)
	}

	var text string
	for _, c := range result.Content {
		if c.Type == "text" {
			text += c.Text
		}
	}
	return core.Message{Role: core.RoleAssistant, Content: text}, nil
}

func (a *Anthropic) Models(ctx context.Context) ([]core.Model, error) {
	headers := map[string]string{
		"x-api-key":         a.apiKey,
		"anthropic-version": "2023-06-01",
	}

	var models []core.Model
	afterID := ""

	for {
		path := "/v1/models?limit=1000"
		if afterID != "" {
			path = fmt.Sprintf("%s&after_id=%s", path, url.QueryEscape(afterID))
		}

		resp, err := a.doRequest(ctx, http.MethodGet, path, nil, headers)
		if err != nil {
			return nil, err
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(data))
		}

		var result struct {
			Data []struct {
				ID          string `json:"id"`
				DisplayName string `json:"display_name"`
				Type        string `json:"type"`
			} `json:"data"`
			HasMore bool   `json:"has_more"`
			LastID  string `json:"last_id"`
		}

		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("decode: %w", err)
		}

		for _, m := range result.Data {
			if m.Type == "model" {
				models = append(models, core.Model{
					ID:   m.ID,
					Name: m.DisplayName,
					// ContextLength is not provided by the Anthropic models API
				})
			}
		}

		if !result.HasMore {
			break
		}
		afterID = result.LastID
	}

	return models, nil
}
