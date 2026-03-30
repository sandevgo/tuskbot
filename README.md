<div align="center">

# TuskBot 🦣

[![Build](https://img.shields.io/github/actions/workflow/status/sandevgo/tuskbot/release.yml?label=build&style=for-the-badge)](https://github.com/sandevgo/tuskbot/actions)
[![Go Report](https://goreportcard.com/badge/github.com/sandevgo/tuskbot?style=for-the-badge)](https://goreportcard.com/badge/github.com/sandevgo/tuskbot)
[![Release](https://img.shields.io/github/v/release/sandevgo/tuskbot?include_prereleases&style=for-the-badge)](https://github.com/sandevgo/tuskbot/releases)

### Personal AI Agent. MCP-First, Local RAG, Easy to Run.

**[Website](https://tuskbot.ai)** | **[Documentation](https://tuskbot.ai/docs/getting-started/)** | **[Releases](https://github.com/sandevgo/tuskbot/releases)**

</div>

## Quick Start

```shell
curl -fsSL https://tuskbot.ai/install.sh | sh
```

The installer will:
- Install the `tusk` binary
- Run `tusk install` (interactive setup)
- Install and start the system service
- Verify that the service is running

Docker guide → https://tuskbot.ai/docs/getting-started/docker

## Prerequisites

Before running the installer, have these ready:
- Telegram Bot Token
- Telegram Owner ID
- One LLM provider configured (API key or local Ollama endpoint)

## Supported Platforms

- Linux amd64
- Linux arm64
- macOS arm64

## What it does

- Runs as a self-hosted AI agent in Telegram
- Supports slash commands for quick actions
- Connects tools via MCP servers
- Handles background tasks using sub-agents
- Llama.cpp built-in for local embeddings
- Works with OpenAI, Anthropic, OpenRouter, Ollama, and compatible APIs
- Can be deployed as a binary, service, or Docker container

## Slash Commands

- `/help` — List available commands
- `/model` — Show or switch active model
- `/mcp` — List connected MCP tools
- `/tasks` — List active scheduled tasks
- `/stats` — Show session statistics

## Documentation

- Getting started → https://tuskbot.ai/docs/getting-started/
- Configuration → https://tuskbot.ai/docs/configuration/
- MCP tools → https://tuskbot.ai/docs/mcp/

## License

[MIT License](https://github.com/sandevgo/tuskbot/blob/main/LICENSE)
