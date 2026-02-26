# Docker Setup

Running TuskBot with Docker.

## Docker Compose

You can use the following `docker-compose.yml` to run TuskBot:

```yaml
services:
  tuskbot:
    image: tuskbot:latest
    volumes:
      - tuskbot-data:/root/.tuskbot
    command: start

volumes:
  tuskbot-data:
```

## Environment Variables

When using Docker, pass your configuration via the `environment` section or an `.env` file:
- `TUSK_TELEGRAM_TOKEN`
- `TUSK_TELEGRAM_OWNER_ID`
- `TUSK_MAIN_MODEL`

## Volume Mounts

The volume `tuskbot-data` is mapped to `/root/.tuskbot`. This ensures that your:
- SQLite database
- Local embedding models
- Logs
- Workspace files
are persisted across container restarts.

## Building from Dockerfile

If you are building locally:
```bash
docker build -t tuskbot:latest .
docker compose up -d
```
