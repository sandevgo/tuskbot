# Chat Commands

TuskBot provides a set of administrative commands to manage the agent's state and inspect connected services directly from the chat interface.

## Command Reference

| Command | Description | Usage |
| :--- | :--- | :--- |
| `/model` | View or update the active LLM provider and model. | `/model [provider]/[model]` |
| `/mcp` | List MCP tools currently available to the agent. | `/mcp` |
| `/task` | List active scheduled background tasks and their next run time. | `/task` |
| `/stats` | Show session statistics (context tokens and message count). | `/stats` |
| `/help` | Display all available commands and basic usage instructions. | `/help` |

## Command Behavior

### `/model`

- Run `/model` with no arguments to display current provider and model.
- Run `/model <provider>/<model>` to switch the active model.
- Expect persistent configuration updates after successful model change.

### `/mcp`

- Run `/mcp` to list MCP tools currently available to the agent.

### `/task`

- Run `/task` to list active scheduled tasks.
- Read `next run` values as formatted strings (`YYYY-MM-DD HH:MM`).

### `/stats`

- Run `/stats` to inspect session statistics.
- Read the output fields:
    - `Session ID: string`
    - `Context Size: string` (token count)
    - `Messages: string` (count)

### `/help`

- Run `/help` to list all registered commands with descriptions.
