# Configuration Schema

TuskBot utilizes environment variables for runtime configuration. These variables are typically persisted in a `.env` file within the `TUSK_RUNTIME_PATH`.

## Core Configuration Variables

| Variable | Type | Description |
| :--- | :--- | :--- |
| `TUSK_TELEGRAM_TOKEN` | `string` | API token issued by @BotFather. |
| `TUSK_TELEGRAM_OWNER_ID` | `int64` | Telegram User ID for access control. |
| `TUSK_MAIN_MODEL` | `string` | Model identifier (e.g., `anthropic/claude-3-5-sonnet`). |
| `TUSK_EMBEDDING_MODEL` | `string` | Filename of the GGUF model in `/models` directory. |
| `TUSK_RUNTIME_PATH` | `string` | Absolute path for data persistence (Default: `~/.tuskbot`). |

## Filesystem Hierarchy

Upon initialization, the system generates the following structure in the `TUSK_RUNTIME_PATH`:

- `tuskbot.db`: SQLite database containing message history and vector embeddings.
- `mcp_config.json`: JSON-encoded configuration for external MCP servers.
- `models/`: Directory for local GGUF model storage.
- `.env`: Local environment variable overrides.
