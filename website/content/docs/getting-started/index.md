# Getting Started

TuskBot is a self-hosted AI agent for Telegram.

## Prerequisites

Before setup, have these ready:

- Telegram Bot Token
- Telegram Owner ID
- One LLM provider (OpenAI, Anthropic, OpenRouter, Ollama, or OpenAI-compatible)

## Choose Your Deployment

- [Binary Installation](/docs/getting-started/installation)
- [Docker Setup](/docs/getting-started/docker)

## Typical Setup Flow

1. Install TuskBot using binary or Docker setup.
2. Run `tusk install` (automatic when using the quick installer).
3. Start the bot:
   - Binary service mode: verify with `tusk service status`
   - Foreground mode: `tusk run`
   - Docker mode: `docker compose up -d`
4. Open Telegram and run `/help`.

## Next Steps

- [CLI Reference](/docs/getting-started/cli)
- [First Steps](/docs/getting-started/first-steps)
- [Configuration Basics](/docs/getting-started/configuration)
