# TuskBot — A Multi-Agent System

- Stick to the SOLID principles.
- Follow Clean Code/Architecture best practices.
- Don't write spaghetti code.
- Refactor your code when it gets messy.
- Write table-driven tests if asked.
- Use abstractions and interfaces where possible and reasonable.
- Use short but meaningful variable and function names.
- Make thin wrappers for external libraries in `pkg` folder. Treat it as a "framework" layer.
- Don't over-engineer solutions, but keep in mind scalability and maintainability.
- Don't wire dependencies in main.go. Use dependency setup.go file instead.
- Don't add comments where the code is self-explanatory.
- **Configuration Pattern**: Define config structs in `config/` package. Create a `New<Name>Config(ctx)` function for each struct that handles environment variable parsing (using `caarlos0/env` or similar) and validation. Inject these config structs into your services in `setup.go`. Do not use `os.Getenv` directly in service constructors or `setup.go`.

## Architecture & Folder Structure

We follow a Hexagonal / Clean Architecture approach to ensure scalability and testability.

### 1. `cmd/` (Entry Points)
- Contains the main application entry points.
- Will use `cobra` or similar for CLI commands (e.g., `tusk start`, `tusk version`).
- **Responsibility**: Parse flags, initialize configuration, call `setup.go`, and start the application.

### 2. `internal/core/` (Domain)
- Contains pure domain entities and business rules.
- **No external dependencies** (no SQL, no HTTP, no LLM SDKs).
- Examples: `Tool`, `Message`, `Model` structs.

### 3. `internal/storage/` (Driven Adapters - Persistence)
- Implementations of repositories for data storage.
- Examples: `sqlite`, `postgres`, `vector`.

### 4. `internal/providers/` (Driven Adapters - External Services)
- Clients for external APIs and tools.
- **Responsibility**: Adapt external libraries/APIs to our internal interfaces.
- Currently: `llm` (OpenRouter, OpenAI, etc), `mcp` (Tool Managers), `rag` (Embedding management).

### 5. `internal/transport/` (Driving Adapters - Inputs)
- The "Ports" through which the outside world interacts with the application.
- **Responsibility**: Receive input, call the appropriate Service, and return output.
- Currently: only `telegram`. Examples: `cli` (readline), `http` (REST/WebSocket), `cron`.

### 6. `internal/service/` (Application Logic)
- The "Glue" that orchestrates the flow between Transports, Providers, and Storage.
- **Responsibility**: Implement business use cases.
- Currently:
    - `agent`: The core ReAct loop.
    - `command`: Slash-command routing (e.g., `/reset`, `/help`).
    - `installer`: TUI installer.
    - `memory`: Memory management (short-term and long-term).

### Dependency Flow
`Transport` -> `Service` -> `Core` <- `Storage` / `Providers`

(Transports depend on Services. Services depend on Core interfaces. Storage/Providers implement Core interfaces.)