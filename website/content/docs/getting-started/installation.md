# Binary Deployment

This document specifies the requirements and procedures for deploying TuskBot as a standalone binary.

## Hardware Specifications

| Resource | Minimum Requirement | Recommended |
| :--- | :--- | :--- |
| **CPU** | 1 Core (x86_64/ARM64) | 2+ Cores |
| **RAM** | 1 GB | 2 GB |
| **Storage** | 500 MB | 2 GB+ (for vector growth) |

## Installation Procedure

### 1. Quick Install (Linux/macOS)
Run the installer script:

```bash
curl -fsSL https://raw.githubusercontent.com/sandevgo/tuskbot/main/scripts/install.sh | sh
```

Supported release artifacts:
- Linux amd64
- Linux arm64
- macOS arm64

The installer downloads the latest stable release binary, installs it to a user-local bin directory, then runs:
1. `tusk install` (interactive setup)
2. `tusk service install`
3. `tusk service start`

### 2. Manual Binary Acquisition (fallback)
Download the architecture-specific archive from the official repository and install manually.

```bash
# Example for Linux amd64
tar -xzvf tusk-linux-amd64.tar.gz
chmod +x bin/tusk-linux-amd64
mkdir -p ~/.local/bin
mv bin/tusk-linux-amd64 ~/.local/bin/tusk
```

### 3. Environment Initialization (if manually installed)
Execute the interactive configuration utility to generate the required directory structure and `.env` specification.

```bash
tusk install
```

The utility automates the following:
- Provider authentication configuration.
- GGUF embedding model acquisition.
- Telegram Bot API credential validation.

## Service Verification
Initiate the process and monitor STDOUT for the initialization sequence.

```bash
tusk service status
```
