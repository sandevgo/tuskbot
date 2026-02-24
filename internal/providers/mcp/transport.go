package mcp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/client"
	mcptransport "github.com/mark3labs/mcp-go/client/transport"
	mcpproto "github.com/mark3labs/mcp-go/mcp"
	"github.com/sandevgo/tuskbot/internal/core"
)

type Transport = func(ctx context.Context, cfg ServerConfig) (*client.Client, error)

func NewTransport(t TransportType) (Transport, error) {
	switch t {
	case TransportStdio:
		return StdioTransport, nil
	case TransportHTTP:
		return HttpTransport, nil
	case TransportSSE:
		return SseTransport, nil
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

func HttpTransport(ctx context.Context, cfg ServerConfig) (*client.Client, error) {
	// Create fresh transport to avoid shared state issues
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	headers := make(map[string]string)
	for k, v := range cfg.Headers {
		headers[k] = v
	}

	cli, err := client.NewStreamableHttpClient(
		cfg.URL,
		mcptransport.WithHTTPHeaders(headers),
		mcptransport.WithHTTPBasicClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSE transport: %w", err)
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

func SseTransport(ctx context.Context, cfg ServerConfig) (*client.Client, error) {
	// Create fresh transport to avoid shared state issues
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	headers := make(map[string]string)
	for k, v := range cfg.Headers {
		headers[k] = v
	}

	cli, err := client.NewSSEMCPClient(
		cfg.URL,
		mcptransport.WithHeaders(headers),
		mcptransport.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSE transport: %w", err)
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
