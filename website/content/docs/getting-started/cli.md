# CLI Reference

TuskBot is managed via the `tusk` binary.

## Global Flag

| Flag | Shorthand | Description |
| :--- | :--- | :--- |
| `--debug` | `-d` | Enable debug logging. |

## Core Commands

### `tusk install`

Run interactive setup to create `.env` and initialize runtime data.

```bash
tusk install
```

### `tusk run`

Run TuskBot in the foreground (recommended for Docker and manual local runs).

```bash
tusk run
```

### `tusk start`

Alias for `tusk service start`.

```bash
tusk start
```

### `tusk stop`

Alias for `tusk service stop`.

```bash
tusk stop
```

### `tusk version`

Print the current TuskBot version.

```bash
tusk version
```

## Service Management

### `tusk service install`

Install TuskBot as a background service.

```bash
tusk service install
```

### `tusk service uninstall`

Uninstall the configured background service.

```bash
tusk service uninstall
```

### `tusk service start`

Start the installed service.

```bash
tusk service start
```

### `tusk service stop`

Stop the running service.

```bash
tusk service stop
```

### `tusk service restart`

Restart the service.

```bash
tusk service restart
```

### `tusk service status`

Show service status (`running`, `stopped`, or `unknown`).

```bash
tusk service status
```
