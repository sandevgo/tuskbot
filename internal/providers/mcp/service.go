package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	mcpproto "github.com/mark3labs/mcp-go/mcp"
	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type Timeouts struct {
	Connect  time.Duration
	ToolList time.Duration
	ToolCall time.Duration
}

func NewDefaultTimeouts() *Timeouts {
	return &Timeouts{
		Connect:  30 * time.Second,
		ToolList: 5 * time.Second,
		ToolCall: 2 * time.Minute,
	}
}

// NativeHandler defines a function signature for internal tools
type NativeHandler func(ctx context.Context, args json.RawMessage) (string, error)

var _ core.MCPServer = (*Service)(nil)

type Service struct {
	registry *Registry
	pool     ConnectionPool
	cache    *ToolCache
	timeouts *Timeouts

	// Native tools support
	nativeTools    map[string]NativeHandler
	nativeToolDefs []core.Tool
}

func NewService(
	runtimePath string,
	pool ConnectionPool,
	registry *Registry,
	cache *ToolCache,
) (*Service, error) {
	nativeTools, nativeToolDefs := RegisterNativeTools(runtimePath)

	return &Service{
		pool:           pool,
		registry:       registry,
		cache:          cache,
		timeouts:       NewDefaultTimeouts(),
		nativeTools:    nativeTools,
		nativeToolDefs: nativeToolDefs,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	// Load initial config
	if err := s.registry.Load(ctx); err != nil {
		return err
	}

	servers := s.registry.List()

	// Start servers in parallel background goroutines
	for name, srv := range servers {
		go func(n string, c ServerConfig) {
			// Use a timeout derived from the parent context
			connectCtx, cancel := context.WithTimeout(ctx, s.timeouts.Connect)
			defer cancel()

			logger := log.FromCtx(ctx).With().Str("server", n).Logger()
			logger.Info().Msg("starting mcp server")

			if _, err := s.pool.Add(connectCtx, n, c); err != nil {
				logger.Error().Err(err).Msg("failed to start mcp server")
				return
			}

			s.cache.Invalidate()

			logger.Info().Msg("mcp server connected")
		}(name, srv)
	}

	return nil
}

func (s *Service) Shutdown(ctx context.Context) error {
	return s.pool.Close()
}

func (s *Service) GetTools(ctx context.Context) ([]core.Tool, error) {
	if tools, _, ok := s.cache.Get(); ok {
		return tools, nil
	}

	// Start with native tools (already in memory)
	allTools := make([]core.Tool, len(s.nativeToolDefs))
	copy(allTools, s.nativeToolDefs)

	// Fetch from external servers concurrently
	serverTools, routing := s.fetchToolsFromServers(ctx)

	// Aggregate
	for _, tools := range serverTools {
		allTools = append(allTools, tools...)
	}

	// Update Cache
	s.cache.Update(allTools, routing)

	return allTools, nil
}

func (s *Service) fetchToolsFromServers(ctx context.Context) (map[string][]core.Tool, map[string]string) {
	type toolResult struct {
		serverName string
		tools      []core.Tool
		err        error
	}

	clients := s.pool.All()
	results := make(chan toolResult, len(clients))
	var wg sync.WaitGroup

	for name, cli := range clients {
		wg.Add(1)
		go func(n string, c *ManagedClient) {
			defer wg.Done()
			tools, err := s.listToolsFromServer(ctx, n, c)
			results <- toolResult{serverName: n, tools: tools, err: err}
		}(name, cli)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	serverTools := make(map[string][]core.Tool)
	routing := make(map[string]string)

	for res := range results {
		if res.err != nil {
			log.FromCtx(ctx).Error().Err(res.err).Str("server", res.serverName).Msg("failed to list tools")
			continue
		}
		serverTools[res.serverName] = res.tools
		for _, t := range res.tools {
			routing[t.Function.Name] = res.serverName
		}
	}

	return serverTools, routing
}

func (s *Service) listToolsFromServer(ctx context.Context, name string, cli *ManagedClient) ([]core.Tool, error) {
	tCtx, cancel := context.WithTimeout(ctx, s.timeouts.ToolList)
	defer cancel()

	resp, err := cli.ListTools(tCtx, mcpproto.ListToolsRequest{})
	if err != nil {
		return nil, err
	}

	tools := make([]core.Tool, 0, len(resp.Tools))
	for _, t := range resp.Tools {
		schemaBytes, _ := json.Marshal(t.InputSchema)
		tools = append(tools, core.Tool{
			Type: "function",
			Function: core.Function{
				Name:        fmt.Sprintf("%s.%s", name, t.Name),
				Description: t.Description,
				Parameters:  schemaBytes,
			},
		})
	}
	return tools, nil
}

func (s *Service) CallTool(ctx context.Context, name string, args string) (string, error) {
	log.FromCtx(ctx).Info().Str("tool", name).Str("args", args).Msg("executing tool")

	// 1. Check Native Tools first
	if handler, ok := s.nativeTools[name]; ok {
		return handler(ctx, json.RawMessage(args))
	}

	// 2. Resolve Server
	_, routing, _ := s.cache.Get()
	serverName, ok := routing[name]

	if !ok {
		return "", fmt.Errorf("tool not found: %s", name)
	}

	// 3. Get Client from Pool
	cli, ok := s.pool.Get(serverName)
	if !ok {
		return "", fmt.Errorf("server %s is not available", serverName)
	}

	// 4. Execute
	var argsMap map[string]interface{}
	if err := json.Unmarshal([]byte(args), &argsMap); err != nil {
		return "", fmt.Errorf("invalid json arguments: %w", err)
	}

	req := mcpproto.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = argsMap

	// Set a reasonable timeout for tool execution
	tCtx, cancel := context.WithTimeout(ctx, s.timeouts.ToolCall)
	defer cancel()

	res, err := cli.CallTool(tCtx, req)
	if err != nil {
		return "", err
	}

	if res.IsError {
		return "", fmt.Errorf("tool execution failed")
	}

	var output string
	for _, content := range res.Content {
		if text, ok := content.(mcpproto.TextContent); ok {
			output += text.Text + "\n"
		} else if textPtr, ok := content.(*mcpproto.TextContent); ok {
			output += textPtr.Text + "\n"
		}
	}
	return output, nil
}
