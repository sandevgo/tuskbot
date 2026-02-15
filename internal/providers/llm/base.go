package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type baseProvider struct {
	client  *http.Client
	baseURL string
	apiKey  string
	model   string
}

func newBaseProvider(baseURL, apiKey, model string) baseProvider {
	return baseProvider{
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
	}
}

func (b *baseProvider) doRequest(ctx context.Context, method, path string, body any, headers map[string]string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, b.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	return resp, nil
}
