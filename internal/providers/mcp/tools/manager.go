package tools

//const manageMcpSchema = `
//{
//  "type": "object",
//  "properties": {
//    "action": {
//      "type": "string",
//      "enum": ["add", "remove", "reload"],
//      "description": "What to do with the server"
//    },
//    "server_name": {
//      "type": "string",
//      "description": "Unique name for the server"
//    },
//    "command": {
//      "type": "string",
//      "description": "Command to run (e.g. npx, python, node). Required for 'add' with stdio."
//    },
//    "args": {
//      "type": "array",
//      "items": { "type": "string" },
//      "description": "Arguments for the command"
//    },
//    "env": {
//      "type": "object",
//      "additionalProperties": { "type": "string" },
//      "description": "Environment variables (e.g. API keys)"
//    },
//    "url": {
//      "type": "string",
//      "description": "Server URL (e.g. http://localhost:8080/sse). Required for 'add' with http."
//    }
//  },
//  "required": ["action", "server_name"]
//}
//`
//
//const (
//	actionAdd      = "add"
//	actionRemove   = "remove"
//	actionReload   = "reload"
//	connectTimeout = 30 * time.Second
//)
//
//type managementInput struct {
//	Action     string            `json:"action"`
//	ServerName string            `json:"server_name"`
//	Command    string            `json:"command"`
//	Args       []string          `json:"args"`
//	Env        map[string]string `json:"env"`
//	URL        string            `json:"url"`
//}
//
//type ManageTool struct {
//	registry *Registry
//	pool     ConnectionPool
//	cache    *ToolCache
//	timeout  time.Duration
//}
//
//func NewManageTool(
//	registry *Registry,
//	pool ConnectionPool,
//	cache *ToolCache,
//) *ManageTool {
//	return &ManageTool{
//		registry: registry,
//		pool:     pool,
//		cache:    cache,
//		timeout:  connectTimeout,
//	}
//}
//
//func (t *ManageTool) ManageMCP(ctx context.Context, args json.RawMessage) (string, error) {
//	var input managementInput
//	if err := json.Unmarshal(args, &input); err != nil {
//		return "", fmt.Errorf("invalid arguments: %w", err)
//	}
//
//	switch input.Action {
//	case actionAdd:
//		return t.handleAdd(ctx, input)
//	case actionRemove:
//		return t.handleRemove(ctx, input)
//	case actionReload:
//		return t.handleReload(ctx, input)
//	default:
//		return "", fmt.Errorf("unknown action: %s", input.Action)
//	}
//}
//
//func (t *ManageTool) handleAdd(ctx context.Context, input managementInput) (string, error) {
//	if input.Command == "" && input.URL == "" {
//		return "", fmt.Errorf("command or url is required for add action")
//	}
//
//	cleanEnv := make(map[string]string)
//	for k, v := range input.Env {
//		cleanKey := strings.Trim(k, "\"'")
//		cleanEnv[cleanKey] = v
//	}
//
//	newCfg := ServerConfig{
//		Command: input.Command,
//		Args:    input.Args,
//		Env:     cleanEnv,
//		URL:     input.URL,
//	}
//
//	// 1. Add to Pool (Handles connection and verification)
//	connectCtx, cancel := context.WithTimeout(ctx, t.timeout)
//	defer cancel()
//
//	if _, err := t.pool.Add(connectCtx, input.ServerName, newCfg); err != nil {
//		return "", fmt.Errorf("failed to connect to new server: %w", err)
//	}
//
//	// 2. Update Registry
//	if err := t.registry.Add(ctx, input.ServerName, newCfg); err != nil {
//		return "", fmt.Errorf("server started but registry save failed: %w", err)
//	}
//
//	t.cache.Invalidate()
//
//	return fmt.Sprintf("Server %s added and started", input.ServerName), nil
//}
//
//func (t *ManageTool) handleRemove(ctx context.Context, input managementInput) (string, error) {
//	// 1. Remove from Pool
//	if err := t.pool.Del(input.ServerName); err != nil {
//		log.FromCtx(ctx).Warn().Err(err).Str("server", input.ServerName).Msg("error closing server during removal")
//	}
//
//	// 2. Update Registry
//	if err := t.registry.Remove(ctx, input.ServerName); err != nil {
//		return "", err
//	}
//
//	t.cache.Invalidate()
//
//	return fmt.Sprintf("Server %s removed", input.ServerName), nil
//}
//
//func (t *ManageTool) handleReload(ctx context.Context, input managementInput) (string, error) {
//	// Refresh registry from storage in case of manual edits
//	if err := t.registry.Load(ctx); err != nil {
//		return "", fmt.Errorf("failed to load registry: %w", err)
//	}
//
//	srvCfg, exists := t.registry.Get(input.ServerName)
//	if !exists {
//		return "", fmt.Errorf("server %s not found in registry", input.ServerName)
//	}
//
//	// Pool.Add handles closing the old connection if it exists
//	connectCtx, cancel := context.WithTimeout(ctx, t.timeout)
//	defer cancel()
//
//	if _, err := t.pool.Add(connectCtx, input.ServerName, srvCfg); err != nil {
//		return "", fmt.Errorf("failed to reload server: %w", err)
//	}
//
//	t.cache.Invalidate()
//
//	return fmt.Sprintf("Server %s reloaded", input.ServerName), nil
//}
//
//func (t *ManageTool) GetDefinitions() map[string]struct {
//	Description string
//	Schema      string
//	Handler     func(context.Context, json.RawMessage) (string, error)
//} {
//	return map[string]struct {
//		Description string
//		Schema      string
//		Handler     func(context.Context, json.RawMessage) (string, error)
//	}{
//		"manage_mcp": {"Manage MCP servers (add, remove, reload)", manageMcpSchema, t.ManageMCP},
//	}
//}
