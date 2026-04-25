package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigData holds all configuration fields from config.json.
type ConfigData struct {
	Agent                string `json:"agent"`
	DBType               string `json:"db_type"`
	GeminiAPI            string `json:"gemini_api"`
	GeminiModel          string `json:"gemini_model"`
	GeminiEmbeddingModel string `json:"gemini_embedding_model"`
	Ollama               string `json:"ollama"`
	Model                string `json:"model"`
	EmbeddingModel       string `json:"embedding_model"`
	EmbeddingDim         string `json:"embedding_dim"`
	MCPServer            string `json:"mcp_server"`
}

// configPath returns the path to $HOME/.scmd/config.json.
func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".scmd", "config.json")
}

// LoadConfig reads $HOME/.scmd/config.json and sets environment variables
// so the rest of the application can continue using os.Getenv as before.
func LoadConfig() {
	// Static SQLite configuration
	setIfNotEmpty("DB_NAME", "scmd")
	setIfNotEmpty("TB_NAME", "data")

	path := configPath()
	if path == "" {
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not read config %s: %v\n", path, err)
		return
	}

	var cfg ConfigData
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not parse config %s: %v\n", path, err)
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

	// When db_type is "mcp", resolve the MCP server config path to an absolute path.
	// This overrides the raw value set above with the fully resolved path.
	if strings.EqualFold(cfg.DBType, "mcp") {
		mcpPath := os.Getenv("MCP_SERVER")
		if mcpPath == "" {
			mcpPath = filepath.Join(ConfigDir(), "mcp_server.json")
		}
		if !filepath.IsAbs(mcpPath) {
			mcpPath = filepath.Join(ConfigDir(), mcpPath)
		}
		os.Setenv("MCP_SERVER", mcpPath)
	}
}

// setIfNotEmpty sets an environment variable only if the value is non-empty
// and the variable is not already set (env vars take precedence).
func setIfNotEmpty(key, value string) {
	if value != "" && os.Getenv(key) == "" {
		os.Setenv(key, value)
	}
}

// GetEnv returns an environment variable with a fallback default.
func GetEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// TableName returns the configured table name.
func TableName() string {
	return "data"
}

// DBName returns the configured database name.
func DBName() string {
	return "scmd"
}

// ConfigFilePath returns the expected config file location for display purposes.
func ConfigFilePath() string {
	return configPath()
}

// ConfigDir returns the directory containing the config file (~/.scmd).
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".scmd")
}

// SaveConfig writes the given ConfigData to ~/.scmd/config.json.
// It creates the directory if it doesn't exist.
func SaveConfig(cfg *ConfigData) error {
	dir := ConfigDir()
	if dir == "" {
		return fmt.Errorf("cannot determine home directory")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create config directory: %v", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %v", err)
	}

	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("cannot write config: %v", err)
	}
	return nil
}

// CurrentConfig reads and returns the current config, or an empty one if not found.
func CurrentConfig() *ConfigData {
	path := configPath()
	if path == "" {
		return &ConfigData{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return &ConfigData{}
	}
	var cfg ConfigData
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &ConfigData{}
	}
	return &cfg
}
