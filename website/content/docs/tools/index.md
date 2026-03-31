# Built-in Tools

TuskBot includes native tools that are available without adding external MCP servers.

## Available Native Tools

| Tool | Purpose |
| :--- | :--- |
| `read_file` | Read file contents |
| `write_file` | Write file contents (creates parent dirs) |
| `edit_file` | Replace exact string occurrences in a file |
| `list_directory` | List directory entries with size info |
| `search_files` | Recursive text search in files |
| `get_file_info` | Show file metadata |
| `execute_command` | Run a shell command |
| `fetch_url` | Fetch URL content and convert HTML to Markdown |
| `schedule_once` | Schedule one-time background task |
| `schedule_cron` | Schedule recurring background task |
| `schedule_list` | List scheduled tasks |
| `schedule_cancel` | Cancel a scheduled task |

## Filesystem Tool Details

### `read_file`

- Input: `path`
- Returns file content as text.

### `write_file`

- Input: `path`, `content`
- Creates parent directories when needed.

### `edit_file`

- Input: `path`, `find`, `replace`
- Replaces all exact matches of `find`.
- Returns an error if `find` is not present.

### `list_directory`

- Input: `path`
- Returns each entry with type marker (`[DIR]` or `[FILE]`) and byte size.

### `search_files`

- Input: `path`, `query`
- Recursive search with practical limits:
  - skips hidden directories
  - skips `vendor` and `node_modules`
  - skips likely binary files (null-byte heuristic)
  - stops after 100 matches

### `get_file_info`

- Input: `path`
- Returns path, size, mode, directory flag, and modification time.

## Shell Tool (`execute_command`)

- Input: `command`
- Executes via:
  - `sh -c` on Unix-like systems
  - `cmd /C` on Windows
- Constraints:
  - timeout: 5 minutes
  - output truncation: last 200 lines
- Returns both `STDOUT` and `STDERR` in one response payload.

## Fetch Tool (`fetch_url`)

- Input: `url`
- Behavior:
  - HTTP GET with User-Agent `TuskBot-Agent/0.1`
  - request timeout: 15s
  - response size limit: 1MB
  - converts HTML to Markdown-like text
  - retries failed requests via backoff policy

## Scheduling Tools

### `schedule_once`

- Inputs: `name`, `prompt`, `at`
- `at` must be RFC3339 timestamp.
- `name` must be slug-like (`[a-zA-Z0-9-]+`).

### `schedule_cron`

- Inputs: `name`, `prompt`, `at`
- `at` must be a 5-field cron expression.

### `schedule_list`

- Lists active scheduled tasks.

### `schedule_cancel`

- Input: `task_id`
- Cancels and removes an existing task.

## Notes

- Native tools are not sandboxed by default to `TUSK_RUNTIME_PATH`.
- For stricter isolation, run TuskBot in a containerized environment with host-level policy controls.
