# Operational Guide

This document outlines the procedures for interacting with and verifying the TuskBot service.

## Access Control
The system implements a strict whitelist based on the `TUSK_TELEGRAM_OWNER_ID`. Inbound messages from unauthorized identifiers are discarded without processing.

## Functional Verification

Execute the following test cases to validate system integrity:

### 1. Native Tool Execution
**Prompt**: `List files in the current directory.`
**Expected Result**: Output from the `list_directory` tool.

### 2. RAG Pipeline Verification
**Prompt**: `Store the fact that my server port is 8080.`
**Follow-up**: `What is my server port?`
**Expected Result**: Retrieval of the stored fact from the vector database.

### 3. MCP Connectivity
**Prompt**: Invoke a tool specific to a configured external MCP server.
**Expected Result**: Successful tool execution and response formatting.
