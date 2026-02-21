package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sandevgo/tuskbot/pkg/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetch_FetchURL(t *testing.T) {
	tests := []struct {
		name            string
		args            json.RawMessage
		mockServer      func() *httptest.Server
		useShortTimeout bool
		wantErr         bool
		wantContains    string
		wantErrMsg      string
	}{
		{
			name: "successful HTML fetch",
			args: json.RawMessage(`{"url": "REPLACE_URL"}`),
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(http.StatusOK)
					fmt.Fprint(w, `<html><body><h1>Test Page</h1><p>Hello World</p><a href="/link">Click here</a></body></html>`)
				}))
			},
			wantContains: "Test Page",
		},
		{
			name: "successful JSON fetch",
			args: json.RawMessage(`{"url": "REPLACE_URL"}`),
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					fmt.Fprint(w, `{"message": "Hello JSON"}`)
				}))
			},
			wantContains: `{"message": "Hello JSON"}`,
		},
		{
			name: "404 error",
			args: json.RawMessage(`{"url": "REPLACE_URL"}`),
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			wantErr:    true,
			wantErrMsg: "HTTP 404",
		},
		{
			name: "500 error",
			args: json.RawMessage(`{"url": "REPLACE_URL"}`),
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			wantErr:    true,
			wantErrMsg: "HTTP 500",
		},
		{
			name:       "invalid JSON args",
			args:       json.RawMessage(`{"invalid`),
			wantErr:    true,
			wantErrMsg: "invalid arguments",
		},
		{
			name:       "missing URL",
			args:       json.RawMessage(`{}`),
			wantErr:    true,
			wantErrMsg: "failed to fetch url",
		},
		{
			name: "large response gets truncated",
			args: json.RawMessage(`{"url": "REPLACE_URL"}`),
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusOK)
					// Write more than 1MB
					for i := 0; i < 1024*1024+100; i++ {
						w.Write([]byte("a"))
					}
				}))
			},
			wantContains: strings.Repeat("a", 1024*1024), // Should be exactly 1MB
		},
		{
			name: "timeout handling",
			args: json.RawMessage(`{"url": "REPLACE_URL"}`),
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Sleep longer than the short test timeout (100ms)
					time.Sleep(500 * time.Millisecond)
				}))
			},
			useShortTimeout: true,
			wantErr:         true,
			wantErrMsg:      "failed to fetch url",
		},
		{
			name: "preserves links in HTML",
			args: json.RawMessage(`{"url": "REPLACE_URL"}`),
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(http.StatusOK)
					fmt.Fprint(w, `<html><body>
						<nav>
							<a href="/home">Home</a>
							<a href="/about">About</a>
						</nav>
						<article>
							<h1>Main Article</h1>
							<p>Content with <a href="https://example.com">external link</a></p>
						</article>
						<aside>
							<a href="/related1">Related 1</a>
							<a href="/related2">Related 2</a>
						</aside>
					</body></html>`)
				}))
			},
			wantContains: "Home", // Should preserve navigation links
		},
		{
			name: "handles special characters in HTML",
			args: json.RawMessage(`{"url": "REPLACE_URL"}`),
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.WriteHeader(http.StatusOK)
					fmt.Fprint(w, `<html><body><p>Special chars: &lt;&gt;&amp; "quotes" 'apostrophe'</p></body></html>`)
				}))
			},
			wantContains: "Special chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			args := tt.args

			// Start mock server if needed
			if tt.mockServer != nil {
				server = tt.mockServer()
				defer server.Close()

				// Replace URL placeholder with actual server URL
				args = json.RawMessage(strings.Replace(string(tt.args), "REPLACE_URL", server.URL, 1))
			}

			// Create fetch instance
			// Create fetch instance
			retryCfg := &retry.Config{
				MaxRetries:    2,
				InitialDelay:  time.Millisecond,
				MaxDelay:      time.Millisecond,
				BackoffFactor: 1.0,
			}

			var fetch *Fetch
			if tt.useShortTimeout {
				fetch = NewFetchWithTimeout(100*time.Millisecond, retryCfg)
			} else {
				fetch = NewFetchWithTimeout(defaultFetchTimeout, retryCfg)
			}

			// Execute
			ctx := context.Background()
			result, err := fetch.FetchURL(ctx, args)

			// Verify
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				require.NoError(t, err)
				if tt.wantContains != "" {
					assert.Contains(t, result, tt.wantContains)
				}
			}
		})
	}
}

func TestFetch_RetryBehavior(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// Fail first 2 attempts
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Succeed on 3rd attempt
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Success after retries")
	}))
	defer server.Close()

	fetch := NewFetch()
	args := json.RawMessage(fmt.Sprintf(`{"url": "%s"}`, server.URL))

	ctx := context.Background()
	result, err := fetch.FetchURL(ctx, args)

	require.NoError(t, err)
	assert.Contains(t, result, "Success after retries")
	assert.Equal(t, 3, attempts, "should retry failed requests")
}

func TestFetch_UserAgent(t *testing.T) {
	var receivedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	fetch := NewFetch()
	args := json.RawMessage(fmt.Sprintf(`{"url": "%s"}`, server.URL))

	ctx := context.Background()
	_, err := fetch.FetchURL(ctx, args)

	require.NoError(t, err)
	assert.Contains(t, receivedUA, "TuskBot", "should set custom user agent")
}
