# Docker Setup

Run TuskBot with Docker Compose.

## Compose File

```yaml
services:
  tuskbot:
    image: ghcr.io/sandevgo/tuskbot:latest
    restart: unless-stopped
    volumes:
      - tuskbot_data:/root/.tuskbot
    command: run

volumes:
  tuskbot_data:
```

Use `command: run` in containers (foreground process).

## First-Time Setup

Run the interactive installer once to create configuration in the mounted runtime path:

```bash
docker compose run --rm -it tuskbot install
```

## Start and Observe

```bash
docker compose up -d
docker compose logs -f tuskbot
```

## Runtime Data and Persistence

The `/root/.tuskbot` volume stores:

- `.env`
- `tuskbot.db`
- `config/mcp_config.json`
- `models/` (local embedding model files)

Keep this volume persistent between restarts.

## Environment Overrides

Variables passed through Docker Compose (`environment` or `env_file`) override values from the internal `.env` file.
