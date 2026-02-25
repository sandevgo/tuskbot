import { defineConfig } from 'vitepress'

export default defineConfig({
    srcDir: "./content",

    title: "TuskBot",
    description: "Autonomous AI Agent in Your Messenger",

    themeConfig: {
        nav: [
            { text: 'Documentation', link: '/docs/getting-started/' }, 
        ],

        sidebar: {
            '/docs/getting-started/': [
                {
                    text: 'Getting Started',
                    items: [
                        { text: 'Overview', link: '/docs/getting-started/' },
                        { text: 'Installation', link: '/docs/getting-started/installation' },
                        { text: 'Docker Setup', link: '/docs/getting-started/docker' },
                        { text: 'Configuration', link: '/docs/getting-started/configuration' },
                        { text: 'First Steps', link: '/docs/getting-started/first-steps' }
                    ]
                }
            ],
            '/docs/architecture/': [
                {
                    text: 'Architecture',
                    items: [
                        { text: 'Overview', link: '/docs/architecture/' },
                        { text: 'MCP (Model Context Protocol)', link: '/docs/architecture/mcp' },
                        { text: 'RAG & Memory', link: '/docs/architecture/rag-memory' },
                        { text: 'Providers', link: '/docs/architecture/providers' }
                    ]
                }
            ],
            '/docs/configuration/': [
                {
                    text: 'Configuration',
                    items: [
                        { text: 'Overview', link: '/docs/configuration/' },
                        { text: 'LLM Providers', link: '/docs/configuration/providers' },
                        { text: 'Telegram Setup', link: '/docs/configuration/telegram' },
                        { text: 'Memory & Embeddings', link: '/docs/configuration/memory' },
                        { text: 'MCP Servers', link: '/docs/configuration/mcp-servers' }
                    ]
                }
            ],
            '/docs/cli-reference/': [
                {
                    text: 'CLI Reference',
                    items: [
                        { text: 'Overview', link: '/docs/cli-reference/' },
                        { text: 'tusk start', link: '/docs/cli-reference/start' },
                        { text: 'tusk install', link: '/docs/cli-reference/install' },
                        { text: 'Environment Variables', link: '/docs/cli-reference/environment' }
                    ]
                }
            ],
            '/docs/tools/': [
                {
                    text: 'Tools & MCP',
                    items: [
                        { text: 'Overview', link: '/docs/tools/' },
                        { text: 'Filesystem', link: '/docs/tools/filesystem' },
                        { text: 'Shell', link: '/docs/tools/shell' },
                        { text: 'Fetch', link: '/docs/tools/fetch' },
                        { text: 'Custom MCP Servers', link: '/docs/tools/custom-mcp' }
                    ]
                }
            ],
            '/docs/development/': [
                {
                    text: 'Development',
                    items: [
                        { text: 'Building from Source', link: '/docs/development/' },
                        { text: 'Project Structure', link: '/docs/development/architecture' },
                        { text: 'Adding Providers', link: '/docs/development/adding-providers' },
                        { text: 'Testing', link: '/docs/development/testing' }
                    ]
                }
            ],
            '/docs/troubleshooting/': [
                {
                    text: 'Troubleshooting',
                    items: [
                        { text: 'FAQ', link: '/docs/troubleshooting/' },
                        { text: 'Common Issues', link: '/docs/troubleshooting/common-issues' },
                        { text: 'Migration Guides', link: '/docs/troubleshooting/migration' }
                    ]
                }
            ]
        },

        socialLinks: [
            { icon: 'github', link: 'https://github.com/sandevgo/tuskbot' }
        ],

        footer: {
            message: 'Released under the MIT License.',
            copyright: 'Copyright © 2026 TuskBot Contributors'
        }
    }
})