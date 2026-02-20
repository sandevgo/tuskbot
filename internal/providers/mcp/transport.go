package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	mcpproto "github.com/mark3labs/mcp-go/mcp"
	"github.com/sandevgo/tuskbot/internal/core"
)

type Transport interface {
	Connect(ctx context.Context, cfg ServerConfig) (*client.Client, error)
}

type StdioTransport struct{}

func (t *StdioTransport) Connect(ctx context.Context, cfg ServerConfig) (*client.Client, error) {
	var env []string
	for k, v := range cfg.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	cli, err := client.NewStdioMCPClient(cfg.Command, env, cfg.Args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if err := cli.Start(ctx); err != nil {
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
