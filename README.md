# TuskBot 🦣

![Go Version](https://img.shields.io/github/go-mod/go-version/sandevgo/TuskBot)
![Stage](https://img.shields.io/badge/stage-MVP-blue)
![MCP](https://img.shields.io/badge/MCP-compliant-green)
![SQLite](https://img.shields.io/badge/storage-SQLite--vec-blue)
![License](https://img.shields.io/github/license/sandevgo/TuskBot)

**Autonomous AI Agent**

TuskBot is a Go-based autonomous agent designed to handle tasks directly in your Telegram. It evolves the ideas of OpenClaw into a more scalable, tool-oriented assistant. Unlike standard bots, TuskBot is built to interact with your system, remember context through local RAG, and connect to any external service via the Model Context Protocol (MCP).

## 🎯 Motivation

TuskBot originated as an evolution of the **OpenClaw** concept, addressing its core architectural limitations:

*   **Performance:** Transitioning from JavaScript to **Go** provides the concurrency and execution speed required for intensive agentic tasks.
*   **Stability via Isolation:** Unlike OpenClaw’s integrated scripts—which are prone to crashing the entire process—TuskBot uses a **Model Context Protocol (MCP)** approach. Tools are isolated, making the system modular and resilient.
*   **Reliable Memory (RAG vs. Files):** Instead of relying on fragile daily log files and summarization (which quickly bloat the context window), TuskBot implements a **full RAG pipeline**. Using vector embeddings (SQLite-vec), it retrieves only relevant context, allowing the bot to "remember" details from weeks ago without overflowing the context window.
*   **Safe Self-Improvement:** OpenClaw’s self-improvement often leads to recursive code corruption. TuskBot enables **safe evolution** by allowing the agent to extend its own capabilities through connecting or generating new MCP-compliant servers, keeping the core logic intact.

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
*   **Web Fetch:** Pull and parse content from URLs for analysis.

## 🛠 Tech Stack
*   **Language:** Go (Golang)
*   **Inference:** `llama.cpp` (embedded)
*   **Storage:** SQLite (Metadata & Vectors)
*   **Connectivity:** OpenRouter (LLM), Telegram (Bot API), MCP (Tools)

## 🗺 Roadmap

*   **[ ] Unified Command Interface:** Support of slash-commands (`/`).
*   **[ ] Multi-Agent Orchestration:** Sub-agents to delegate specialized tasks

## 🔧 Configuration

TuskBot uses environment variables for configuration.

| Variable | Description                                          |
| :--- |:-----------------------------------------------------|
| `TUSK_TELEGRAM_TOKEN` | Your Telegram Bot Token                              |
| `TUSK_TELEGRAM_OWNER_ID` | Your Telegram User ID (for security)                 |
| `TUSK_OPENROUTER_API_KEY` | API Key for OpenRouter                               |
| `TUSK_RAG_MODEL_PATH` | Path to the local GGUF embedding model               |
| `TUSK_MAIN_MODEL` | Main LLM model (format: `provider/model`)            |
| `TUSK_RUNTIME_PATH` | Path for logs, database, and workspace               |
| `TUSK_CONTEXT_WINDOW_SIZE`| Number of messages in active context (default: `30`) |
| `TUSK_ENABLE_TELEGRAM` | Enable/Disable Telegram transport                    |
| `TUSK_DEBUG` | Enable debug logging (set to `1`)                    |
