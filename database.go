package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type CommandRecord struct {
	Id   int    `json:"id"`
	Key  string `json:"key"`
	Data string `json:"data"`
}

var db *sql.DB

// InitDB initializes the PostgreSQL database connection
func InitDB() error {
	// Load .env file from executable's directory
	execPath, err := os.Executable()
	if err != nil {
		log.Println("Warning: could not determine executable path, trying current directory")
		execPath = "."
	}
	execDir := filepath.Dir(execPath)
	envPath := filepath.Join(execDir, ".env")

	if loadErr := godotenv.Load(envPath); loadErr != nil {
		log.Printf("Warning: .env file not found at %s, using environment variables\n", envPath)
	} else {
		log.Printf("Loaded .env from: %s\n", envPath)
	}

	// Get database configuration from environment
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if db != nil {
		db.Close()
	}
}

// SearchCommands searches for commands matching the pattern
func SearchCommands(pattern string, format string) ([]byte, error) {
	tableName := os.Getenv("TB_NAME")
	if tableName == "" {
		tableName = "scmd"
	}

	var query string
	var args []interface{}

	// If pattern is empty, return all commands
	if pattern == "" {
		query = fmt.Sprintf("SELECT id, key, data FROM %s ORDER BY id", tableName)
	} else {
		// Check if pattern contains comma (OR search) or spaces (AND search)
		if strings.Contains(pattern, ",") {
			// Comma-separated: OR logic (match any pattern)
			patterns := strings.Split(pattern, ",")
			var conditions []string

			for _, p := range patterns {
				p = strings.TrimSpace(p)
				if p != "" {
					// For each comma-separated pattern, check if it contains spaces
					if strings.Contains(p, " ") {
						// Multiple words in this pattern - all must match (AND)
						words := strings.Fields(p)
						var wordConditions []string
						for _, word := range words {
							argIndex := len(args) + 1
							wordConditions = append(wordConditions,
								fmt.Sprintf("(key ILIKE $%d OR data ILIKE $%d)", argIndex, argIndex+1))
							searchPattern := "%" + word + "%"
							args = append(args, searchPattern, searchPattern)
						}
						conditions = append(conditions, "("+strings.Join(wordConditions, " AND ")+")")
					} else {
						// Single word pattern
						argIndex := len(args) + 1
						conditions = append(conditions, fmt.Sprintf("(key ILIKE $%d OR data ILIKE $%d)", argIndex, argIndex+1))
						searchPattern := "%" + p + "%"
						args = append(args, searchPattern, searchPattern)
					}
				}
			}

			if len(conditions) == 0 {
				query = fmt.Sprintf("SELECT id, key, data FROM %s ORDER BY id", tableName)
			} else {
				query = fmt.Sprintf("SELECT id, key, data FROM %s WHERE %s ORDER BY id",
					tableName, strings.Join(conditions, " OR "))
			}
		} else {
			// No comma: space-separated words - AND logic (all words must match)
			words := strings.Fields(pattern)
			var conditions []string

			for _, word := range words {
				word = strings.TrimSpace(word)
				if word != "" {
					argIndex := len(args) + 1
					conditions = append(conditions, fmt.Sprintf("(key ILIKE $%d OR data ILIKE $%d)", argIndex, argIndex+1))
					searchPattern := "%" + word + "%"
					args = append(args, searchPattern, searchPattern)
				}
			}

			if len(conditions) == 0 {
				query = fmt.Sprintf("SELECT id, key, data FROM %s ORDER BY id", tableName)
			} else {
				// Use AND to ensure all words are present
				query = fmt.Sprintf("SELECT id, key, data FROM %s WHERE %s ORDER BY id",
					tableName, strings.Join(conditions, " AND "))
			}
		}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %v", err)
	}
	defer rows.Close()

	var results []CommandRecord
	for rows.Next() {
		var record CommandRecord
		if err := rows.Scan(&record.Id, &record.Key, &record.Data); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		results = append(results, record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(results)
	if err != nil {
		return nil, fmt.Errorf("error marshaling to JSON: %v", err)
	}

	return jsonData, nil
}

// AddCommand adds a new command to the database
func AddCommand(command, description string) (bool, error) {
	tableName := os.Getenv("TB_NAME")
	if tableName == "" {
		tableName = "scmd"
	}

	var embedding []float64
	var embeddingErr error
	hasEmbedding := false

	// Try Ollama first
	if IsOllamaAvailable() {
		text := command + " " + description
		embedding, embeddingErr = GetEmbedding(text)
		if embeddingErr != nil {
			log.Printf("Warning: Ollama embedding failed: %v\n", embeddingErr)
		} else {
			hasEmbedding = true
			log.Println("✓ Generated embedding using Ollama")
		}
	}

	// Fallback to Gemini if Ollama failed or unavailable
	if !hasEmbedding && IsGeminiAvailable() {
		text := command + " " + description
		embedding, embeddingErr = GetGeminiEmbedding(text)
		if embeddingErr != nil {
			log.Printf("Warning: Gemini embedding failed: %v\n", embeddingErr)
		} else {
			hasEmbedding = true
			log.Println("✓ Generated embedding using Gemini API")
		}
	}

	// Insert command with or without embedding
	if hasEmbedding {
		// Convert embedding to PostgreSQL vector format
		embeddingStr := "["
		for i, val := range embedding {
			if i > 0 {
				embeddingStr += ","
			}
			embeddingStr += fmt.Sprintf("%f", val)
		}
		embeddingStr += "]"

		// Insert with embedding
		query := fmt.Sprintf("INSERT INTO %s (key, data, embedding) VALUES ($1, $2, $3::vector)", tableName)
		_, err := db.Exec(query, command, description, embeddingStr)
		if err != nil {
			return false, fmt.Errorf("error inserting command with embedding: %v", err)
		}
	} else {
		// Insert without embedding
		log.Println("⚠ No embedding provider available, saving without vector")
		query := fmt.Sprintf("INSERT INTO %s (key, data) VALUES ($1, $2)", tableName)
		_, err := db.Exec(query, command, description)
		if err != nil {
			return false, fmt.Errorf("error inserting command: %v", err)
		}
	}

	return true, nil
}

// CheckCommandExists checks if a command already exists in the database
func CheckCommandExists(command string) (bool, error) {
	tableName := os.Getenv("TB_NAME")
	if tableName == "" {
		tableName = "scmd"
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE key = $1", tableName)
	var count int
	err := db.QueryRow(query, command).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking command existence: %v", err)
	}

	return count > 0, nil
}
