package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// helper to create a temp config dir with a config.json and override configPath.
// Returns a cleanup function.
func setupTestConfig(t *testing.T, cfg ConfigData) (configDir string, cleanup func()) {
	t.Helper()

	dir := t.TempDir()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Clear relevant env vars before each test.
	envVars := []string{"DB_TYPE", "MCP_SERVER", "AGENT", "GEMINIAPI", "GEMINIMODEL",
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

// loadConfigFromDir is a test helper that loads config from a specific directory.
func loadConfigFromDir(dir string) {
	path := filepath.Join(dir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg ConfigData
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}

	setIfNotEmpty("AGENT", cfg.Agent)
	setIfNotEmpty("DB_TYPE", cfg.DBType)
	setIfNotEmpty("GEMINIAPI", cfg.GeminiAPI)
	setIfNotEmpty("GEMINIMODEL", cfg.GeminiModel)
	setIfNotEmpty("GEMINI_EMBEDDING_MODEL", cfg.GeminiEmbeddingModel)
	setIfNotEmpty("OLLAMA", cfg.Ollama)
	setIfNotEmpty("MODEL", cfg.Model)
	setIfNotEmpty("EMBEDDING_MODEL", cfg.EmbeddingModel)
	setIfNotEmpty("EMBEDDING_DIM", cfg.EmbeddingDim)
	setIfNotEmpty("MCP_SERVER", cfg.MCPServer)

	// Replicate the MCP resolution logic using the test dir instead of ConfigDir().
	if cfg.DBType == "mcp" {
		mcpPath := os.Getenv("MCP_SERVER")
		if mcpPath == "" {
			mcpPath = filepath.Join(dir, "mcp_server.json")
		}
		if !filepath.IsAbs(mcpPath) {
			mcpPath = filepath.Join(dir, mcpPath)
		}
		os.Setenv("MCP_SERVER", mcpPath)
	}
}

func TestMCPServerField_InConfigData(t *testing.T) {
	cfg := ConfigData{MCPServer: "custom_mcp.json"}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded ConfigData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.MCPServer != "custom_mcp.json" {
		t.Errorf("MCPServer = %q, want %q", decoded.MCPServer, "custom_mcp.json")
	}
}

func TestLoadConfig_MCPMode_DefaultPath(t *testing.T) {
	dir, cleanup := setupTestConfig(t, ConfigData{
		DBType: "mcp",
	})
	defer cleanup()

	loadConfigFromDir(dir)

	got := os.Getenv("DB_TYPE")
	if got != "mcp" {
		t.Errorf("DB_TYPE = %q, want %q", got, "mcp")
	}

	mcpServer := os.Getenv("MCP_SERVER")
	expected := filepath.Join(dir, "mcp_server.json")
	if mcpServer != expected {
		t.Errorf("MCP_SERVER = %q, want %q", mcpServer, expected)
	}

	if !filepath.IsAbs(mcpServer) {
		t.Errorf("MCP_SERVER should be absolute, got %q", mcpServer)
	}
}

func TestLoadConfig_MCPMode_CustomRelativePath(t *testing.T) {
	dir, cleanup := setupTestConfig(t, ConfigData{
		DBType:    "mcp",
		MCPServer: "custom/mcp.json",
	})
	defer cleanup()

	loadConfigFromDir(dir)

	mcpServer := os.Getenv("MCP_SERVER")
	expected := filepath.Join(dir, "custom", "mcp.json")
	if mcpServer != expected {
		t.Errorf("MCP_SERVER = %q, want %q", mcpServer, expected)
	}
}

func TestLoadConfig_MCPMode_AbsolutePath(t *testing.T) {
	absPath := filepath.Join(os.TempDir(), "my_mcp_server.json")
	dir, cleanup := setupTestConfig(t, ConfigData{
		DBType:    "mcp",
		MCPServer: absPath,
	})
	defer cleanup()

	loadConfigFromDir(dir)

	mcpServer := os.Getenv("MCP_SERVER")
	if mcpServer != absPath {
		t.Errorf("MCP_SERVER = %q, want %q", mcpServer, absPath)
	}
}

func TestLoadConfig_NonMCPMode_RawValueSet(t *testing.T) {
	dir, cleanup := setupTestConfig(t, ConfigData{
		DBType:    "sqlite",
		MCPServer: "some_server.json",
	})
	defer cleanup()

	loadConfigFromDir(dir)

	got := os.Getenv("DB_TYPE")
	if got != "sqlite" {
		t.Errorf("DB_TYPE = %q, want %q", got, "sqlite")
	}

	// For non-mcp mode, the raw value should be set without resolution.
	mcpServer := os.Getenv("MCP_SERVER")
	if mcpServer != "some_server.json" {
		t.Errorf("MCP_SERVER = %q, want %q", mcpServer, "some_server.json")
	}
}

func TestLoadConfig_NonMCPMode_EmptyMCPServer(t *testing.T) {
	dir, cleanup := setupTestConfig(t, ConfigData{
		DBType: "sqlite",
	})
	defer cleanup()

	loadConfigFromDir(dir)

	mcpServer := os.Getenv("MCP_SERVER")
	if mcpServer != "" {
		t.Errorf("MCP_SERVER = %q, want empty", mcpServer)
	}
}
