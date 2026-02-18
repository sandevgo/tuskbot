package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/client"
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

type Manager struct {
	mu           sync.RWMutex
	configPath   string
	config       Config
	clients      map[string]*client.Client
	toolToClient map[string]*client.Client

	// Caching
	cachedTools []core.Tool
	cacheValid  bool

	// Native tools support
	nativeTools    map[string]NativeHandler
	nativeToolDefs []core.Tool
}

func NewManager(ctx context.Context, configPath string) (*Manager, error) {
	mgr := &Manager{
		configPath:     configPath,
		clients:        make(map[string]*client.Client),
		toolToClient:   make(map[string]*client.Client),
		nativeTools:    make(map[string]NativeHandler),
		nativeToolDefs: make([]core.Tool, 0),
	}

	if err := mgr.loadConfig(ctx); err != nil {
		return nil, err
	}

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
	// This ensures that one hanging server doesn't block the entire bot startup
	for name, srv := range servers {
		go func(n string, s ServerConfig) {
			// Use a timeout for the connection attempt
			connectCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Create a logger with context for this specific server
			logger := log.FromCtx(ctx).With().Str("server", n).Logger()
			logger.Info().Msg("starting mcp server")

			cli, err := m.connectToServer(connectCtx, s)
			if err != nil {
				logger.Error().Err(err).Msg("failed to start mcp server")
				return
			}

			m.mu.Lock()
			m.clients[n] = cli
			m.cacheValid = false // Invalidate cache to include new client
			m.mu.Unlock()

			logger.Info().Msg("mcp server connected")
		}(name, srv)
	}

	return nil
}

func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.RLock()
	clients := make(map[string]*client.Client, len(m.clients))
	for k, v := range m.clients {
		clients[k] = v
	}
	m.mu.RUnlock()

	var wg sync.WaitGroup
	for name, cli := range clients {
		wg.Add(1)
		go func(n string, c *client.Client) {
			defer wg.Done()

			// Create a channel to handle close timeout
			done := make(chan struct{})
			go func() {
				if err := c.Close(); err != nil {
					log.FromCtx(ctx).Error().Err(err).Str("server", n).Msg("failed to close client")
				}
				close(done)
			}()

			select {
			case <-done:
				return
			case <-time.After(5 * time.Second):
				log.FromCtx(ctx).Warn().Str("server", n).Msg("mcp server close timed out, forcing shutdown")
			}
		}(name, cli)
	}
	wg.Wait()
	return nil
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

	// Snapshot clients to avoid holding lock during network I/O
	m.mu.RLock()
	clientsSnapshot := make(map[string]*client.Client, len(m.clients))
	for k, v := range m.clients {
		clientsSnapshot[k] = v
	}
	m.mu.RUnlock()

	// Prepare for parallel fetching
	type toolResult struct {
		serverName string
		tools      []mcpproto.Tool
		err        error
	}
	results := make(chan toolResult, len(clientsSnapshot))
	var wg sync.WaitGroup

	for name, cli := range clientsSnapshot {
		wg.Add(1)
		go func(n string, c *client.Client) {
			defer wg.Done()
			tCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
	newToolToClient := make(map[string]*client.Client)

	for res := range results {
		if res.err != nil {
			log.FromCtx(ctx).Error().Err(res.err).Str("server", res.serverName).Msg("failed to list tools")
			continue
		}

		for _, t := range res.tools {
			// We need the client to call the tool later
			// Note: If multiple servers have the same tool name, the last one wins (randomly due to map iteration)
			// In a real system, we might want namespacing (e.g. server__tool)
			newToolToClient[t.Name] = clientsSnapshot[res.serverName]

			schemaBytes, _ := json.Marshal(t.InputSchema)
			allTools = append(allTools, core.Tool{
				Type: "function",
				Function: core.Function{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  schemaBytes,
				},
			})
		}
	}

	// Update Cache
	m.mu.Lock()
	m.cachedTools = allTools
	m.toolToClient = newToolToClient
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

	// 2. Check External Clients
	m.mu.RLock()
	cli, ok := m.toolToClient[name]
	m.mu.RUnlock()

	if !ok {
		// If tool not found, maybe cache is stale?
		// We could force a refresh here, but for now just return error
		return "", fmt.Errorf("tool not found: %s", name)
	}

	var argsMap map[string]interface{}
	if err := json.Unmarshal([]byte(args), &argsMap); err != nil {
		return "", fmt.Errorf("invalid json arguments: %w", err)
	}

	req := mcpproto.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = argsMap

	// Set a reasonable timeout for tool execution
	tCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
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

		// Sanitize environment keys
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

		// 1. Connect WITHOUT lock (Heavy I/O)
		// Use a strict timeout for adding new tools
		connectCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		newClient, err := m.connectToServer(connectCtx, newCfg)
		if err != nil {
			return "", fmt.Errorf("failed to connect to new server: %w", err)
		}

		// 2. Update State WITH lock
		m.mu.Lock()
		if oldCli, exists := m.clients[input.ServerName]; exists {
			_ = oldCli.Close()
		}
		m.clients[input.ServerName] = newClient
		m.config.MCPServers[input.ServerName] = newCfg
		m.cacheValid = false // Invalidate cache
		m.mu.Unlock()

		if err := m.saveConfig(); err != nil {
			return "Server started but config save failed", err
		}
		return fmt.Sprintf("Server %s added and started", input.ServerName), nil

	case "remove":
		m.mu.Lock()
		if oldCli, exists := m.clients[input.ServerName]; exists {
			_ = oldCli.Close()
			delete(m.clients, input.ServerName)
		}
		delete(m.config.MCPServers, input.ServerName)
		m.cacheValid = false // Invalidate cache
		m.mu.Unlock()

		if err := m.saveConfig(); err != nil {
			return "", err
		}
		return fmt.Sprintf("Server %s removed", input.ServerName), nil

	case "reload":
		// 1. Get Config (Read Lock)
		m.mu.RLock()
		srvCfg, exists := m.config.MCPServers[input.ServerName]
		m.mu.RUnlock()

		if !exists {
			return "", fmt.Errorf("server %s not found in config", input.ServerName)
		}

		// 2. Connect New Client (No Lock)
		connectCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		newClient, err := m.connectToServer(connectCtx, srvCfg)
		if err != nil {
			return "", fmt.Errorf("failed to reconnect: %w", err)
		}

		// 3. Swap Clients (Write Lock)
		m.mu.Lock()
		if oldCli, exists := m.clients[input.ServerName]; exists {
			_ = oldCli.Close()
		}
		m.clients[input.ServerName] = newClient
		m.cacheValid = false // Invalidate cache
		m.mu.Unlock()

		return fmt.Sprintf("Server %s reloaded", input.ServerName), nil

	default:
		return "", fmt.Errorf("unknown action: %s", input.Action)
	}
}

