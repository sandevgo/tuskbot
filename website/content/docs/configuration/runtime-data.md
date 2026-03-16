# Runtime Data

TuskBot stores local state under `TUSK_RUNTIME_PATH` (default: `~/.tuskbot`).

## Runtime Layout

```text
<TUSK_RUNTIME_PATH>/
├── .env
├── tuskbot.db
├── config/
│   └── mcp_config.json
├── models/
│   └── <TUSK_EMBEDDING_MODEL>
└── prompt/
    ├── SYSTEM.md
    ├── IDENTITY.md
    ├── USER.md
    ├── MEMORY.md
    └── SUBAGENT.md
```

## Bootstrap Behavior

Runtime bootstrap is handled by `internal/service/migration/migrations.go`.

On startup, TuskBot:

- Ensures `config/` and `prompt/` directories exist.
- Creates these files from embedded defaults **if they are missing**:
  - `config/mcp_config.json`
  - `prompt/SYSTEM.md`
  - `prompt/IDENTITY.md`
  - `prompt/USER.md`
  - `prompt/MEMORY.md`
  - `prompt/SUBAGENT.md`
- Verifies the embedding model exists at `models/<TUSK_EMBEDDING_MODEL>`.
  - If it is missing, startup fails and asks you to run `tusk install`.

## Database (`tuskbot.db`)

`internal/config/app.go` resolves the database path as `<TUSK_RUNTIME_PATH>/tuskbot.db`.

Schema is managed by SQL migrations in `internal/storage/sqlite/migrations/` and currently includes:

- `messages`
- `messages_vec` (vector storage)
- `knowledge`
- `knowledge_vec` (vector storage)
- `task` (scheduled/background jobs)

## Environment File (`.env`)

`.env` is stored in the runtime root and loaded at startup when present.
Configuration updates that call `SetModel(...)` persist values back to this file.
