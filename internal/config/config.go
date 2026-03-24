package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigData holds all configuration fields from config.json.
type ConfigData struct {
	Agent               string `json:"agent"`
	DBHost              string `json:"db_host"`
	DBPort              string `json:"db_port"`
	DBUser              string `json:"db_user"`
	DBPass              string `json:"db_pass"`
	DBName              string `json:"db_name"`
	TBName              string `json:"tb_name"`
	GeminiAPI           string `json:"gemini_api"`
	GeminiModel         string `json:"gemini_model"`
	GeminiEmbeddingModel string `json:"gemini_embedding_model"`
	Ollama              string `json:"ollama"`
	Model               string `json:"model"`
	EmbeddingModel      string `json:"embedding_model"`
	EmbeddingDim        string `json:"embedding_dim"`
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
	setIfNotEmpty("DB_HOST", cfg.DBHost)
	setIfNotEmpty("DB_PORT", cfg.DBPort)
	setIfNotEmpty("DB_USER", cfg.DBUser)
	setIfNotEmpty("DB_PASS", cfg.DBPass)
	setIfNotEmpty("DB_NAME", cfg.DBName)
	setIfNotEmpty("TB_NAME", cfg.TBName)
	setIfNotEmpty("GEMINIAPI", cfg.GeminiAPI)
	setIfNotEmpty("GEMINIMODEL", cfg.GeminiModel)
	setIfNotEmpty("GEMINI_EMBEDDING_MODEL", cfg.GeminiEmbeddingModel)
	setIfNotEmpty("OLLAMA", cfg.Ollama)
	setIfNotEmpty("MODEL", cfg.Model)
	setIfNotEmpty("EMBEDDING_MODEL", cfg.EmbeddingModel)
	setIfNotEmpty("EMBEDDING_DIM", cfg.EmbeddingDim)
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

// TableName returns the configured table name from TB_NAME or the default.
func TableName() string {
	return GetEnv("TB_NAME", "scmd")
}

// ConfigFilePath returns the expected config file location for display purposes.
func ConfigFilePath() string {
	return configPath()
}
