package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sandevgo/tuskbot/pkg/retry"
)

type baseProvider struct {
	client  *http.Client
	baseURL string
	apiKey  string
	model   string
	retrier *retry.Retrier
}

func newBaseProvider(baseURL, apiKey, model string) baseProvider {
	return baseProvider{
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		retrier: retry.NewDefaultRetrier(),
	}
}

func (b *baseProvider) GetModel() string {
	return b.model
}

func (b *baseProvider) doRequest(ctx context.Context, method, path string, body any, headers map[string]string) (*http.Response, error) {
	var bodyData []byte
	if body != nil {
		var err error
		bodyData, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal: %w", err)
		}
	}

	var resp *http.Response

	err := b.retrier.Do(ctx, func() error {
		var bodyReader io.Reader
		if bodyData != nil {
			bodyReader = bytes.NewReader(bodyData)
		}

		req, err := http.NewRequestWithContext(ctx, method, b.baseURL+path, bodyReader)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}
		req.Header.Set("Content-Type", "application/json")

		r, err := b.client.Do(req)
		if err != nil {
			return err
		}

		// Retry on server errors (5xx) and rate limiting (429)
		if r.StatusCode >= 500 || r.StatusCode == 429 {
			errBody, _ := io.ReadAll(io.LimitReader(r.Body, 1024))
			r.Body.Close()
			return fmt.Errorf("retryable status %d: %s", r.StatusCode, string(errBody))
		}

		resp = r
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}
