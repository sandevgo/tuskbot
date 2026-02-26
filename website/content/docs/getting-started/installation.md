# Installation

How to install TuskBot on your system.

## System Requirements

- **Go**: Version 1.24 or higher (if building from source).
- **CGO**: Required for building from source (not needed for pre-compiled binaries).
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

The `tusk install` command launches an interactive TUI wizard that automatically configures your environment. It handles:
1. Selecting an AI Provider.
2. Configuring API Keys.
3. Downloading the local embedding model.
4. Setting up Telegram credentials.

This process creates a `.env` file in your runtime directory, so you don't have to set variables manually.

## Verification

Verify the installation by running:
```bash
tusk start
```
The bot should log "starting telegram bot" if configured correctly.
