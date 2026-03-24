package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadEnv loads the .env file from the executable's directory or a parent.
func LoadEnv() {
	// 1. Try executable's directory first (resolving symlinks)
	execPath, err := os.Executable()
	if err == nil {
		execPath, _ = filepath.EvalSymlinks(execPath)
		execDir := filepath.Dir(execPath)

		// Try the exe directory itself, then walk up to parent directories (max 3 levels)
		dir := execDir
		for i := 0; i < 4; i++ {
			envPath := filepath.Join(dir, ".env")
			if _, statErr := os.Stat(envPath); statErr == nil {
				if loadErr := godotenv.Load(envPath); loadErr == nil {
					return
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	// 2. Fallback: try current working directory
	cwd, err := os.Getwd()
	if err == nil {
		envPath := filepath.Join(cwd, ".env")
		godotenv.Load(envPath)
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
