# Chat Commands

These commands are available in Telegram chat.

## Command Reference

| Command | Description | Usage |
| :--- | :--- | :--- |
| `/help` | List available commands. | `/help` |
| `/model` | Show or change active model. | `/model [provider]/[model]` |
| `/mcp` | List connected MCP tools. | `/mcp` |
| `/tasks` | List active scheduled tasks. | `/tasks` |
| `/stats` | Show session statistics. | `/stats` |

## Details

### `/help`

Lists all registered commands with short descriptions.

### `/model`

- `/model` shows current provider/model.
- `/model <provider>/<model>` switches runtime model.
- Successful updates are persisted to `.env`.

### `/mcp`

Shows currently connected tool names from native + MCP tool sources.

### `/tasks`

Lists active scheduled background tasks and next run times.

### `/stats`

Shows session ID, context token size, and message count.
