# System Introduction

TuskBot is a high-concurrency autonomous agent engine implemented in Go. It facilitates secure, tool-augmented interactions via the Telegram Bot API, utilizing the Model Context Protocol (MCP) for extensible capability management.

## Core Architecture Components

| Component | Implementation | Function |
| :--- | :--- | :--- |
| **Runtime** | Go (Golang) | Core logic and service orchestration. |
| **Persistence** | SQLite-vec | Vectorized message history and knowledge storage. |
| **Embeddings** | llama.cpp (GGUF) | Local inference for RAG pipeline operations. |
| **Extensibility** | MCP | Standardized interface for external tool integration. |

## Deployment Workflow

1. **Provisioning**: Execute `tusk install` to initialize the environment.
2. **Execution**: Invoke `tusk start` to instantiate the service lifecycle.
