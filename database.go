package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
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

	var err error
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

	// Split pattern by comma for multiple search terms
	patterns := strings.Split(pattern, ",")
	var conditions []string
	var args []interface{}

	for i, p := range patterns {
		p = strings.TrimSpace(p)
		if p != "" {
			conditions = append(conditions, fmt.Sprintf("(key ILIKE $%d OR data ILIKE $%d)", i*2+1, i*2+2))
			searchPattern := "%" + p + "%"
			args = append(args, searchPattern, searchPattern)
		}
	}

	query := fmt.Sprintf("SELECT id, key, data FROM %s WHERE %s ORDER BY id",
		tableName, strings.Join(conditions, " OR "))

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

	query := fmt.Sprintf("INSERT INTO %s (key, data) VALUES ($1, $2)", tableName)
	_, err := db.Exec(query, command, description)
	if err != nil {
		return false, fmt.Errorf("error inserting command: %v", err)
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
