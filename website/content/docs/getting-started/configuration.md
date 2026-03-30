# Configuration Basics

`tusk install` creates and manages your initial runtime configuration.

## Runtime Path

By default, TuskBot stores runtime data in `~/.tuskbot`.

You can override this with:

- `TUSK_RUNTIME_PATH`

## What Is Stored There

- `.env` (runtime configuration)
- `tuskbot.db` (SQLite database)
- `config/mcp_config.json` (MCP server config)
- `models/` (local embedding models)
- `prompt/` (system and profile prompt files)

## Common Variables

| Variable | Purpose |
| :--- | :--- |
| `TUSK_TELEGRAM_TOKEN` | Telegram bot token |
| `TUSK_TELEGRAM_OWNER_ID` | Allowed Telegram user ID |
| `TUSK_CHAT_CHANNEL` | Active chat transport (`telegram`) |
| `TUSK_MAIN_MODEL` | Primary LLM model (`provider/model`) |
| `TUSK_EMBEDDING_MODEL` | Embedding model filename |
| `TUSK_RUNTIME_PATH` | Runtime data directory |

## Full Configuration Reference

- [Environment Variables](/docs/configuration/environment)
- [LLM Providers](/docs/configuration/providers)
- [Chat Commands](/docs/configuration/commands)
- [Runtime Data](/docs/configuration/runtime-data)