func (m *Manager) connectToServer(ctx context.Context, srv ServerConfig) (*client.Client, error) {
	var env []string
	for k, v := range srv.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	cli, err := client.NewStdioMCPClient(srv.Command, env, srv.Args...)
	if err != nil {
		return nil, err
	}

	if err := cli.Start(ctx); err != nil {
		return nil, err
	}

	initReq := mcpproto.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcpproto.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcpproto.Implementation{
		Name:    core.TuskName,
		Version: core.TaskVersion,
	}
	initReq.Params.Capabilities = mcpproto.ClientCapabilities{}

	if _, err := cli.Initialize(ctx, initReq); err != nil {
		_ = cli.Close()
		return nil, err
	}

	return cli, nil
}

func (m *Manager) loadConfig(ctx context.Context) error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.FromCtx(ctx).Info().Msg("mcp_config.json not found, creating default")

			defaultCfg := Config{MCPServers: make(map[string]ServerConfig)}
			data, err = json.MarshalIndent(defaultCfg, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal default config: %w", err)
			}

			if err := os.WriteFile(m.configPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write default config: %w", err)
			}
		} else {
			return fmt.Errorf("failed to read mcp config: %w", err)
		}
	}

	if err := json.Unmarshal(data, &m.config); err != nil {
		return fmt.Errorf("failed to parse mcp config: %w", err)
	}
	return nil
}

func (m *Manager) saveConfig() error {
	// Note: We should ideally lock before reading m.config, but since this is called
	// inside ManageMCP where we handle locking or have local copies, it's generally okay.
	// For strict correctness, we can RLock here, but we need to be careful of deadlocks if called from within a Lock.
	// In this implementation, saveConfig is called AFTER Unlock in ManageMCP, so we should RLock.

	m.mu.RLock()
	data, err := json.MarshalIndent(m.config, "", "  ")
	m.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}
