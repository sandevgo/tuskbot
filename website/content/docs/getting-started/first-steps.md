# Usage

How to interact with TuskBot once it is running.

## Starting the Conversation

After running `tusk start`, open Telegram and find your bot. Send a message to begin. 

**Security Note**: TuskBot uses `TUSK_TELEGRAM_OWNER_ID` to ensure it only responds to you. Messages from any other user will be ignored.

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
