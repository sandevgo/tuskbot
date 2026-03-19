# TuskBot 🦣

[![Build](https://img.shields.io/github/actions/workflow/status/sandevgo/tuskbot/release.yml?label=build&style=flat-square)](https://github.com/sandevgo/tuskbot/actions)
[![Go Report](https://goreportcard.com/badge/github.com/sandevgo/tuskbot)](https://goreportcard.com/badge/github.com/sandevgo/tuskbot)
[![Release](https://img.shields.io/github/v/release/sandevgo/tuskbot?include_prereleases&style=flat-square)](https://github.com/sandevgo/tuskbot/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sandevgo/TuskBot?style=flat-square&color=00ADD8)](https://github.com/sandevgo/tuskbot/blob/main/go.mod)
[![License](https://img.shields.io/github/license/sandevgo/TuskBot?style=flat-square)](https://github.com/sandevgo/tuskbot/LICENSE)

**Autonomous AI Agent**

TuskBot is a Go-based autonomous agent designed to handle tasks directly in your Telegram. It evolves the ideas of OpenClaw into a more scalable, secure, tool-oriented assistant.

## 🎯 Motivation

TuskBot originated as an evolution of the **OpenClaw** concept, addressing its core architectural limitations:

*   **High-Performance Engine:** Built with **Go** for native concurrency and high-speed execution, no interpreted overheads.
*   **Modular Stability:** Powered by **Model Context Protocol (MCP)**. Tool isolation prevents cascading failures and ensures system resilience.
*   **Persistent Context:** Full **Local RAG pipeline** (SQLite-vec + llama.cpp). No need to send all your chats to OpenAI for embedding.
*   **Privacy-First Design:** Native support for **Ollama** and local embedding models.

## 🚀 Capabilities

### 🔌 Extensible via MCP
TuskBot uses a **Model Context Protocol (MCP)-first** approach. This allows you to plug in any MCP-compliant server (databases, APIs, or local tools) without modifying the core logic. If a tool exists as an MCP server, TuskBot can use it.

### 🧠 Private RAG & Persistent Memory
The bot maintains a long-term memory of your interactions using a local Retrieval-Augmented Generation (RAG) pipeline:
*   **Zero-API Embeddings:** Uses **embedded llama.cpp** (via GGUF models) to process text locally. Your data for semantic search never leaves your hardware.
*   **Vector Storage:** Powered by **SQLite-vec** for fast, local retrieval of conversation history and technical context.

### 🛠️ System Access
TuskBot comes with a set of pre-configured tools for immediate use:
*   **Filesystem:** Manage, read, and write files in the bot's workspace.
*   **Shell Execution:** Run system commands and scripts directly through the chat.
*   **MCP Manager:** Allows agent to connect and restart MCP servers.

## 💾 Installation
Download the pre-compiled binary for your platform from the [Releases](https://github.com/sandevgo/tuskbot/releases) page.

**Quick Install (Linux/macOS):**

```bash
tar -xzvf tusk-*.tar.gz
chmod +x tusk-*
sudo mv tusk-* /usr/local/bin/tusk
tusk install
```

**Running TuskBot**

```bash
tusk run
```

## Using Docker

Docker compose example:

```yaml
services:
  tuskbot:
    image: ghcr.io/sandevgo/tuskbot:latest
    volumes:
      - tuskbot-data:/root/.tuskbot
    command: run

volumes:
  tuskbot-data:
```

**Running installation with Docker Compose:**

```bash
# Configure
docker compose run tuskbot install

# Run
docker compose up -d
```

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
*   `TUSK_SERVICE_SANDBOX_MODE`: Enable OS-native service hardening profile for installed services (systemd/launchd) (default: `true`).
*   `TUSK_SERVICE_LOG_DIRECTORY`: Directory for generated service stdout/stderr log files (default: `TUSK_RUNTIME_PATH`).

## 🗺 Roadmap

*   **[X] Unified Command Interface:** Support of slash-commands (`/`).
*   **[X] Cron/heartbeat:** Scheduled tasks and periodic checks.
*   **[X] Multi-Agent Orchestration:** Sub-agents to delegate specialized tasks
*   **[ ] Skills:** Skills for agents to perform specific actions.
