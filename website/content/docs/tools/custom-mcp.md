# Custom MCP Servers

TuskBot is built with an **MCP-First** philosophy. While it includes powerful native tools, its true strength lies in its ability to connect to any server implementing the Model Context Protocol (MCP).

## Creating Your Own Tools

Modern LLMs are not only trained to use MCP tools but are also exceptionally good at writing the code to create them. Using frameworks like **FastMCP** (Python) or the TypeScript MCP SDK, you can define new tools in minutes. If TuskBot lacks a specific capability, you can often ask the agent itself to write a new MCP server for you.

## Configuration Specification

TuskBot implements the Model Context Protocol (MCP) host specification. External server configurations are defined in `mcp_config.json` using the following schema:

### Schema Definition

| Field | Type | Description |
| :--- | :--- | :--- |
| `mcpServers` | `object` | Map of server identifiers to `ServerConfig` objects. |
| `ServerConfig.command` | `string` | Executable name for `stdio` transport. |
| `ServerConfig.args` | `string[]` | CLI arguments for the executable. |
| `ServerConfig.env` | `object` | Key-value pairs for process environment variables. |
| `ServerConfig.url` | `string` | Endpoint for `http` or `sse` transports. |
| `ServerConfig.type` | `string` | Transport protocol: `stdio`, `http`, or `sse`. |
| `ServerConfig.headers` | `object` | HTTP headers for remote transports. |

### Implementation Example

```json
{
  "mcpServers": {
    "local-sqlite": {
      "type": "stdio",
      "command": "uvx",
      "args": ["mcp-server-sqlite", "--db-path", "/path/to/data.db"]
    },
    "remote-api": {
      "type": "http",
      "url": "https://api.example.com/mcp",
      "headers": {
        "Authorization": "Bearer <token>"
      }
    },
    "legacy-service": {
      "type": "sse",
      "url": "http://localhost:8080/sse"
    }
  }
}
```

> [!IMPORTANT]
> The `sse` transport is maintained for backward compatibility. New implementations should utilize the `http` transport for remote tool execution.

## Configuration & Auto-Reload

TuskBot manages external connections via a configuration file located at `~/.tuskbot/mcp_config.json`. 

One of the most convenient features is the **System Watcher**: TuskBot monitors this file for changes. As soon as you save a new server configuration or modify an existing one, the internal connection pool automatically reloads, making the new tools available to the agent instantly without a restart.

## Autonomous Setup

You don't always need to edit the configuration manually. You can:
1. **Ask the Agent**: Tell TuskBot to "Connect to the Google Maps MCP server at this URL" or "Add a new stdio server using npx".
2. **Provide Documentation**: Give the agent a link to an MCP server's description or GitHub repository. TuskBot can use its `fetch_url` tool to read the requirements and then use its management tools to set up the connection for you.
