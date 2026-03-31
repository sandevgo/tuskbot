# LLM Providers

Set `TUSK_MAIN_MODEL` as `provider/model`.

Examples:

- `anthropic/claude-sonnet-4-5`
- `openai/gpt-4o`
- `openrouter/meta-llama/llama-3.1-70b-instruct`
- `ollama/llama3.2`
- `custom/my-model`

## Supported Provider Prefixes

| Provider | Prefix | Required Variables |
| :--- | :--- | :--- |
| Anthropic | `anthropic/` | `TUSK_ANTHROPIC_API_KEY` |
| OpenAI | `openai/` | `TUSK_OPENAI_API_KEY` |
| OpenRouter | `openrouter/` | `TUSK_OPENROUTER_API_KEY` |
| Ollama | `ollama/` | `TUSK_OLLAMA_BASE_URL` (optional `TUSK_OLLAMA_API_KEY`) |
| Custom OpenAI-compatible | `custom/` | `TUSK_CUSTOM_OPENAI_BASE_URL`, `TUSK_CUSTOM_OPENAI_API_KEY` |

## Runtime Switching

Switch model in chat:

```text
/model openai/gpt-4o
```

This updates active runtime state and persists `TUSK_MAIN_MODEL` to `.env`.

## Installer Behavior

`tusk install` helps with provider setup:

- Provider selection
- API key input (or Ollama URL)
- Model selection (auto-fetch where supported, manual fallback for Ollama/custom)
