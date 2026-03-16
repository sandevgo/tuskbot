# Telegram Transport Specification

Configure Telegram as the active chat transport.

## Required Environment Variables

| Variable | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `TUSK_TELEGRAM_TOKEN` | `string` | Yes | Bot API token issued by `@BotFather`. |
| `TUSK_TELEGRAM_OWNER_ID` | `int64` | Yes | Numeric Telegram user ID authorized to interact with the bot. |
| `TUSK_CHAT_CHANNEL` | `string` | Yes | Must be set to `telegram` to enable Telegram transport. |

## Access Control

- Set `TUSK_TELEGRAM_OWNER_ID` to a single trusted user ID.
- Send messages only from that account.
- Expect all other senders to be ignored.

## Bot Command Menu

Tusk registers it's slash commands with Telegram's Bot API.
