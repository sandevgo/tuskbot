# Tools & MCP

TuskBot comes with a set of built-in "Native Tools" that allow it to interact with your local environment and the web. It also supports the Model Context Protocol (MCP) for connecting to external tool servers.

## Native Tools

### Filesystem
The filesystem tool allows the agent to manage files within its workspace.
- **read_file**: Read the contents of a file.
- **write_file**: Create or overwrite a file with new content.
- **edit_file**: Replace specific strings within a file.
- **list_directory**: View files and folders in a path.
- **search_files**: Recursively search for strings within files.
- **get_file_info**: Retrieve metadata like size and modification time.

### Shell
The shell tool provides command execution capabilities.
- **execute_command**: Runs a shell command (e.g., `ls`, `grep`, `go build`) and returns the output.
- **Safety**: Commands are executed with a timeout and output is truncated if it exceeds 200 lines.

### Fetch
The fetch tool allows the agent to retrieve information from the internet.
- **fetch_url**: Performs an HTTP GET request and converts HTML content to readable text.
- **Reliability**: Includes automatic retries for transient network errors.

## Tool Security

- **Path Restrictions**: Filesystem operations are generally relative to the `TUSK_RUNTIME_PATH`.
- **Timeouts**: All tool executions have strict timeouts to prevent the agent from hanging.
