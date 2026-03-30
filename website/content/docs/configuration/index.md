# Configuration

TuskBot is configured with environment variables, usually stored in a runtime `.env` file.

## Recommended Setup Path

1. Run `tusk install` to generate initial configuration.
2. Start TuskBot (`tusk run` or service mode).
3. Adjust variables in your runtime `.env` when needed.

## Runtime Location

By default, runtime data is stored in `~/.tuskbot`.

Set `TUSK_RUNTIME_PATH` to use a different location.

## What to Read Next

- [Environment Variables](/docs/configuration/environment)
- [LLM Providers](/docs/configuration/providers)
- [Telegram](/docs/configuration/telegram)
- [MCP Servers](/docs/configuration/mcp-servers)
- [Memory & Embeddings](/docs/configuration/memory)
- [Chat Commands](/docs/configuration/commands)
- [Runtime Data](/docs/configuration/runtime-data)
