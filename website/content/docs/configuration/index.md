# Configuration

TuskBot is configured primarily through environment variables, which can be managed via a `.env` file in your runtime directory.

## Configuration Methods

1. **Interactive Installer**: The `tusk install` command provides a TUI wizard to generate your initial configuration.
2. **Environment Variables**: All settings can be passed directly to the process.
3. **Dotenv File**: TuskBot automatically loads and persists variables in a `.env` file located at your `TUSK_RUNTIME_PATH`.

## Runtime Path
The `TUSK_RUNTIME_PATH` (defaulting to `~/.tuskbot`) is the central location for:
- `.env` configuration file.
- `tuskbot.db` (SQLite vector database).
- `mcp_config.json` (MCP server definitions).
- `models/` (Local GGUF embedding models).
