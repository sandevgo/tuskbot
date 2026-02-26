# Configuration

Basic configuration for TuskBot.

## Environment Variables

TuskBot uses environment variables for core settings:
- `TUSK_TELEGRAM_TOKEN`: Your Telegram Bot Token.
- `TUSK_TELEGRAM_OWNER_ID`: Your Telegram User ID.
- `TUSK_RUNTIME_PATH`: Path for data (default: `~/.tuskbot`).
- `TUSK_MAIN_MODEL`: Format `provider/model` (e.g., `openai/gpt-4o`).

## Configuration File

The `tusk install` command generates a configuration environment in your runtime path. You can manually edit these values in your shell profile or a `.env` file.

## First Run

On the first run, TuskBot will initialize the directory structure in `~/.tuskbot`, including:
- `tuskbot.db`: The SQLite vector database.
- `mcp.json`: MCP server configurations.
- `models/`: Local GGUF embedding models.
