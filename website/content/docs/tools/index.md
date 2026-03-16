# Native Tools Specification

TuskBot provides a suite of built-in "Native Tools" for local environment interaction and web data retrieval. These tools are implemented as internal functions and do not require external MCP server processes.

## Filesystem Interface

The filesystem toolset facilitates CRUD operations and metadata retrieval within the host environment.

### Methods

| Method | Description | Parameters |
| :--- | :--- | :--- |
| `read_file` | Retrieves raw content of a specified file. | `path: string` |
| `write_file` | Persists content to a specified path. Creates parent directories if absent. | `path: string`, `content: string` |
| `edit_file` | Performs an exact string replacement within a file. | `path: string`, `find: string`, `replace: string` |
| `list_directory` | Enumerates entries within a directory including size metadata. | `path: string` |
| `search_files` | Recursively scans for a query string. Excludes `node_modules`, `vendor`, and hidden directories. | `path: string`, `query: string` |
| `get_file_info` | Returns POSIX-compliant metadata (size, mode, modification time). | `path: string` |

## Shell Interface

The shell interface enables arbitrary command execution via the host operating system's command processor (`sh` on Unix-like systems, `cmd` on Windows).

### Methods

| Method | Description | Parameters |
| :--- | :--- | :--- |
| `execute_command` | Executes a shell command and returns `STDOUT` and `STDERR`. | `command: string` |

### Execution Constraints
- **Timeout**: Execution is forcibly terminated after 300 seconds (5 minutes).
- **Output Truncation**: Returns only the final 200 lines of output if the buffer exceeds this limit.
- **Working Directory**: Operations are scoped to the configured `WorkDir`.

## Fetch Interface

The fetch interface provides idempotent HTTP GET capabilities with automated content transformation.

### Methods

| Method | Description | Parameters |
| :--- | :--- | :--- |
| `fetch_url` | Retrieves remote resources and converts HTML to Markdown-formatted text. | `url: string` |

### Technical Specifications
- **User-Agent**: Identifies as `TuskBot-Agent/0.1`.
- **Payload Limit**: Maximum response size is 1MB.
- **Resiliency**: Implements an exponential backoff retry strategy for 5xx status codes and network-level errors.
- **Timeout**: Request lifecycle is limited to 15 seconds.

## Schedule Interface

The schedule interface lets the agent create, list, and cancel background tasks.

### Methods

| Method | Description | Parameters |
| :--- | :--- | :--- |
| `schedule_once` | Schedules a task to run once at a specific RFC3339 time. | `name: string`, `prompt: string`, `at: string` |
| `schedule_cron` | Schedules a recurring task using a cron expression. | `name: string`, `prompt: string`, `at: string` |
| `schedule_list` | Lists all currently scheduled tasks. | None |
| `schedule_cancel` | Cancels a scheduled task by ID. | `task_id: string` |

### Scheduling Notes
- `schedule_once` expects `at` in RFC3339 format (for example, `2026-03-16T14:30:00Z`).
- `schedule_cron` expects a 5-field cron expression.
- Task names must use alphanumeric characters and hyphens.

## Security & Governance

- **Path Scoping**: Filesystem operations are restricted to paths relative to the `TUSK_RUNTIME_PATH` environment variable.
- **Resource Protection**: Strict timeouts are enforced at the provider level to prevent resource exhaustion.
- **Binary Safety**: `search_files` implements a null-byte heuristic to skip binary file processing.
