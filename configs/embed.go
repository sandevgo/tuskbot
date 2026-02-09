package configs

import "embed"

//go:embed IDENTITY.md MEMORY.md SYSTEM.md USER.md mcp_config.json
var FS embed.FS
