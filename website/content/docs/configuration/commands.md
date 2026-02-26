# Chat Commands

TuskBot provides a set of administrative commands to manage the agent's state and inspect connected services directly from the chat interface.

## Command Reference

| Command | Description | Usage |
| :--- | :--- | :--- |
| `/model` | View or update the active LLM provider and model. | `/model [provider]/[model]` |
| `/mcp` | List all active MCP servers and their registered tool definitions. | `/mcp` |
| `/reset` | Clear the current session's short-term memory (context). | `/reset` |
| `/help` | Display available commands and basic usage instructions. | `/help` |

## Model Management

The `/model` command allows for hot-swapping the underlying LLM without restarting the service.

- **View Current**: Invoke `/model` without arguments to see the active provider and model.
- **Switch Model**: Provide the model identifier in `provider/model` format.
  - Example: `/model anthropic/claude-3-5-sonnet`
  - Example: `/model openai/gpt-4o`

## MCP Inspection

The `/mcp` command provides visibility into the tool ecosystem currently available to the agent. It enumerates:
- Connected MCP server names.
- Status of the connection pool.
- List of functions provided by each server.
