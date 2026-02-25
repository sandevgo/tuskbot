import { defineConfig } from 'vitepress'

export default defineConfig({
  srcDir: "../docs",
  
  title: "TuskBot",
  description: "Autonomous AI Agent for Telegram",
  
  themeConfig: {
    nav: [
      { text: 'Guide', link: '/getting-started/' },
      { text: 'Architecture', link: '/architecture/' },
      { text: 'Configuration', link: '/configuration/' },
      { text: 'CLI', link: '/cli-reference/' },
      { text: 'Tools', link: '/tools/' },
      { text: 'Development', link: '/development/' },
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
      ],
      '/troubleshooting/': [
        {
          text: 'Troubleshooting',
          items: [
            { text: 'FAQ', link: '/troubleshooting/' },
            { text: 'Common Issues', link: '/troubleshooting/common-issues' },
            { text: 'Migration Guides', link: '/troubleshooting/migration' }
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
