# MCP Servers

TuskBot extends tool access via MCP servers.

## Config File

MCP server definitions are stored at:

- `<TUSK_RUNTIME_PATH>/config/mcp_config.json`

## Schema

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/workspace"],
      "env": {
        "NODE_ENV": "production"
      }
    },
    "weather": {
      "url": "http://localhost:3000/sse",
      "type": "sse",
      "headers": {
        "Authorization": "Bearer <token>"
      }
    }
  }
}
```

## Transports

TuskBot supports:

- `stdio` (when `command` is set)
- `sse` (when `url` is set with `type: "sse"`)
- `http` (when `url` is set with `type: "http"`)

If `url` is set, `type` must be either `http` or `sse`.

## Runtime Behavior

- MCP config is loaded at startup.
- TuskBot watches `mcp_config.json` and applies updates automatically.
- Added/changed/removed servers are connected/restarted/removed without full process restart.

## In Chat

Use:

```text
/mcp
```

to view currently available MCP tools.
