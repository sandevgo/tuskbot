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

## Model Management

The `/model` command allows for hot-swapping the underlying LLM without restarting the service.

- **View Current**: Invoke `/model` without arguments to see the active provider and model.
- **Switch Model**: Provide the model identifier in `provider/model` format.
  - Example: `/model anthropic/claude-3-5-sonnet`
  - Example: `/model openai/gpt-4o`

## MCP Inspection

The `/mcp` command provides visibility into the tool ecosystem currently available to the agent.

## Task & Session Inspection

- **`/task`**: Lists active one-time and recurring scheduled tasks.
- **`/stats`**: Shows the current session ID, context size in tokens, and total messages in context.
- **`/help`**: Prints the command list with descriptions.
