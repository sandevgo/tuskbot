# Docker Setup

Running TuskBot with Docker.

## Docker Compose

You can use the following `docker-compose.yml` to run TuskBot. Note that for the first run, you should use the `install` command to set up your environment.

```yaml
services:
  tuskbot:
    image: ghcr.io/sandevgo/tuskbot:latest
    volumes:
      - tuskbot-data:/root/.tuskbot
    command: start

volumes:
  tuskbot-data:
```

## Essential Variables

When using Docker, you must at least provide these variables to boot:
- `TUSK_TELEGRAM_TOKEN`
- `TUSK_TELEGRAM_OWNER_ID`
- `TUSK_MAIN_MODEL`
- `TUSK_EMBEDDING_MODEL`

## Volume Mounts

The volume `tuskbot-data` is mapped to `/root/.tuskbot`. This ensures that your: database and runtime data are persisted across container restarts.

## Building from Dockerfile

If you are building locally:
```bash
docker build -t tuskbot:latest .
docker compose up -d
```
