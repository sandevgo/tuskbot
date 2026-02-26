# Essential Config

TuskBot is designed to be configured via the [Interactive Installer](./installation.md), but you can also set environment variables manually.

## Required Variables

To get the bot online, you need these four core settings:

- `TUSK_TELEGRAM_TOKEN`: Your Telegram Bot Token from @BotFather.
- `TUSK_TELEGRAM_OWNER_ID`: Your numeric Telegram User ID (the bot only responds to you).
- `TUSK_MAIN_MODEL`: The LLM to use, formatted as `provider/model` (e.g., `openai/gpt-4o`).
- `TUSK_EMBEDDING_MODEL`: The filename of the local GGUF embedding model (e.g., `all-MiniLM-L6-v2-Q8_0.gguf`).

## The .env File

The `tusk install` command generates a `.env` file in your `TUSK_RUNTIME_PATH` (default `~/.tuskbot`). TuskBot automatically loads variables from this file on startup.

For a complete list of all available settings, see the [Full Variable List](../configuration/index.md).

## First Run

On the first run, TuskBot will initialize the directory structure in `~/.tuskbot`, including:
- `tuskbot.db`: The SQLite vector database.
- `mcp_config.json`: MCP server configurations.
- `models/`: Local GGUF embedding models.
