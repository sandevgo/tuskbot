# First Steps

Getting started with TuskBot after installation.

## Initial Setup

After running `tusk start`, open your Telegram client and start a chat with your bot. Since you configured your `TUSK_TELEGRAM_OWNER_ID`, the bot will only respond to you.

## Slash Commands

Interact with the bot using these direct commands:
- `/model`: Display or switch the active LLM provider and model.
- `/mcp`: List connected MCP servers and their available tools.

## Testing Your Setup

Try the following prompts to verify functionality:
1. **Check Tools**: "List the files in the current directory."
2. **Check RAG**: "Remember that my favorite programming language is Go." Then ask later: "What is my favorite language?"
3. **Check MCP**: Use a command specific to a connected MCP server.

## Next Steps

- [Configure MCP Servers](../configuration/mcp-servers.md)
- [Learn about Architecture](../architecture/index.md)
