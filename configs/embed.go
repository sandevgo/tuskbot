package configs

import "embed"

//go:embed IDENTITY.md MEMORY.md SYSTEM.md SUBAGENT.md USER.md mcp_config.json
var FS embed.FS
