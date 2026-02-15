# TuskBot 🦣

[![Build](https://img.shields.io/github/actions/workflow/status/sandevgo/tuskbot/release.yml?label=build&style=flat-square)](https://github.com/sandevgo/tuskbot/actions)
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

## 🗺 Roadmap

*   **[ ] Unified Command Interface:** Support of slash-commands (`/`).
*   **[ ] Cron/heartbeat:** Scheduled tasks and periodic checks.
*   **[ ] Multi-Agent Orchestration:** Sub-agents to delegate specialized tasks

## 🔧 Configuration

TuskBot uses environment variables for configuration.

| Variable | Description |
| :--- | :--- |
| `TUSK_TELEGRAM_TOKEN` | Your Telegram Bot Token |
| `TUSK_TELEGRAM_OWNER_ID` | Your Telegram User ID (for security) |
| `TUSK_CHAT_CHANNEL` | Primary chat interface (e.g., `telegram`) |
| `TUSK_MAIN_MODEL` | Main LLM model (format: `provider/model`) |
| `TUSK_EMBEDDING_MODEL_PATH` | Path to the local GGUF embedding model |
| `TUSK_RUNTIME_PATH` | Path for logs, database, and workspace (default: `~/.tuskbot`) |
| `TUSK_CONTEXT_WINDOW_SIZE` | Number of messages in active context (default: `30`) |
| `TUSK_OPENROUTER_API_KEY` | API Key for OpenRouter |
| `TUSK_OPENAI_API_KEY` | API Key for OpenAI |
| `TUSK_ANTHROPIC_API_KEY` | API Key for Anthropic |
| `TUSK_OLLAMA_BASE_URL` | Base URL for Ollama (default: `http://127.0.0.1:11434`) |
| `TUSK_DEBUG` | Enable debug logging (set to `1`) |
