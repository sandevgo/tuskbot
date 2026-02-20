package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
)

const fetchURLSchema = `
{
  "type": "object",
  "properties": {
    "url": { "type": "string", "description": "The URL to fetch" }
  },
  "required": ["url"]
}
`

type Fetch struct {
	client *http.Client
}

func NewFetch() *Fetch {
	return &Fetch{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (f *Fetch) FetchURL(ctx context.Context, args json.RawMessage) (string, error) {
	var input struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, input.URL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Mimic a browser to avoid some basic blocking
	req.Header.Set("User-Agent", core.TuskUserAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch url: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	return string(body), nil
}

func (f *Fetch) GetDefinitions() map[string]struct {
	Description string
	Schema      string
	Handler     func(context.Context, json.RawMessage) (string, error)
} {
	return map[string]struct {
		Description string
		Schema      string
		Handler     func(context.Context, json.RawMessage) (string, error)
	}{
		"fetch_url": {"Fetch content from a URL (HTTP GET)", fetchURLSchema, f.FetchURL},
	}
}
