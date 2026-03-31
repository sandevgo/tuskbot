# Telegram

Configure Telegram as the active chat channel.

## Required Variables

| Variable | Required | Description |
| :--- | :--- | :--- |
| `TUSK_CHAT_CHANNEL` | Yes | Must be `telegram` |
| `TUSK_TELEGRAM_TOKEN` | Yes | Bot token from `@BotFather` |
| `TUSK_TELEGRAM_OWNER_ID` | Yes | Telegram user ID allowed to use the bot |

## Access Control

TuskBot only processes messages from `TUSK_TELEGRAM_OWNER_ID`.

Messages from other users are ignored.

## Bot Commands

On startup, TuskBot registers available slash commands for the owner in private chats.

Use `/help` to view the current command list.
