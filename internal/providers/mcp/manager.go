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

var _ core.MCPServer = (*Manager)(nil)

type Manager struct {
	registry *Registry
	pool     ConnectionPool
	cache    *ToolCache
	timeouts *Timeouts

	// Native tools support
	nativeTools    map[string]NativeHandler
	nativeToolDefs []core.Tool
}

func NewManager(
	ctx context.Context,
	runtimePath string,
	pool ConnectionPool,
	registry *Registry,
	cache *ToolCache,
) (*Manager, error) {
	nativeTools, nativeToolDefs := RegisterNativeTools(runtimePath, registry, pool, cache)

	mgr := &Manager{
		pool:           pool,
		registry:       registry,
		cache:          cache,
		timeouts:       NewDefaultTimeouts(),
		nativeTools:    nativeTools,
		nativeToolDefs: nativeToolDefs,
	}

	// Load initial config
	if err := mgr.registry.Load(ctx); err != nil {
		return nil, err
	}

	return mgr, nil
}

func (m *Manager) Start(ctx context.Context) error {
	servers := m.registry.List()

	// Start servers in parallel background goroutines
	for name, srv := range servers {
		go func(n string, s ServerConfig) {
			// Use a timeout derived from the parent context
			connectCtx, cancel := context.WithTimeout(ctx, m.timeouts.Connect)
			defer cancel()

			logger := log.FromCtx(ctx).With().Str("server", n).Logger()
			logger.Info().Msg("starting mcp server")

			if _, err := m.pool.Add(connectCtx, n, s); err != nil {
				logger.Error().Err(err).Msg("failed to start mcp server")
				return
			}

			m.cache.Invalidate()

			logger.Info().Msg("mcp server connected")
		}(name, srv)
	}

	return nil
}

func (m *Manager) Shutdown(ctx context.Context) error {
	return m.pool.Close()
}

func (m *Manager) GetTools(ctx context.Context) ([]core.Tool, error) {
	// 1. Check Cache
	if tools, _, _, ok := m.cache.Get(); ok {
		return tools, nil
	}

	// --- Cache Miss: Fetch from Servers ---

	// Start with native tools
	var allTools []core.Tool
	for _, t := range m.nativeToolDefs {
		allTools = append(allTools, t)
	}

	// Get all active clients from pool
	clients := m.pool.All()

	// Prepare for parallel fetching
	type toolResult struct {
		serverName string
		tools      []mcpproto.Tool
		err        error
	}
	results := make(chan toolResult, len(clients))
	var wg sync.WaitGroup

	for name, cli := range clients {
		wg.Add(1)
		go func(n string, c *ManagedClient) {
			defer wg.Done()
			tCtx, cancel := context.WithTimeout(ctx, m.timeouts.ToolList)
			defer cancel()

			resp, err := c.ListTools(tCtx, mcpproto.ListToolsRequest{})
			if err != nil {
				results <- toolResult{serverName: n, err: err}
				return
			}
			results <- toolResult{serverName: n, tools: resp.Tools}
		}(name, cli)
	}

	wg.Wait()
	close(results)

	// Aggregate results
	newToolToServer := make(map[string]string)

	for res := range results {
		if res.err != nil {
			log.FromCtx(ctx).Error().Err(res.err).Str("server", res.serverName).Msg("failed to list tools")
			continue
		}

		for _, t := range res.tools {
			// example: "filesystem.read_file"
			qualifiedName := fmt.Sprintf("%s.%s", res.serverName, t.Name)

			// Map tool name to server name for routing
			newToolToServer[qualifiedName] = res.serverName

			schemaBytes, _ := json.Marshal(t.InputSchema)
			allTools = append(allTools, core.Tool{
				Type: "function",
				Function: core.Function{
					Name:        qualifiedName,
					Description: t.Description,
					Parameters:  schemaBytes,
				},
			})
		}
	}

	// Update Cache
	m.cache.Update(allTools, newToolToServer)

	return allTools, nil
}

func (m *Manager) CallTool(ctx context.Context, name string, args string) (string, error) {
	log.FromCtx(ctx).Info().Str("tool", name).Str("args", args).Msg("executing tool")

	// 1. Check Native Tools first
	if handler, ok := m.nativeTools[name]; ok {
		return handler(ctx, json.RawMessage(args))
	}

	// 2. Resolve Server
	_, routing, _, _ := m.cache.Get()
	serverName, ok := routing[name]

	if !ok {
		return "", fmt.Errorf("tool not found: %s", name)
	}

	// 3. Get Client from Pool
	cli, ok := m.pool.Get(serverName)
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
	tCtx, cancel := context.WithTimeout(ctx, m.timeouts.ToolCall)
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
