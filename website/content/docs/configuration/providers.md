# LLM Providers Configuration

TuskBot supports multiple AI providers. The active model is defined by the `TUSK_MAIN_MODEL` variable using the format `provider/model`.

## Supported Providers

### OpenAI
- **Variable**: `TUSK_OPENAI_API_KEY`
- **Example**: `TUSK_MAIN_MODEL=openai/gpt-4o`

### Anthropic
- **Variable**: `TUSK_ANTHROPIC_API_KEY`
- **Example**: `TUSK_MAIN_MODEL=anthropic/claude-3-5-sonnet-20240620`

### OpenRouter
- **Variable**: `TUSK_OPENROUTER_API_KEY`
- **Example**: `TUSK_MAIN_MODEL=openrouter/meta-llama/llama-3.1-405b`

### Ollama (Local)
- **Variables**: `TUSK_OLLAMA_BASE_URL` (default: `http://127.0.0.1:11434`), `TUSK_OLLAMA_API_KEY` (optional).
- **Example**: `TUSK_MAIN_MODEL=ollama/llama3`

### Custom OpenAI Compatible
- **Variables**: `TUSK_CUSTOM_OPENAI_BASE_URL`, `TUSK_CUSTOM_OPENAI_API_KEY`.
- **Example**: `TUSK_MAIN_MODEL=custom/my-model`
