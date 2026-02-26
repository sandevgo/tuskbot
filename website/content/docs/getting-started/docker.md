# Containerized Deployment

TuskBot is distributed as a Docker image for containerized environments.

## Orchestration Specification

Utilize the following `docker-compose.yml` definition for persistent deployment.

```yaml
services:
  tuskbot:
    image: ghcr.io/sandevgo/tuskbot:latest
    restart: unless-stopped
    volumes:
      - tuskbot_data:/root/.tuskbot
    command: start

volumes:
  tuskbot_data:
```

## Configuration Management

TuskBot generates a `.env` configuration file within the `/root/.tuskbot` directory upon successful execution of the `tusk install` command. 

### Environment Overrides
While the system prioritizes the internal `.env` file, variables defined in the Docker `environment` block or via an `env_file` in `docker-compose.yml` will override the internal file specifications.

## Volume Persistence
The `/root/.tuskbot` mount point contains the SQLite database (`tuskbot.db`), the MCP configuration (`mcp_config.json`), and the local model cache. Ensure the host volume has sufficient write permissions.
