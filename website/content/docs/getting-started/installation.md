# Binary Installation

Install TuskBot as a native binary on Linux or macOS.

## Quick Install

```bash
curl -fsSL https://tuskbot.ai/install.sh | sh
```

The installer performs the full bootstrap flow:

1. Downloads the latest stable release binary.
2. Installs `tusk` into a user-local bin directory.
3. Runs `tusk install` (interactive setup).
4. Runs `tusk service install`.
5. Runs `tusk service start`.
6. Verifies service status is running.

Supported release artifacts:

- Linux amd64
- Linux arm64
- macOS arm64

## Verify Installation

```bash
tusk service status
```

Expected output includes `Service status: running`.

If you prefer foreground execution:

```bash
tusk run
```

## Manual Install (Fallback)

1. Download the correct archive from [GitHub Releases](https://github.com/sandevgo/tuskbot/releases).
2. Extract and install the binary.

```bash
# Example for Linux amd64
tar -xzvf tusk-linux-amd64.tar.gz
chmod +x bin/tusk-linux-amd64
mkdir -p ~/.local/bin
mv bin/tusk-linux-amd64 ~/.local/bin/tusk
```

Then run:

```bash
tusk install
tusk service install
tusk service start
tusk service status
```
