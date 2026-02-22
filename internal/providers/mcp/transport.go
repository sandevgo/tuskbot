package mcp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/client"
	mcpproto "github.com/mark3labs/mcp-go/mcp"
	"github.com/sandevgo/tuskbot/internal/core"
)

type Transport = func(ctx context.Context, cfg ServerConfig) (*client.Client, error)

func NewTransport(t TransportType) (Transport, error) {
	switch t {
	case TransportStdio:
		return StdioTransport, nil
	case TransportHTTP:
		return SSETransport, nil
	}

	return nil, fmt.Errorf("unsupported transport type: %s", t)
}

func StdioTransport(ctx context.Context, cfg ServerConfig) (*client.Client, error) {
	var env []string
	for k, v := range cfg.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	cli, err := client.NewStdioMCPClient(cfg.Command, env, cfg.Args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if err = cli.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start client: %w", err)
	}

	req := mcpproto.InitializeRequest{}
	req.Params.ProtocolVersion = mcpproto.LATEST_PROTOCOL_VERSION
	req.Params.Capabilities = mcpproto.ClientCapabilities{}
	req.Params.ClientInfo = mcpproto.Implementation{
		Name:    core.TuskName,
		Version: core.TaskVersion,
	}

	if _, err := cli.Initialize(ctx, req); err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	return cli, nil
}

func SSETransport(ctx context.Context, cfg ServerConfig) (*client.Client, error) {
	httpClient := http.DefaultClient

	if len(cfg.Headers) > 0 {
		httpClient = &http.Client{
			Transport: &headerTransport{
				transport: http.DefaultTransport,
				headers:   cfg.Headers,
			},
		}
	}

	options := client.WithHTTPClient(httpClient)

	cli, err := client.NewSSEMCPClient(cfg.URL, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if err = cli.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start client: %w", err)
	}

	req := mcpproto.InitializeRequest{}
	req.Params.ProtocolVersion = mcpproto.LATEST_PROTOCOL_VERSION
	req.Params.Capabilities = mcpproto.ClientCapabilities{}
	req.Params.ClientInfo = mcpproto.Implementation{
		Name:    core.TuskName,
		Version: core.TaskVersion,
	}

	if _, err := cli.Initialize(ctx, req); err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	return cli, nil
}

type headerTransport struct {
	transport http.RoundTripper
	headers   map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	return t.transport.RoundTrip(req)
}
