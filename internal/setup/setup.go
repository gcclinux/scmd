package setup

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gcclinux/scmd/internal/config"
	"github.com/gcclinux/scmd/internal/database"
)

const (
	noticeColor = "\033[1;36m%s\033[0m"
	successColor = "\033[1;32m%s\033[0m"
	errorColor  = "\033[1;31m%s\033[0m"
)

// prompt reads a line from stdin with a default value shown in brackets.
func prompt(reader *bufio.Reader, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("  %s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("  %s: ", label)
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}

// promptPassword reads a password, showing asterisks as hint.
func promptPassword(reader *bufio.Reader, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("  %s [****]: ", label)
	} else {
		fmt.Printf("  %s: ", label)
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}

// SetupPostgreSQL runs the interactive PostgreSQL database setup wizard.
func SetupPostgreSQL() {
	reader := bufio.NewReader(os.Stdin)
	cfg := config.CurrentConfig()

	fmt.Println()
	fmt.Printf(noticeColor, "=== PostgreSQL Database Setup ===\n")
	fmt.Println()
	fmt.Println("  This will configure and create a PostgreSQL database for SCMD.")
	fmt.Println("  Press Enter to accept the default value shown in brackets.")
	fmt.Println()

	// Collect connection details
	if cfg.DBHost == "" {
		cfg.DBHost = "localhost"
	}
	cfg.DBHost = prompt(reader, "Database host", cfg.DBHost)

	if cfg.DBPort == "" {
		cfg.DBPort = "5432"
	}
	cfg.DBPort = prompt(reader, "Database port", cfg.DBPort)

	cfg.DBUser = prompt(reader, "Database user", cfg.DBUser)
	cfg.DBPass = promptPassword(reader, "Database password", cfg.DBPass)

	if cfg.DBName == "" {
		cfg.DBName = "scmd"
	}
	cfg.DBName = prompt(reader, "Database name", cfg.DBName)

	if cfg.TBName == "" {
		cfg.TBName = "data"
	}
	cfg.TBName = prompt(reader, "Table name", cfg.TBName)

	if cfg.EmbeddingDim == "" {
		cfg.EmbeddingDim = "384"
	}
	cfg.EmbeddingDim = prompt(reader, "Embedding dimension", cfg.EmbeddingDim)

	// Set db_type
	cfg.DBType = "postgresql"

	// Save config first
	fmt.Println()
	fmt.Printf(noticeColor, "Saving configuration...\n")
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Printf(errorColor, fmt.Sprintf("Error saving config: %v\n", err))
		os.Exit(1)
	}
	fmt.Printf(successColor, fmt.Sprintf("  ✓ Config saved to %s\n", config.ConfigFilePath()))

	// Reload config so env vars are set for database setup
	config.LoadConfig()

	// Create the database
	fmt.Println()
	fmt.Printf(noticeColor, "Creating PostgreSQL database...\n")
	database.SetupDatabase()
}

// SetupSQLite runs the interactive SQLite database setup wizard.
func SetupSQLite() {
	reader := bufio.NewReader(os.Stdin)
	cfg := config.CurrentConfig()

	fmt.Println()
	fmt.Printf(noticeColor, "=== SQLite Database Setup ===\n")
	fmt.Println()
	fmt.Println("  This will configure and create a SQLite database for SCMD.")
	fmt.Println("  The database file will be stored in ~/.scmd/")
	fmt.Println("  Press Enter to accept the default value shown in brackets.")
	fmt.Println()

	if cfg.DBName == "" {
		cfg.DBName = "scmd"
	}
	cfg.DBName = prompt(reader, "Database name (filename without .db)", cfg.DBName)

	if cfg.TBName == "" {
		cfg.TBName = "data"
	}
	cfg.TBName = prompt(reader, "Table name", cfg.TBName)

	if cfg.EmbeddingDim == "" {
		cfg.EmbeddingDim = "384"
	}
	cfg.EmbeddingDim = prompt(reader, "Embedding dimension", cfg.EmbeddingDim)

	// Set db_type and clear PostgreSQL-specific fields
	cfg.DBType = "sqlite"
	cfg.DBHost = ""
	cfg.DBPort = ""
	cfg.DBUser = ""
	cfg.DBPass = ""

	// Save config
	fmt.Println()
	fmt.Printf(noticeColor, "Saving configuration...\n")
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Printf(errorColor, fmt.Sprintf("Error saving config: %v\n", err))
		os.Exit(1)
	}
	fmt.Printf(successColor, fmt.Sprintf("  ✓ Config saved to %s\n", config.ConfigFilePath()))

	// Reload config so env vars are set
	config.LoadConfig()

	// Create the database
	fmt.Println()
	fmt.Printf(noticeColor, "Creating SQLite database...\n")
	database.SetupSQLiteDatabase()
}

