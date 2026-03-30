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

Docker guide → https://tuskbot.ai/docs/getting-started/docker

## What it does

- Runs as a self-hosted AI agent in Telegram
- Connects tools via MCP servers
- Handles background tasks using sub-agents
- Supports local memory and semantic search
- Works with OpenAI, Anthropic, OpenRouter, Ollama, and compatible APIs
- Can be deployed as a binary, service, or Docker container

## Documentation
- Getting started → https://tuskbot.ai/docs/getting-started/
- Configuration → https://tuskbot.ai/docs/configuration/
- MCP tools → https://tuskbot.ai/docs/mcp/

## License

MIT

## ⌨️ Slash Commands

TuskBot supports the following slash commands for direct interaction:

- **/model** Display/Switch the currently active LLM provider and model.
- **/mcp** List all currently connected MCP servers and their available tools.

## 🔧 Configuration

TuskBot uses environment variables for configuration.

### Core Settings

*   `TUSK_TELEGRAM_TOKEN`: Your Telegram Bot Token.
*   `TUSK_TELEGRAM_OWNER_ID`: Your Telegram User ID (for security).
*   `TUSK_CHAT_CHANNEL`: Primary chat interface (e.g., `telegram`).
*   `TUSK_RUNTIME_PATH`: Path for logs, database, and workspace (default: `~/.tuskbot`).
*   `TUSK_DEBUG`: Enable debug logging (set to `1`).

### AI & Memory

*   `TUSK_MAIN_MODEL`: Main LLM model (format: `provider/model`).
*   `TUSK_EMBEDDING_MODEL`: Embedding model file name (gguf).
*   `TUSK_CONTEXT_WINDOW_SIZE`: Number of messages in active context (default: `30`).

### Providers

*   `TUSK_OPENROUTER_API_KEY`: API Key for OpenRouter.
*   `TUSK_OPENAI_API_KEY`: API Key for OpenAI.
*   `TUSK_ANTHROPIC_API_KEY`: API Key for Anthropic.
*   `TUSK_OLLAMA_BASE_URL`: Base URL for Ollama (default: `http://127.0.0.1:11434`).
*   `TUSK_OLLAMA_API_KEY`: API Key for Ollama (optional).
*   `TUSK_CUSTOM_OPENAI_BASE_URL`: Base URL for Custom OpenAI provider.
*   `TUSK_CUSTOM_OPENAI_API_KEY`: API Key for Custom OpenAI provider.

### System Service

*   `TUSK_SERVICE_USER_MODE`: Install service in user mode by default (default: `true`).
*   `TUSK_SERVICE_LOG_DIRECTORY`: Directory for generated service stdout/stderr log files (default: `TUSK_RUNTIME_PATH`).
