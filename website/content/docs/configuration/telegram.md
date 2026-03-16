# Telegram Setup

TuskBot uses the Telegram Bot API as its primary interface.

## Required Settings

- `TUSK_TELEGRAM_TOKEN`: The API token obtained from [@BotFather](https://t.me/botfather).
- `TUSK_TELEGRAM_OWNER_ID`: Your numeric Telegram User ID. TuskBot implements a strict security policy where it only responds to this specific ID.

## Chat Channel
The variable `TUSK_CHAT_CHANNEL` must be set to `telegram` to enable the Telegram transport layer.

## Bot Menu
It is recommended to configure the following commands in BotFather for quick access:
- `/model` - View or change the current LLM provider.
- `/mcp` - List currently available MCP tools.
- `/task` - List active scheduled tasks.
- `/stats` - Show current session statistics.
- `/help` - Show all available commands.
