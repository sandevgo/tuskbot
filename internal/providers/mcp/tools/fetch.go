package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/inbucket/html2text"
	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/retry"
)

const (
	maxResponseSize     = 1 << 20 // 1MB limit
	defaultFetchTimeout = 15 * time.Second
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
	client  *http.Client
	retrier *retry.Retrier
}

func NewFetchWithTimeout(timeout time.Duration, retryCfg *retry.Config) *Fetch {
	if retryCfg == nil {
		retryCfg = retry.NewDefaultConfig()
	}
	return &Fetch{
		client: &http.Client{
			Timeout: timeout,
		},
		retrier: retry.NewRetrier(retryCfg),
	}
}

func NewFetch() *Fetch {
	return NewFetchWithTimeout(defaultFetchTimeout, nil)
}

func (f *Fetch) FetchURL(ctx context.Context, args json.RawMessage) (string, error) {
	var input struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	var body string
	err := f.retrier.Do(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, input.URL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("User-Agent", core.TuskUserAgent)

		resp, err := f.client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to fetch url: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}

		limitedReader := io.LimitReader(resp.Body, maxResponseSize)

		body, err = html2text.FromReader(limitedReader, html2text.Options{
			OmitLinks:    false,
			PrettyTables: true,
		})
		if err != nil {
			return fmt.Errorf("failed to read body: %w", err)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return body, nil
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
