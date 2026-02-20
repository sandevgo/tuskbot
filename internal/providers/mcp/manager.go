package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	mcpproto "github.com/mark3labs/mcp-go/mcp"
	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

// NativeHandler defines a function signature for internal tools
type NativeHandler func(ctx context.Context, args json.RawMessage) (string, error)

const manageMcpSchema = `
{
  "type": "object",
  "properties": {
    "action": { 
      "type": "string", 
      "enum": ["add", "remove", "reload"], 
      "description": "What to do with the server" 
    },
    "server_name": { 
      "type": "string", 
      "description": "Unique name for the server" 
    },
    "command": { 
      "type": "string", 
      "description": "Command to run (e.g. npx, python, node). Required for 'add'." 
    },
    "args": { 
      "type": "array", 
      "items": { "type": "string" }, 
      "description": "Arguments for the command" 
    },
    "env": { 
      "type": "object", 
      "additionalProperties": { "type": "string" }, 
      "description": "Environment variables (e.g. API keys)" 
    }
  },
  "required": ["action", "server_name"]
}
`

var _ core.MCPServer = (*Manager)(nil)

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

type Manager struct {
	config   Config
	storage  Storage
	pool     ConnectionPool
	timeouts *Timeouts

	mu           sync.RWMutex
	toolToServer map[string]string // Maps tool name -> server name

	// Caching
	cachedTools []core.Tool
	cacheValid  bool

	// Native tools support
	nativeTools    map[string]NativeHandler
	nativeToolDefs []core.Tool
}

func NewManager(
	ctx context.Context,
	pool ConnectionPool,
	storage Storage,
	timeouts *Timeouts,
) (*Manager, error) {
	mgr := &Manager{
		pool:           pool,
		storage:        storage,
		timeouts:       timeouts,
		nativeTools:    make(map[string]NativeHandler),
		nativeToolDefs: make([]core.Tool, 0),
	}

	cfg, err := mgr.storage.Load(ctx)
	if err != nil {
		return nil, err
	}
	mgr.config = *cfg

	// Register the manage_mcp tool
	mgr.RegisterNativeTool(
		"manage_mcp",
		"Manage MCP servers (add, remove, reload)",
		json.RawMessage(manageMcpSchema),
		mgr.ManageMCP,
	)

	return mgr, nil
}

// RegisterNativeTool allows adding hardcoded Go functions as tools
func (m *Manager) RegisterNativeTool(name, description string, schema json.RawMessage, handler NativeHandler) {
	m.nativeTools[name] = handler
	m.nativeToolDefs = append(m.nativeToolDefs, core.Tool{
		Type: "function",
		Function: core.Function{
			Name:        name,
			Description: description,
			Parameters:  schema,
		},
	})
}

func (m *Manager) Start(ctx context.Context) error {
	// Snapshot config to avoid holding lock during connection
	m.mu.RLock()
	servers := make(map[string]ServerConfig, len(m.config.MCPServers))
	for k, v := range m.config.MCPServers {
		servers[k] = v
	}
	m.mu.RUnlock()

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

			m.mu.Lock()
			m.cacheValid = false
			m.mu.Unlock()

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
	m.mu.RLock()
	if m.cacheValid {
		tools := m.cachedTools
		m.mu.RUnlock()
		return tools, nil
	}
	m.mu.RUnlock()

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
	m.mu.Lock()
	m.cachedTools = allTools
	m.toolToServer = newToolToServer
	m.cacheValid = true
	m.mu.Unlock()

	return allTools, nil
}

func (m *Manager) CallTool(ctx context.Context, name string, args string) (string, error) {
	log.FromCtx(ctx).Info().Str("tool", name).Str("args", args).Msg("executing tool")

	// 1. Check Native Tools first
	if handler, ok := m.nativeTools[name]; ok {
		return handler(ctx, json.RawMessage(args))
	}

	// 2. Resolve Server
	m.mu.RLock()
	serverName, ok := m.toolToServer[name]
	m.mu.RUnlock()

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

func (m *Manager) ManageMCP(ctx context.Context, args json.RawMessage) (string, error) {
	var input struct {
		Action     string            `json:"action"`
		ServerName string            `json:"server_name"`
		Command    string            `json:"command"`
		Args       []string          `json:"args"`
		Env        map[string]string `json:"env"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	switch input.Action {
	case "add":
		if input.Command == "" {
			return "", fmt.Errorf("command is required for add action")
		}

		cleanEnv := make(map[string]string)
		for k, v := range input.Env {
			cleanKey := strings.Trim(k, "\"'")
			cleanEnv[cleanKey] = v
		}

		newCfg := ServerConfig{
			Command: input.Command,
			Args:    input.Args,
			Env:     cleanEnv,
		}

		// 1. Add to Pool (Handles connection and verification)
		connectCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		if _, err := m.pool.Add(connectCtx, input.ServerName, newCfg); err != nil {
			return "", fmt.Errorf("failed to connect to new server: %w", err)
		}

		// 2. Update MCPConfig
		m.mu.Lock()
		m.config.MCPServers[input.ServerName] = newCfg
		m.cacheValid = false
		m.mu.Unlock()

		if err := m.storage.Save(ctx, &m.config); err != nil {
			return "", fmt.Errorf("server started but config save failed: %w", err)
		}
		return fmt.Sprintf("Server %s added and started", input.ServerName), nil

	case "remove":
		// 1. Remove from Pool
		if err := m.pool.Remove(input.ServerName); err != nil {
			log.FromCtx(ctx).Warn().Err(err).Str("server", input.ServerName).Msg("error closing server during removal")
		}

		// 2. Update MCPConfig
		m.mu.Lock()
		delete(m.config.MCPServers, input.ServerName)
		m.cacheValid = false
		m.mu.Unlock()

		if err := m.storage.Save(ctx, &m.config); err != nil {
			return "", err
		}
		return fmt.Sprintf("Server %s removed", input.ServerName), nil

	case "reload":
		m.mu.RLock()
		srvCfg, exists := m.config.MCPServers[input.ServerName]
		m.mu.RUnlock()

		if !exists {
			return "", fmt.Errorf("server %s not found in config", input.ServerName)
		}

		// Pool.Add handles closing the old connection if it exists
		connectCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		if _, err := m.pool.Add(connectCtx, input.ServerName, srvCfg); err != nil {
			return "", fmt.Errorf("failed to reload server: %w", err)
		}

		m.mu.Lock()
		m.cacheValid = false
		m.mu.Unlock()

		return fmt.Sprintf("Server %s reloaded", input.ServerName), nil

	default:
		return "", fmt.Errorf("unknown action: %s", input.Action)
	}
}
