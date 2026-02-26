---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "TuskBot"
  text: "Autonomous AI Agent"
  tagline: Privacy-first AI assistant with local RAG, MCP tools, and multi-provider LLM support
  actions:
    - theme: brand
      text: Get Started
      link: /docs/getting-started/
    - theme: alt
      text: View on GitHub
      link: https://github.com/sandevgo/tuskbot

features:
  - title: 🔌 MCP-First Architecture
    details: Extensible via Model Context Protocol. Connect any MCP server without modifying core code.
  - title: 🧠 Private RAG Memory
    details: Local embeddings with llama.cpp and SQLite-vec. Your data never leaves your hardware.
  - title: 🛠️ System Tools
    details: Built-in filesystem, shell execution, and fetch capabilities. Autonomous task execution.
  - title: 🤖 Multi-Provider LLM
    details: Support for OpenAI, Anthropic, OpenRouter, Ollama, and custom OpenAI-compatible APIs.
---
