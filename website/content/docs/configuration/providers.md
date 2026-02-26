# LLM Providers Specification

TuskBot supports multiple Large Language Model (LLM) providers through a unified interface. Providers are configured via the `TUSK_MAIN_MODEL` variable using the `provider/model` identifier format.

## Supported Providers

| Provider | Identifier | Required Variable |
| :--- | :--- | :--- |
| **Anthropic** | `anthropic/` | `TUSK_ANTHROPIC_API_KEY` |
| **OpenAI** | `openai/` | `TUSK_OPENAI_API_KEY` |
| **OpenRouter** | `openrouter/` | `TUSK_OPENROUTER_API_KEY` |
| **Ollama** | `ollama/` | `TUSK_OLLAMA_BASE_URL` |
| **Custom OpenAI** | `custom/` | `TUSK_CUSTOM_OPENAI_BASE_URL` |

## Configuration Details

### Anthropic
Direct integration with Anthropic's Messages API.
- **Identifier Example**: `anthropic/claude-3-5-sonnet-latest`
- **Authentication**: Requires a valid API key from the Anthropic Console.

### OpenAI
Direct integration with OpenAI's Chat Completions API.
- **Identifier Example**: `openai/gpt-4o`
- **Authentication**: Requires a valid API key from the OpenAI Platform.

### OpenRouter
Unified interface for accessing various models (Llama, Mistral, Gemini, etc.) via OpenRouter.ai.
- **Identifier Example**: `openrouter/meta-llama/llama-3.1-405b`
- **Authentication**: Requires an OpenRouter API key.

### Ollama
Local inference via the Ollama API.
- **Identifier Example**: `ollama/llama3.1`
- **Endpoint**: Defaults to `http://127.0.0.1:11434`. Override using `TUSK_OLLAMA_BASE_URL`.
- **Note**: Ensure the model is pulled locally (`ollama pull <model>`) before instantiation.

### Custom OpenAI-Compatible
Generic provider for any service implementing the OpenAI Chat Completions schema (e.g., vLLM, LocalAI, Groq).
- **Identifier Example**: `custom/my-model-name`
- **Configuration**: Must define `TUSK_CUSTOM_OPENAI_BASE_URL` and `TUSK_CUSTOM_OPENAI_API_KEY`.

## Runtime Model Switching

The active provider and model can be modified during runtime without service interruption using the `/model` command.

```text
/model anthropic/claude-3-5-sonnet
```

> [!WARNING]
> Switching to a provider without a configured API key will result in a `401 Unauthorized` or `403 Forbidden` error during the next inference cycle.
