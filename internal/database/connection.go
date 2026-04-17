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

// mcpClient holds the active MCP client as an untyped reference to avoid
// an import cycle between the database and mcpclient packages. Code in
// queries_mcp.go accesses the concrete *mcpclient.Client via MCPClient().
var mcpClient any

// MCPInitFunc is the function used to initialize the MCP client connection.
// It must be set before InitDB() is called with db_type=mcp. The main
// package wires this to mcpclient.Init to break the import cycle.
// Returns (client, error) where client is *mcpclient.Client stored as any.
var MCPInitFunc func(configPath string) (any, error)

// MCPCloseFunc is the function used to close the MCP client connection.
// It is set alongside MCPInitFunc by the main package.
var MCPCloseFunc func(client any)

// DB returns the current database connection for use by other packages.
func DB() *sql.DB {
	return db
}

// IsPostgreSQL returns true if the configured database type is PostgreSQL.
func IsPostgreSQL() bool {
	dbType := strings.ToLower(os.Getenv("DB_TYPE"))
	return dbType != "sqlite" && dbType != "mcp"
}

// IsMCP returns true if the configured database type is MCP.
func IsMCP() bool {
	return strings.ToLower(os.Getenv("DB_TYPE")) == "mcp"
}

// MCPClient returns the active MCP client for use by queries_mcp.go.
// The caller must type-assert to *mcpclient.Client.
func MCPClient() any {
	return mcpClient
}

// InitDB initializes the database connection based on the configured db_type.
func InitDB() error {
	config.LoadConfig()

	dbType := strings.ToLower(os.Getenv("DB_TYPE"))
	switch dbType {
	case "sqlite":
		return InitSQLiteDB()
	case "mcp":
		return initMCPDB()
	default:
		return initPostgresDB()
	}
}

// initMCPDB initializes the MCP client connection.
func initMCPDB() error {
	if MCPInitFunc == nil {
		return fmt.Errorf("MCP init function not registered — ensure mcpclient bridge is wired")
	}
	client, err := MCPInitFunc(os.Getenv("MCP_SERVER"))
	if err != nil {
		return fmt.Errorf("error initializing MCP client: %v", err)
	}
	mcpClient = client
	log.Println("Successfully connected to MCP server")
	return nil
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
	if mcpClient != nil && MCPCloseFunc != nil {
		MCPCloseFunc(mcpClient)
		mcpClient = nil
	}
}
