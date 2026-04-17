package mcpclient

import (
	"encoding/json"
	"fmt"
	"os"
)

// MCPServerConfig represents the mcp_server.json file structure.
type MCPServerConfig struct {
	MCPServers map[string]MCPServerEntry `json:"mcpServers"`
}

// MCPServerEntry represents a single MCP server entry.
type MCPServerEntry struct {
	URL string `json:"url"`
}

// ParseConfig reads and validates the mcp_server.json file.
// It returns the first server name and its SSE endpoint URL.
func ParseConfig(path string) (serverName string, endpoint string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("MCP server config not found: %s", path)
		}
		return "", "", fmt.Errorf("failed to read MCP server config: %s: %v", path, err)
	}

	var cfg MCPServerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", "", fmt.Errorf("failed to parse MCP server config: %v", err)
	}

	if len(cfg.MCPServers) == 0 {
		return "", "", fmt.Errorf("no valid MCP server entry found in %s", path)
	}

	// Use the first entry in the map.
	for name, entry := range cfg.MCPServers {
		if entry.URL == "" {
			return "", "", fmt.Errorf("no valid MCP server entry found in %s", path)
		}
		return name, entry.URL, nil
	}

	// Unreachable, but satisfies the compiler.
	return "", "", fmt.Errorf("no valid MCP server entry found in %s", path)
}
