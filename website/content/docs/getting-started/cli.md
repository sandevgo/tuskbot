# CLI Reference

TuskBot is managed via a single binary with two primary operational commands.

## tusk install

Initiates an interactive TUI wizard to provision the local environment and generate configuration artifacts.

### Usage
```bash
tusk install
```

### Functional Scope
- **AI Provider**: Configures LLM credentials and model identifiers.
- **RAG Setup**: Downloads GGUF embedding models to the local `models/` directory.
- **Telegram**: Configures Bot API tokens and Owner ID verification.

---

## tusk start

Instantiates the service lifecycle, including transports, providers, and background RAG workers.

### Usage
```bash
tusk start [flags]
```

### Flags
| Flag | Shorthand | Description |
| :--- | :--- | :--- |
| `--debug` | `-d` | Enables verbose debug logging. |

### Service Sequence
1. **Environment**: Loads `.env` from the runtime directory.
2. **Workers**: Starts Knowledge Extractor and Message Embedder.
3. **Transports**: Activates the Telegram Bot API polling loop.
