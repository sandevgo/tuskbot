# Installation

How to install TuskBot on your system.

## System Requirements

TuskBot uses lightweight embedding models via llama.cpp and is designed to be resource-efficient. However, it's recommended to have at least the following:

- **CPU**: 2 cores recommended.
- **Memory**: 1GB RAM minimum.
- **Disk**: ~300MB+ for runtime and data storage.

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
