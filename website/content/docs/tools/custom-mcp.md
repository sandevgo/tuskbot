# Custom MCP Servers

TuskBot is built with an **MCP-First** philosophy. While it includes powerful native tools, its true strength lies in its ability to connect to any server implementing the Model Context Protocol (MCP).

## Creating Your Own Tools

Modern LLMs are not only trained to use MCP tools but are also exceptionally good at writing the code to create them. Using frameworks like **FastMCP** (Python) or the TypeScript MCP SDK, you can define new tools in minutes. If TuskBot lacks a specific capability, you can often ask the agent itself to write a new MCP server for you.

## Configuration & Auto-Reload

TuskBot manages external connections via a configuration file located at `~/.tuskbot/mcp_config.json`. 

One of the most convenient features is the **System Watcher**: TuskBot monitors this file for changes. As soon as you save a new server configuration or modify an existing one, the internal connection pool automatically reloads, making the new tools available to the agent instantly without a restart.

## Autonomous Setup

You don't always need to edit the configuration manually. You can:
1. **Ask the Agent**: Tell TuskBot to "Connect to the Google Maps MCP server at this URL" or "Add a new stdio server using npx".
2. **Provide Documentation**: Give the agent a link to an MCP server's description or GitHub repository. TuskBot can use its `fetch_url` tool to read the requirements and then use its management tools to set up the connection for you.

## Configuration Structure

The `mcp_config.json` follows the standard MCP host format. Here is an example showing both `stdio` (local) and `sse` (remote) configurations:

```json
{
  "mcpServers": {
    "sqlite": {
      "command": "uvx",
      "args": ["mcp-server-sqlite", "--db-path", "/path/to/my.db"]
    },
    "weather": {
      "url": "http://localhost:8080/sse",
      "type": "sse"
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "your_token_here"
      }
    }
  }
}
```
