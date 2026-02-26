# Installation

How to install TuskBot on your system.

## System Requirements

- **Go**: Version 1.21 or higher (if building from source).
- **CGO**: Required for `sqlite-vec` and `llama.cpp` bindings.
- **Storage**: Space for the runtime directory (default `~/.tuskbot`).

## Binary Installation

Download the pre-compiled binary for your platform from the Releases page.

**Quick Install (Linux/macOS):**

```bash
tar -xzvf tusk-*.tar.gz
chmod +x tusk-*
sudo mv tusk-* /usr/local/bin/tusk
tusk install
```

## Quick Install Script

The `tusk install` command launches an interactive TUI wizard that guides you through:
1. Selecting an AI Provider.
2. Configuring API Keys.
3. Downloading the local embedding model.
4. Setting up Telegram credentials.

## Verification

Verify the installation by running:
```bash
tusk start
```
The bot should log "starting telegram bot" if configured correctly.
