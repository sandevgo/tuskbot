# MCP Servers Configuration

TuskBot uses the Model Context Protocol (MCP) to extend its capabilities with external tools.

## Configuration File
Servers are defined in `mcp_config.json` within the runtime directory.

## Transport Types
TuskBot supports the following MCP transports:
- **stdio**: For local executable tools (e.g., Node.js or Python scripts).
- **sse / http**: For remote tools connected via webhooks or streaming.

## Example Configuration
```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/workspace"]
    },
    "weather": {
      "url": "http://localhost:3000/sse",
      "type": "sse"
    }
  }
}
```

## Environment Variables
You can pass specific environment variables to `stdio` servers using the `env` key within the server's configuration object.
