# First Steps

After installation, use these checks in Telegram to confirm your setup.

## 1) Verify Command Routing

Send:

```text
/help
```

You should get a list of available slash commands.

## 2) Verify Model Configuration

Send:

```text
/model
```

You should see current provider and model.

## 3) Verify MCP Tool Registration

Send:

```text
/mcp
```

You should see connected MCP tools (or a clear message if none are configured).

## 4) Verify General Agent Execution

Send a basic prompt, for example:

```text
List files in the current directory.
```

You should get a normal tool-backed response.

## 5) Verify Session Metrics

Send:

```text
/stats
```

You should see session message count and context token usage.

## 6) Verify Scheduled Tasks View

Send:

```text
/tasks
```

You should see active scheduled tasks (or an empty-state message).
