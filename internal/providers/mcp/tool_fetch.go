package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/retry"
)

const (
	maxResponseSize     = 1 << 20 // 1MB limit
	defaultFetchTimeout = 15 * time.Second
)

type Fetch struct {
	client  *http.Client
	retrier *retry.Retrier
}

func NewFetch() *Fetch {
	return &Fetch{
		client: &http.Client{
			Timeout: defaultFetchTimeout,
		},
		retrier: retry.NewRetrier(retry.NewDefaultConfig()),
	}
}

func (f *Fetch) FetchURL(ctx context.Context, args json.RawMessage) (string, error) {
	var input struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	var body []byte
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

		// Limit response size
		limitedReader := io.LimitReader(resp.Body, maxResponseSize)
		body, err = io.ReadAll(limitedReader)
		if err != nil {
			return fmt.Errorf("failed to read body: %w", err)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return string(body), nil
}
