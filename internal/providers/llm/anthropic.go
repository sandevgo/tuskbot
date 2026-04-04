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
			Type  string          `json:"type"`
			Text  string          `json:"text"`
			ID    string          `json:"id"`
			Name  string          `json:"name"`
			Input json.RawMessage `json:"input"`
		} `json:"content"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return core.Message{}, fmt.Errorf("decode: %w", err)
	}

	var text string
	var toolCalls []core.ToolCall
	for _, c := range result.Content {
		switch c.Type {
		case "text":
			text += c.Text
		case "tool_use":
			toolCalls = append(toolCalls, core.ToolCall{
				ID:   c.ID,
				Type: "function",
				Function: core.FunctionCall{
					Name:      c.Name,
					Arguments: string(c.Input),
				},
			})
		}
	}
	return core.Message{
		Role:      core.RoleAssistant,
		Content:   text,
		ToolCalls: toolCalls,
	}, nil
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
		absoluteIdx := systemCount + i
		msg := anthropicMessage(m)
		if _, ok := breakpoints[absoluteIdx]; ok {
			addAnthropicCacheControl(msg)
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
	if len(req.Tools) > 0 {
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

func anthropicMessage(msg core.Message) map[string]any {
	switch msg.Role {
	case core.RoleAssistant:
		content := make([]map[string]any, 0, 1+len(msg.ToolCalls))
		if strings.TrimSpace(msg.Content) != "" {
			content = append(content, textBlock(msg.Content))
		}
		for _, tc := range msg.ToolCalls {
			content = append(content, map[string]any{
				"type":  "tool_use",
				"id":    tc.ID,
				"name":  tc.Function.Name,
				"input": anthropicToolInput(tc.Function.Arguments),
			})
		}
		if len(content) == 0 {
			content = append(content, textBlock(msg.Content))
		}
		return map[string]any{
			"role":    "assistant",
			"content": content,
		}
	case core.RoleTool:
		return map[string]any{
			"role": "user",
			"content": []map[string]any{
				{
					"type":        "tool_result",
					"tool_use_id": msg.ToolCallID,
					"content":     msg.Content,
				},
			},
		}
	default:
		return map[string]any{
			"role":    normalizeAnthropicRole(msg.Role),
			"content": []map[string]any{textBlock(msg.Content)},
		}
	}
}

func addAnthropicCacheControl(msg map[string]any) {
	content, ok := msg["content"].([]map[string]any)
	if !ok || len(content) == 0 {
		return
	}
	content[0]["cache_control"] = map[string]any{"type": "ephemeral"}
}

func anthropicToolInput(arguments string) any {
	if strings.TrimSpace(arguments) == "" {
		return map[string]any{}
	}

	var input any
	if err := json.Unmarshal([]byte(arguments), &input); err != nil {
		return map[string]any{"raw": arguments}
	}

	return input
}

func normalizeAnthropicRole(role string) string {
	switch role {
	case core.RoleAssistant:
		return "assistant"
	case core.RoleUser:
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
