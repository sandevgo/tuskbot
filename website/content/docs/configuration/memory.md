# Memory & Embeddings

TuskBot features a local RAG (Retrieval-Augmented Generation) pipeline for persistent memory.

## Local Embeddings
TuskBot uses `llama.cpp` to run embedding models locally.
- `TUSK_EMBEDDING_MODEL`: The filename of the GGUF model (e.g., `nomic-embed-text-v1.5.Q8_0.gguf`).
- Models must be placed in the `models/` directory within your runtime path.

## Context Management
- `TUSK_CONTEXT_WINDOW_SIZE`: Defines how many recent messages are included in the short-term conversation history (default: `30`).

## Vector Storage
The system uses `sqlite-vec` for high-performance local vector search. This is stored in `tuskbot.db` and requires no manual configuration.
