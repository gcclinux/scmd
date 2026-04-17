package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"pgregory.net/rapid"
)

// setupTestConfigForPBT creates a temp config dir with a config.json for use in
// property-based tests. It uses the standard *testing.T for directory creation
// and returns the dir path and a cleanup function.
func setupTestConfigForPBT(t *testing.T, cfg ConfigData) (configDir string, cleanup func()) {
	t.Helper()

	dir := t.TempDir()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	envVars := []string{"DB_TYPE", "MCP_SERVER", "AGENT", "DB_HOST", "DB_PORT",
		"DB_USER", "DB_PASS", "DB_NAME", "TB_NAME", "GEMINIAPI", "GEMINIMODEL",
		"GEMINI_EMBEDDING_MODEL", "OLLAMA", "MODEL", "EMBEDDING_MODEL", "EMBEDDING_DIM"}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	return dir, func() {
		for _, v := range envVars {
			os.Unsetenv(v)
		}
	}
}

// Feature: mcp-client-integration, Property 1: Config loading sets MCP environment variables correctly
// **Validates: Requirements 1.1, 1.2, 1.4**
func TestProperty_ConfigLoadingSetsMCPEnvVars(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random db_type: either "mcp" or some other value.
		dbType := rapid.SampledFrom([]string{
			"mcp", "sqlite", "postgresql", "",
		}).Draw(rt, "db_type")

		// Generate a random mcp_server relative path (non-empty filename).
		mcpServerFilename := rapid.StringMatching(`[a-zA-Z0-9_]{1,20}\.json`).Draw(rt, "mcp_server")

		// Set up a temp config directory with the generated values.
		dir, cleanup := setupTestConfigForPBT(t, ConfigData{
			DBType:    dbType,
			MCPServer: mcpServerFilename,
		})
		defer cleanup()

		// Load config from the temp directory.
		loadConfigFromDir(dir)

		if dbType == "mcp" {
			// Requirement 1.1: DB_TYPE must be "mcp"
			gotDBType := os.Getenv("DB_TYPE")
			if gotDBType != "mcp" {
				rt.Fatalf("DB_TYPE = %q, want %q", gotDBType, "mcp")
			}

			// Requirement 1.4: MCP_SERVER must be the resolved absolute path
			gotMCPServer := os.Getenv("MCP_SERVER")
			if !filepath.IsAbs(gotMCPServer) {
				rt.Fatalf("MCP_SERVER should be absolute, got %q", gotMCPServer)
			}

			// Requirement 1.2: The resolved path should incorporate the mcp_server value.
			// Since mcpServerFilename is relative, it should be joined with the config dir.
			expectedPath := filepath.Join(dir, mcpServerFilename)
			if gotMCPServer != expectedPath {
				rt.Fatalf("MCP_SERVER = %q, want %q", gotMCPServer, expectedPath)
			}
		} else if dbType != "" {
			// For non-mcp modes, DB_TYPE should be set to the provided value.
			gotDBType := os.Getenv("DB_TYPE")
			if gotDBType != dbType {
				rt.Fatalf("DB_TYPE = %q, want %q", gotDBType, dbType)
			}

			// MCP_SERVER should be the raw value (not resolved), since db_type is not "mcp".
			gotMCPServer := os.Getenv("MCP_SERVER")
			if gotMCPServer != mcpServerFilename {
				rt.Fatalf("MCP_SERVER = %q, want raw value %q", gotMCPServer, mcpServerFilename)
			}
		}
	})
}
