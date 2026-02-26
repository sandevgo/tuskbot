# Containerized Deployment

TuskBot is distributed as a Docker image for containerized environments.

## Orchestration Specification

Utilize the following `docker-compose.yml` definition for persistent deployment.

```yaml
services:
  tuskbot:
    image: ghcr.io/sandevgo/tuskbot:latest
    restart: unless-stopped
    env_file: .env
    volumes:
      - tuskbot_data:/root/.tuskbot
    command: start

volumes:
  tuskbot_data:
```

## Mandatory Environment Variables

The container requires the following variables for successful instantiation:

| Variable | Description |
| :--- | :--- |
| `TUSK_TELEGRAM_TOKEN` | Telegram Bot API authentication token. |
| `TUSK_TELEGRAM_OWNER_ID` | Numeric identifier for the authorized user. |
| `TUSK_MAIN_MODEL` | LLM identifier (format: `provider/model`). |
| `TUSK_EMBEDDING_MODEL` | Filename of the local GGUF embedding model. |

## Volume Persistence
The `/root/.tuskbot` mount point contains the SQLite database (`tuskbot.db`), the MCP configuration (`mcp_config.json`), and the local model cache. Ensure the host volume has sufficient write permissions.
