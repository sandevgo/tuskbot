# Binary Deployment

This document specifies the requirements and procedures for deploying TuskBot as a standalone binary.

## Hardware Specifications

| Resource | Minimum Requirement | Recommended |
| :--- | :--- | :--- |
| **CPU** | 1 Core (x86_64/ARM64) | 2+ Cores |
| **RAM** | 1 GB | 2 GB |
| **Storage** | 500 MB | 2 GB+ (for vector growth) |

## Installation Procedure

### 1. Binary Acquisition
Download the architecture-specific archive from the official repository.

```bash
# Example for Linux x86_64
tar -xzvf tusk-linux-amd64.tar.gz
chmod +x tusk
sudo mv tusk /usr/local/bin/
```

### 2. Environment Initialization
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
tusk start
```
