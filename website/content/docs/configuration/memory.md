# Memory & Embeddings

TuskBot uses a local RAG pipeline for persistent memory.

## Embedding Model

Key variable:

- `TUSK_EMBEDDING_MODEL`

This file is expected at:

- `<TUSK_RUNTIME_PATH>/models/<TUSK_EMBEDDING_MODEL>`

During `tusk install`, TuskBot downloads the default embedding model (`multilingual-e5-base-q8.gguf`) if missing.

## Context Window

`TUSK_CONTEXT_WINDOW_SIZE` controls how many recent messages are included in short-term context (default: `30`).

## Storage

Vector and memory data are persisted in:

- `<TUSK_RUNTIME_PATH>/tuskbot.db`

No external vector DB is required.
