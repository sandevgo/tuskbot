# Runtime Data

TuskBot maintains a specific directory structure within the `TUSK_RUNTIME_PATH` to ensure persistence of state, configuration, and local models.

## Filesystem Hierarchy

Upon initialization, the system generates the following structure:

- `tuskbot.db`: SQLite database containing message history and vector embeddings.
- `mcp_config.json`: JSON-encoded configuration for external MCP servers.
- `models/`: Directory for local GGUF model storage.
- `.env`: Local environment variable overrides.
- `IDENTITY.md`: System identity prompt definition.
- `USER.md`: User profile and preference data.
- `MEMORY.md`: Long-term memory extraction rules.
