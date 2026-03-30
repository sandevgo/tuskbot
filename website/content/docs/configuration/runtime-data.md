# Runtime Data

TuskBot stores local state under `TUSK_RUNTIME_PATH` (default: `~/.tuskbot`).

## Layout

```text
<TUSK_RUNTIME_PATH>/
├── .env
├── tuskbot.db
├── config/
│   └── mcp_config.json
├── models/
│   └── <TUSK_EMBEDDING_MODEL>
├── prompt/
│   ├── SYSTEM.md
│   ├── IDENTITY.md
│   ├── USER.md
│   ├── MEMORY.md
│   └── SUBAGENT.md
└── logs/                    # default service logs directory
```

## Bootstrap and Migration

On startup, TuskBot:

- Ensures required directories exist.
- Creates default prompt/config files if missing.
- Verifies embedding model exists at `models/<TUSK_EMBEDDING_MODEL>`.

If the embedding model is missing, startup fails and asks you to run `tusk install`.

## Database

`<TUSK_RUNTIME_PATH>/tuskbot.db` stores:

- messages and message vectors
- knowledge and knowledge vectors
- scheduled task records

## `.env`

- Loaded from runtime path at startup if present.
- Typically created by `tusk install`.
- Updated when model is changed via `/model`.
