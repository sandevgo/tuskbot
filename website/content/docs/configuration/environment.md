# Environment Variables

TuskBot reads configuration from environment variables.

In normal usage, these values are stored in `<TUSK_RUNTIME_PATH>/.env` and loaded on startup.

## Core Runtime

| Variable | Required | Default | Description |
| :--- | :--- | :--- | :--- |
| `TUSK_RUNTIME_PATH` | No | `~/.tuskbot` | Runtime data directory. |
| `TUSK_CHAT_CHANNEL` | Yes | - | Active transport (currently `telegram`). |
| `TUSK_CONTEXT_WINDOW_SIZE` | No | `30` | Number of recent messages in short-term context. |
| `TUSK_DEBUG` | No | `0` | Set `1` to enable debug logging. |

## Model and Embeddings

| Variable | Required | Description |
| :--- | :--- | :--- |
| `TUSK_MAIN_MODEL` | Yes | Main model in `provider/model` format. |
| `TUSK_EMBEDDING_MODEL` | Yes | Embedding model filename inside `models/`. |

## Provider Credentials

Set credentials for the provider used by `TUSK_MAIN_MODEL`.

| Variable | Description |
| :--- | :--- |
| `TUSK_ANTHROPIC_API_KEY` | Anthropic API key |
| `TUSK_OPENAI_API_KEY` | OpenAI API key |
| `TUSK_OPENROUTER_API_KEY` | OpenRouter API key |
| `TUSK_OLLAMA_BASE_URL` | Ollama base URL (default `http://127.0.0.1:11434`) |
| `TUSK_OLLAMA_API_KEY` | Optional Ollama API key |
| `TUSK_CUSTOM_OPENAI_BASE_URL` | OpenAI-compatible base URL |
| `TUSK_CUSTOM_OPENAI_API_KEY` | OpenAI-compatible API key |

## Telegram

| Variable | Required | Description |
| :--- | :--- | :--- |
| `TUSK_TELEGRAM_TOKEN` | Yes | Telegram bot token from `@BotFather` |
| `TUSK_TELEGRAM_OWNER_ID` | Yes | Telegram user ID allowed to use the bot |

## Service Mode

These variables affect `tusk service ...` commands.

| Variable | Required | Default | Description |
| :--- | :--- | :--- | :--- |
| `TUSK_SERVICE_NAME` | No | `tuskbot` | Service name/unit label |
| `TUSK_SERVICE_DISPLAY_NAME` | No | `TuskBot` | Human-readable service name |
| `TUSK_SERVICE_DESCRIPTION` | No | `TuskBot background agent service` | Service description |
| `TUSK_SERVICE_USER_MODE` | No | `true` | Install service for current user by default |
| `TUSK_SERVICE_LOG_DIRECTORY` | No | `<TUSK_RUNTIME_PATH>/logs` | Service stdout/stderr log directory |
| `TUSK_SERVICE_PATH` | No | auto-built | PATH used by service processes |

## Notes

- `tusk install` writes `.env` with file mode `0600`.
- Changing model via `/model` persists `TUSK_MAIN_MODEL` back to `.env`.