// SetupOllama runs the interactive Ollama AI server setup wizard.
func SetupOllama() {
	reader := bufio.NewReader(os.Stdin)
	cfg := config.CurrentConfig()

	fmt.Println()
	fmt.Printf(noticeColor, "=== Ollama AI Server Setup ===\n")
	fmt.Println()
	fmt.Println("  This will configure Ollama as the AI provider for SCMD.")
	fmt.Println("  Make sure Ollama is running before using AI features.")
	fmt.Println("  Press Enter to accept the default value shown in brackets.")
	fmt.Println()

	cfg.Agent = "ollama"

	if cfg.Ollama == "" {
		cfg.Ollama = "localhost"
	}
	cfg.Ollama = prompt(reader, "Ollama host", cfg.Ollama)

	if cfg.Model == "" {
		cfg.Model = "llama2"
	}
	cfg.Model = prompt(reader, "Chat model", cfg.Model)

	if cfg.EmbeddingModel == "" {
		cfg.EmbeddingModel = cfg.Model
	}
	cfg.EmbeddingModel = prompt(reader, "Embedding model", cfg.EmbeddingModel)

	if cfg.EmbeddingDim == "" {
		cfg.EmbeddingDim = "384"
	}
	cfg.EmbeddingDim = prompt(reader, "Embedding dimension", cfg.EmbeddingDim)

	// Save config
	fmt.Println()
	fmt.Printf(noticeColor, "Saving configuration...\n")
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Printf(errorColor, fmt.Sprintf("Error saving config: %v\n", err))
		os.Exit(1)
	}
	fmt.Printf(successColor, fmt.Sprintf("  ✓ Config saved to %s\n", config.ConfigFilePath()))

	fmt.Println()
	fmt.Printf(successColor, "=== Ollama Setup Complete ===\n")
	fmt.Println()
	fmt.Println("  Agent:           ollama")
	fmt.Printf("  Host:            %s\n", cfg.Ollama)
	fmt.Printf("  Chat model:      %s\n", cfg.Model)
	fmt.Printf("  Embedding model: %s\n", cfg.EmbeddingModel)
	fmt.Printf("  Embedding dim:   %s\n", cfg.EmbeddingDim)
	fmt.Println()
	fmt.Println("  Ensure Ollama is running: ollama serve")
	fmt.Println("  Pull your models:         ollama pull", cfg.Model)
	if cfg.EmbeddingModel != cfg.Model {
		fmt.Println("                            ollama pull", cfg.EmbeddingModel)
	}
	fmt.Println()
}

// SetupGemini runs the interactive Gemini AI server setup wizard.
func SetupGemini() {
	reader := bufio.NewReader(os.Stdin)
	cfg := config.CurrentConfig()

	fmt.Println()
	fmt.Printf(noticeColor, "=== Gemini AI Server Setup ===\n")
	fmt.Println()
	fmt.Println("  This will configure Google Gemini as the AI provider for SCMD.")
	fmt.Println("  You need a valid Gemini API key from Google AI Studio.")
	fmt.Println("  Press Enter to accept the default value shown in brackets.")
	fmt.Println()

	cfg.Agent = "gemini"

	cfg.GeminiAPI = promptPassword(reader, "Gemini API key", cfg.GeminiAPI)
	if cfg.GeminiAPI == "" {
		fmt.Printf(errorColor, "  Error: API key is required for Gemini setup.\n")
		os.Exit(1)
	}

	if cfg.GeminiModel == "" {
		cfg.GeminiModel = "gemini-2.5-flash-lite"
	}
	cfg.GeminiModel = prompt(reader, "Chat model", cfg.GeminiModel)

	if cfg.GeminiEmbeddingModel == "" {
		cfg.GeminiEmbeddingModel = "gemini-embedding-001"
	}
	cfg.GeminiEmbeddingModel = prompt(reader, "Embedding model", cfg.GeminiEmbeddingModel)

	if cfg.EmbeddingDim == "" {
		cfg.EmbeddingDim = "384"
	}
	cfg.EmbeddingDim = prompt(reader, "Embedding dimension", cfg.EmbeddingDim)

	// Save config
	fmt.Println()
	fmt.Printf(noticeColor, "Saving configuration...\n")
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Printf(errorColor, fmt.Sprintf("Error saving config: %v\n", err))
		os.Exit(1)
	}
	fmt.Printf(successColor, fmt.Sprintf("  ✓ Config saved to %s\n", config.ConfigFilePath()))

	fmt.Println()
	fmt.Printf(successColor, "=== Gemini Setup Complete ===\n")
	fmt.Println()
	fmt.Println("  Agent:           gemini")
	fmt.Printf("  Chat model:      %s\n", cfg.GeminiModel)
	fmt.Printf("  Embedding model: %s\n", cfg.GeminiEmbeddingModel)
	fmt.Printf("  Embedding dim:   %s\n", cfg.EmbeddingDim)
	fmt.Println("  API key:         ****")
	fmt.Println()
}
