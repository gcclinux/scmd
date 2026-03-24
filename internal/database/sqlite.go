package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gcclinux/scmd/internal/config"
	_ "modernc.org/sqlite"
)

// SQLitePath returns the full path to the SQLite database file.
func SQLitePath() string {
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "scmd"
	}
	return filepath.Join(config.ConfigDir(), dbName+".db")
}

// InitSQLiteDB initializes a SQLite database connection.
func InitSQLiteDB() error {
	config.LoadConfig()
	var err error

	dbPath := SQLitePath()
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("error opening SQLite database: %v", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("error connecting to SQLite database: %v", err)
	}

	// Enable WAL mode for better concurrent access
	if _, err = db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		log.Printf("Warning: could not enable WAL mode: %v", err)
	}

	log.Println("Successfully connected to SQLite database:", dbPath)
	return nil
}

// SetupSQLiteDatabase creates the SQLite database and tables from scratch.
func SetupSQLiteDatabase() {
	config.LoadConfig()

	dbPath := SQLitePath()
	dataTbl := os.Getenv("TB_NAME")
	if dataTbl == "" {
		dataTbl = "data"
	}

	fmt.Printf("\n=== Step 1: Create SQLite database '%s' ===\n", dbPath)

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Cannot create SQLite database: %v", err)
	}
	defer conn.Close()

	if err = conn.Ping(); err != nil {
		log.Fatalf("Cannot ping SQLite database: %v", err)
	}
	fmt.Printf("  Database file: %s\n", dbPath)

	// Enable WAL mode
	if _, err = conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		log.Printf("Warning: could not enable WAL mode: %v", err)
	}

	// Create the main commands table (no vector type in SQLite, use TEXT for embeddings)
	fmt.Printf("\n=== Step 2: Create table '%s' ===\n", dataTbl)
	createSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			key        TEXT    NOT NULL,
			data       TEXT    NOT NULL,
			embedding  TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`, dataTbl)

	if _, err = conn.Exec(createSQL); err != nil {
		log.Fatalf("Failed to create table '%s': %v", dataTbl, err)
	}
	fmt.Printf("  Table '%s' created.\n", dataTbl)

	fmt.Println()
	fmt.Println("======================================================")
	fmt.Println("  SQLite database setup complete!")
	fmt.Println()
	fmt.Printf("  Database file: %s\n", dbPath)
	fmt.Printf("  Table name:    %s\n", dataTbl)
	fmt.Println()
	fmt.Println("  Note: SQLite stores embeddings as JSON text.")
	fmt.Println("  Vector search uses cosine similarity in Go.")
	fmt.Println("  Run --generate-embeddings after adding data.")
	fmt.Println("======================================================")
}
