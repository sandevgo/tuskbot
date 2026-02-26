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
            '/docs/': [
                {
                    text: 'Getting Started',
                    collapsed: false,
                    items: [
                        { text: 'Overview', link: '/docs/getting-started/' },
                        { text: 'Installation', link: '/docs/getting-started/installation' },
                        { text: 'Docker Setup', link: '/docs/getting-started/docker' },
                        { text: 'Configuration', link: '/docs/getting-started/configuration' },
                        { text: 'First Steps', link: '/docs/getting-started/first-steps' }
                    ]
                },
                {
                    text: 'Architecture',
                    collapsed: true,
                    items: [
                        { text: 'Overview', link: '/docs/architecture/' },
                        { text: 'MCP (Model Context Protocol)', link: '/docs/architecture/mcp' },
                        { text: 'RAG & Memory', link: '/docs/architecture/rag-memory' },
                        { text: 'Providers', link: '/docs/architecture/providers' }
                    ]
                },
                {
                    text: 'Configuration',
                    collapsed: true,
                    items: [
                        { text: 'Overview', link: '/docs/configuration/' },
                        { text: 'LLM Providers', link: '/docs/configuration/providers' },
                        { text: 'Telegram Setup', link: '/docs/configuration/telegram' },
                        { text: 'Memory & Embeddings', link: '/docs/configuration/memory' },
                        { text: 'MCP Servers', link: '/docs/configuration/mcp-servers' }
                    ]
                },
                {
                    text: 'Tools & MCP',
                    collapsed: true,
                    items: [
                        { text: 'Overview', link: '/docs/tools/' },
                        { text: 'Filesystem', link: '/docs/tools/filesystem' },
                        { text: 'Shell', link: '/docs/tools/shell' },
                        { text: 'Fetch', link: '/docs/tools/fetch' },
                        { text: 'Custom MCP Servers', link: '/docs/tools/custom-mcp' }
                    ]
                },
                {
                    text: 'CLI Reference',
                    collapsed: true,
                    items: [
                        { text: 'Overview', link: '/docs/cli-reference/' },
                        { text: 'tusk start', link: '/docs/cli-reference/start' },
                        { text: 'tusk install', link: '/docs/cli-reference/install' },
                        { text: 'Environment Variables', link: '/docs/cli-reference/environment' }
                    ]
                },
                {
                    text: 'Development',
                    collapsed: true,
                    items: [
                        { text: 'Building from Source', link: '/docs/development/' },
                        { text: 'Project Structure', link: '/docs/development/architecture' },
                        { text: 'Adding Providers', link: '/docs/development/adding-providers' },
                        { text: 'Testing', link: '/docs/development/testing' }
                    ]
                },
                {
                    text: 'Troubleshooting',
                    collapsed: true,
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
