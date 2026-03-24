package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"github.com/gcclinux/scmd/internal/config"
)

var db *sql.DB

// DB returns the current database connection for use by other packages.
func DB() *sql.DB {
	return db
}

// IsPostgreSQL returns true if the configured database type is PostgreSQL.
func IsPostgreSQL() bool {
	return strings.ToLower(os.Getenv("DB_TYPE")) != "sqlite"
}

// InitDB initializes the database connection based on the configured db_type.
func InitDB() error {
	config.LoadConfig()

	dbType := strings.ToLower(os.Getenv("DB_TYPE"))
	if dbType == "sqlite" {
		return InitSQLiteDB()
	}
	return initPostgresDB()
}

// initPostgresDB initializes the PostgreSQL database connection.
func initPostgresDB() error {
	var err error

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return nil
}

// CloseDB closes the database connection.
func CloseDB() {
	if db != nil {
		db.Close()
	}
}
