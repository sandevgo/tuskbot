# Environment Variables

TuskBot utilizes environment variables for runtime configuration. These variables are typically persisted in a `.env` file within the `TUSK_RUNTIME_PATH`.

## Core System Configuration

| Variable | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `TUSK_RUNTIME_PATH` | `string` | No | `~/.tuskbot` | Absolute path for data persistence and configuration. |
| `TUSK_CHAT_CHANNEL` | `string` | Yes | - | Communication interface (e.g., `telegram`). |
| `TUSK_CONTEXT_WINDOW_SIZE` | `int` | No | `30` | Number of recent messages to include in LLM context. |
| `TUSK_DEBUG` | `bool` | No | `0` | Set to `1` to enable verbose debug logging. |

## AI Provider Configuration

| Variable | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `TUSK_MAIN_MODEL` | `string` | Yes | Primary LLM identifier (format: `provider/model`). |
| `TUSK_EMBEDDING_MODEL` | `string` | Yes | Filename of the GGUF model in `/models` directory. |

### Provider Credentials

| Variable | Type | Description |
| :--- | :--- | :--- |
| `TUSK_ANTHROPIC_API_KEY` | `string` | API key for Anthropic Claude models. |
| `TUSK_OPENAI_API_KEY` | `string` | API key for OpenAI GPT models. |
| `TUSK_OPENROUTER_API_KEY` | `string` | API key for OpenRouter.ai. |
| `TUSK_OLLAMA_API_KEY` | `string` | API key for Ollama (if required by proxy). |
| `TUSK_OLLAMA_BASE_URL` | `string` | Endpoint for Ollama (Default: `http://127.0.0.1:11434`). |
| `TUSK_CUSTOM_OPENAI_BASE_URL` | `string` | Base URL for OpenAI-compatible providers. |
| `TUSK_CUSTOM_OPENAI_API_KEY` | `string` | API key for OpenAI-compatible providers. |

## Telegram Transport Configuration

| Variable | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `TUSK_TELEGRAM_TOKEN` | `string` | Yes | Bot API token issued by @BotFather. |
| `TUSK_TELEGRAM_OWNER_ID` | `int64` | Yes | Numeric Telegram User ID for exclusive access control. |

## System Service Configuration

| Variable | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `TUSK_SERVICE_NAME` | `string` | No | `tuskbot` | Service unit/label name. |
| `TUSK_SERVICE_DISPLAY_NAME` | `string` | No | `TuskBot` | Human-readable service name. |
| `TUSK_SERVICE_DESCRIPTION` | `string` | No | `TuskBot background agent service` | Service description shown by the init system. |
| `TUSK_SERVICE_USER_MODE` | `bool` | No | `true` | Install service as current-user service by default. |
| `TUSK_SERVICE_SANDBOX_MODE` | `bool` | No | `true` | Enables service-manager hardening profile (`SystemdScript`/`LaunchdConfig`) for installed services. |
| `TUSK_SERVICE_LOG_DIRECTORY` | `string` | No | `TUSK_RUNTIME_PATH` | Directory for generated service stdout/stderr logs in service definitions. |

Service containment is enforced at the OS service manager layer (systemd/launchd), not by tool-level filesystem path scoping.
