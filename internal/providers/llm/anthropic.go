package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

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

func (a *Anthropic) Chat(ctx context.Context, req core.ChatRequest) (core.Message, error) {
	payload := anthropicPayload(a.model, req)

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

func (a *Anthropic) Capabilities() core.ProviderCapabilities {
	return core.ProviderCapabilities{
		PromptCache: core.PromptCacheSupportExplicit,
	}
}

func anthropicPayload(model string, req core.ChatRequest) map[string]any {
	systemCount := 0
	for _, msg := range req.Messages {
		if msg.Role != core.RoleSystem {
			break
		}
		systemCount++
	}

	breakpoints := map[int]struct{}{}
	if req.PromptCache != nil && req.PromptCache.Mode != core.PromptCacheModeBypass {
		for _, idx := range req.PromptCache.MessageBreakpoints {
			breakpoints[idx] = struct{}{}
		}
	}

	systemBlocks := make([]map[string]any, 0, systemCount)
	for i := 0; i < systemCount; i++ {
		block := textBlock(req.Messages[i].Content)
		if _, ok := breakpoints[i]; ok {
			block["cache_control"] = map[string]any{"type": "ephemeral"}
		}
		systemBlocks = append(systemBlocks, block)
	}

	messages := make([]map[string]any, 0, len(req.Messages)-systemCount)
	for i, m := range req.Messages[systemCount:] {
		role := normalizeAnthropicRole(m.Role)
		block := textBlock(m.Content)
		absoluteIdx := systemCount + i
		if _, ok := breakpoints[absoluteIdx]; ok {
			block["cache_control"] = map[string]any{"type": "ephemeral"}
		}

		msg := map[string]any{
			"role":    role,
			"content": []map[string]any{block},
		}
		messages = append(messages, msg)
	}

	payload := map[string]any{
		"model":      model,
		"max_tokens": 4096,
		"messages":   messages,
	}
	if len(systemBlocks) > 0 {
		payload["system"] = systemBlocks
	}
	if req.PromptCache != nil && req.PromptCache.IncludeTools && len(req.Tools) > 0 {
		payload["tools"] = anthropicTools(req.Tools)
	}

	return payload
}

func anthropicTools(tools []core.Tool) []map[string]any {
	result := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		result = append(result, map[string]any{
			"name":         tool.Function.Name,
			"description":  tool.Function.Description,
			"input_schema": tool.Function.Parameters,
		})
	}
	return result
}

func normalizeAnthropicRole(role string) string {
	switch role {
	case core.RoleAssistant:
		return "assistant"
	case core.RoleUser:
		return "user"
	case core.RoleTool:
		return "user"
	default:
		return "user"
	}
}

func textBlock(content string) map[string]any {
	return map[string]any{
		"type": "text",
		"text": strings.TrimSpace(content),
	}
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
