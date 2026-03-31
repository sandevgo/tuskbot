# Custom MCP Servers

TuskBot can connect to external MCP servers for additional capabilities.

## Where Configuration Lives

External MCP servers are configured in:

- `<TUSK_RUNTIME_PATH>/config/mcp_config.json`

(Usually under `~/.tuskbot/config/mcp_config.json`.)

## Minimal Schema

```json
{
  "mcpServers": {
    "server-name": {
      "command": "...",
      "args": [],
      "env": {},
      "url": "...",
      "type": "http|sse",
      "headers": {}
    }
  }
}
```

## Common Examples

### Local `stdio` server

```json
{
  "mcpServers": {
    "local-sqlite": {
      "command": "uvx",
      "args": ["mcp-server-sqlite", "--db-path", "/path/to/data.db"]
    }
  }
}
```

### Remote `http` server

```json
{
  "mcpServers": {
    "remote-api": {
      "url": "https://api.example.com/mcp",
      "type": "http",
      "headers": {
        "Authorization": "Bearer <token>"
      }
    }
  }
}
```

### Remote `sse` server

```json
{
  "mcpServers": {
    "legacy-service": {
      "url": "http://localhost:8080/sse",
      "type": "sse"
    }
  }
}
```

## Transport Rules

- If `command` is set, transport is `stdio`.
- If `url` is set, `type` must be either `http` or `sse`.
- `headers` apply to remote transports.
- `env` applies to `stdio` process environments.

## Auto Reload Behavior

TuskBot watches `mcp_config.json` and applies changes live:

- new server config → connect
- changed server config → reconnect
- removed server config → disconnect

No full TuskBot restart is required.

## Verify in Chat

Use:

```text
/mcp
```

to see currently available MCP tools.

## Practical Workflow

1. Add/update server entry in `mcp_config.json`.
2. Save file.
3. Run `/mcp` in Telegram.
4. Test a tool call from chat.
