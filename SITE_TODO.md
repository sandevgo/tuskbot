# Documentation Structure TODO

This document outlines the proposed file structure and content organization for the TuskBot documentation site.

## Directory Structure

```
docs/
├── index.md                    # Landing page (already exists)
├── getting-started/
│   ├── index.md               # Installation overview
│   ├── installation.md        # Binary install, quick install script
│   ├── docker.md              # Docker/docker-compose setup
│   ├── configuration.md       # Basic env vars setup
│   └── first-steps.md         # Initial setup, /model command, testing
├── architecture/
│   ├── index.md               # Overview of Clean/Hexagonal architecture
│   ├── mcp.md                 # Model Context Protocol deep dive
│   ├── rag-memory.md          # SQLite-vec, embeddings, context window
│   └── providers.md           # LLM provider abstraction
├── configuration/
│   ├── index.md               # Configuration overview
│   ├── providers.md           # OpenAI, Anthropic, Ollama, OpenRouter
│   ├── telegram.md            # Bot token, owner ID setup
│   ├── memory.md              # Embedding models, context window size
│   └── mcp-servers.md         # MCP config file format
├── cli-reference/
│   ├── index.md               # CLI overview
│   ├── start.md               # `tusk start` command
│   ├── install.md             # `tusk install` command
│   └── environment.md         # All env vars reference
├── tools/
│   ├── index.md               # Built-in tools overview
│   ├── filesystem.md          # File operations
│   ├── shell.md               # Command execution
│   ├── fetch.md               # HTTP requests
│   └── custom-mcp.md          # Writing/connecting MCP servers
├── development/
│   ├── index.md               # Building from source
│   ├── architecture.md        # Code structure (internal/, pkg/, cmd/)
│   ├── adding-providers.md    # How to add new LLM providers
│   └── testing.md             # Test setup, mocks
└── troubleshooting/
    ├── index.md               # FAQ
    ├── common-issues.md         # Debug mode, logs
    └── migration.md             # Version upgrade guides
```

## VitePress Config Updates

Update `website/.vitepress/config.mts` to include the navigation structure:

```typescript
import { defineConfig } from 'vitepress'

export default defineConfig({
  srcDir: "../docs",
  
  title: "TuskBot",
  description: "Autonomous AI Agent for Telegram",
  
  themeConfig: {
    nav: [
      { text: 'Guide', link: '/getting-started/' },
      { text: 'Architecture', link: '/architecture/' },
      { text: 'CLI', link: '/cli-reference/' },
      { text: 'Tools', link: '/tools/' },
      { text: 'GitHub', link: 'https://github.com/sandevgo/tuskbot' }
    ],

    sidebar: {
      '/getting-started/': [
        {
          text: 'Getting Started',
          items: [
            { text: 'Overview', link: '/getting-started/' },
            { text: 'Installation', link: '/getting-started/installation' },
            { text: 'Docker Setup', link: '/getting-started/docker' },
            { text: 'Configuration', link: '/getting-started/configuration' },
            { text: 'First Steps', link: '/getting-started/first-steps' }
          ]
        }
      ],
      '/architecture/': [
        {
          text: 'Architecture',
          items: [
            { text: 'Overview', link: '/architecture/' },
            { text: 'MCP (Model Context Protocol)', link: '/architecture/mcp' },
            { text: 'RAG & Memory', link: '/architecture/rag-memory' },
            { text: 'Providers', link: '/architecture/providers' }
          ]
        }
      ],
      '/configuration/': [
        {
          text: 'Configuration',
          items: [
            { text: 'Overview', link: '/configuration/' },
            { text: 'LLM Providers', link: '/configuration/providers' },
            { text: 'Telegram Setup', link: '/configuration/telegram' },
            { text: 'Memory & Embeddings', link: '/configuration/memory' },
            { text: 'MCP Servers', link: '/configuration/mcp-servers' }
          ]
        }
      ],
      '/cli-reference/': [
        {
          text: 'CLI Reference',
          items: [
            { text: 'Overview', link: '/cli-reference/' },
            { text: 'tusk start', link: '/cli-reference/start' },
            { text: 'tusk install', link: '/cli-reference/install' },
            { text: 'Environment Variables', link: '/cli-reference/environment' }
          ]
        }
      ],
      '/tools/': [
        {
          text: 'Tools & MCP',
          items: [
            { text: 'Overview', link: '/tools/' },
            { text: 'Filesystem', link: '/tools/filesystem' },
            { text: 'Shell', link: '/tools/shell' },
            { text: 'Fetch', link: '/tools/fetch' },
            { text: 'Custom MCP Servers', link: '/tools/custom-mcp' }
          ]
        }
      ],
      '/development/': [
        {
          text: 'Development',
          items: [
            { text: 'Building from Source', link: '/development/' },
            { text: 'Project Structure', link: '/development/architecture' },
            { text: 'Adding Providers', link: '/development/adding-providers' },
            { text: 'Testing', link: '/development/testing' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/sandevgo/tuskbot' }
    ],
    
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024-present TuskBot Contributors'
    }
  }
})
```

## Content Priorities

1. **Getting Started**: Focus on the `tusk install` TUI flow and Docker setup since the project has a sophisticated installer
2. **MCP Documentation**: This is a core differentiator - document how the `internal/providers/mcp/` system works and how users can connect external MCP servers
3. **Provider Matrix**: Document which providers support tool calling vs. just chat (based on `internal/providers/llm/` implementations)
4. **Memory/RAG**: Explain the local embedding flow (E5Base → llama.cpp → SQLite-vec) since this is a privacy-focused feature
5. **Slash Commands**: Document `/model`, `/mcp`, `/reset` commands from `internal/service/command/`

## Existing Files to Migrate

- `docs/index.md` - Keep as landing page (already configured)
- `docs/markdown-examples.md` - Move to `getting-started/` or delete (was template content)
- `docs/api-examples.md` - Move to `development/` or delete (was template content)
- `README.md` - Keep as repo root README, but move detailed sections to docs site

## Notes

- The structure scales from user-facing docs (installation, configuration) to developer docs (architecture, adding providers)
- Highlights the MCP-first design and local RAG capabilities
- Follows the Clean Architecture structure outlined in AGENTS.md
