package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gcclinux/scmd/internal/config"
)

// SearchCommands searches for commands matching the pattern.
func SearchCommands(pattern string, format string) ([]byte, error) {
	if IsMCP() {
		return searchCommandsMCP(pattern)
	}
	if !IsPostgreSQL() {
		return searchCommandsSQLite(pattern)
	}
	tableName := config.TableName()

	var query string
	var args []interface{}

	if pattern == "" {
		query = fmt.Sprintf("SELECT id, key, data FROM %s ORDER BY id", tableName)
	} else {
		if strings.Contains(pattern, ",") {
			// Comma-separated: OR logic
			patterns := strings.Split(pattern, ",")
			var conditions []string

			for _, p := range patterns {
				p = strings.TrimSpace(p)
				if p != "" {
					if strings.Contains(p, " ") {
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
			// Space-separated: AND logic
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

	jsonData, err := json.Marshal(results)
	if err != nil {
		return nil, fmt.Errorf("error marshaling to JSON: %v", err)
	}

	return jsonData, nil
}

// AddCommand adds a new command to the database.
// embeddingFn is an optional callback to generate embeddings.
func AddCommand(command, description string, embeddingFn func(string) ([]float64, error)) (bool, error) {
	if IsMCP() {
		return addCommandMCP(command, description, embeddingFn)
	}
	if !IsPostgreSQL() {
		return addCommandSQLite(command, description, embeddingFn)
	}
	tableName := config.TableName()

	var embedding []float64
	hasEmbedding := false

	if embeddingFn != nil {
		text := command + " " + description
		emb, err := embeddingFn(text)
		if err != nil {
			log.Printf("Warning: embedding generation failed: %v\n", err)
		} else {
			embedding = emb
			hasEmbedding = true
		}
	}

	// Get next available id
	var nextID int
	idQuery := fmt.Sprintf("SELECT COALESCE(MAX(id), 0) + 1 FROM %s", tableName)
	err := db.QueryRow(idQuery).Scan(&nextID)
	if err != nil {
		return false, fmt.Errorf("error getting next id: %v", err)
	}

	if hasEmbedding {
		embeddingStr := FormatEmbedding(embedding)
		query := fmt.Sprintf("INSERT INTO %s (id, key, data, embedding) VALUES ($1, $2, $3, $4::vector)", tableName)
		_, err := db.Exec(query, nextID, command, description, embeddingStr)
		if err != nil {
			return false, fmt.Errorf("error inserting command with embedding: %v", err)
		}
		log.Println("✓ Generated embedding for new command")
	} else {
		log.Println("⚠ No embedding provider available, saving without vector")
		query := fmt.Sprintf("INSERT INTO %s (id, key, data) VALUES ($1, $2, $3)", tableName)
		_, err := db.Exec(query, nextID, command, description)
		if err != nil {
			return false, fmt.Errorf("error inserting command: %v", err)
		}
	}

	return true, nil
}

// CheckCommandExists checks if a command already exists in the database.
func CheckCommandExists(command string) (bool, error) {
	if IsMCP() {
		return checkCommandExistsMCP(command)
	}
	if !IsPostgreSQL() {
		return checkCommandExistsSQLite(command)
	}
	tableName := config.TableName()

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE key = $1", tableName)
	var count int
	err := db.QueryRow(query, command).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking command existence: %v", err)
	}

	return count > 0, nil
}

// DeleteCommand deletes a command from the database by ID.
func DeleteCommand(id int) (bool, error) {
	if IsMCP() {
		return deleteCommandMCP(id)
	}
	if !IsPostgreSQL() {
		return deleteCommandSQLite(id)
	}
	tableName := config.TableName()

	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName)
	result, err := db.Exec(query, id)
	if err != nil {
		return false, fmt.Errorf("error deleting command: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("error checking affected rows: %v", err)
	}

	return rows > 0, nil
}

// GetCommandByID retrieves a single command record by its ID.
func GetCommandByID(id int) (*CommandRecord, error) {
	if IsMCP() {
		return getCommandByIDMCP(id)
	}
	if !IsPostgreSQL() {
		return getCommandByIDSQLite(id)
	}
	tableName := config.TableName()

	query := fmt.Sprintf("SELECT id, key, data FROM %s WHERE id = $1", tableName)
	var record CommandRecord
	err := db.QueryRow(query, id).Scan(&record.Id, &record.Key, &record.Data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no command found with ID %d", id)
		}
		return nil, fmt.Errorf("error querying command: %v", err)
	}

	return &record, nil
}

// GetCommandsWithoutEmbeddings returns all commands that have no embedding.
func GetCommandsWithoutEmbeddings() ([]CommandRecord, error) {
	if IsMCP() {
		return getCommandsWithoutEmbeddingsMCP()
	}
	if !IsPostgreSQL() {
		return getCommandsWithoutEmbeddingsSQLite()
	}
	tableName := config.TableName()
	query := fmt.Sprintf("SELECT id, key, data FROM %s WHERE embedding IS NULL", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %v", err)
	}
	defer rows.Close()

	var commands []CommandRecord
	for rows.Next() {
		var record CommandRecord
		if err := rows.Scan(&record.Id, &record.Key, &record.Data); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		commands = append(commands, record)
	}
	return commands, nil
}

// UpdateEmbedding updates the embedding for a command by ID.
func UpdateEmbedding(id int, embedding []float64) error {
	if IsMCP() {
		return updateEmbeddingMCP(id, embedding)
	}
	if !IsPostgreSQL() {
		return updateEmbeddingSQLite(id, embedding)
	}
	tableName := config.TableName()
	embeddingStr := FormatEmbedding(embedding)
	query := fmt.Sprintf("UPDATE %s SET embedding = $1::vector WHERE id = $2", tableName)
	_, err := db.Exec(query, embeddingStr, id)
	return err
}

// GetEmbeddingStats returns total commands and count with embeddings.
func GetEmbeddingStats() (total int, withEmbeddings int, err error) {
	if IsMCP() {
		return getEmbeddingStatsMCP()
	}
	if !IsPostgreSQL() {
		return getEmbeddingStatsSQLite()
	}
	tableName := config.TableName()

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	if err = db.QueryRow(query).Scan(&total); err != nil {
		return 0, 0, fmt.Errorf("error counting commands: %v", err)
	}

	query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE embedding IS NOT NULL", tableName)
	if err = db.QueryRow(query).Scan(&withEmbeddings); err != nil {
		return 0, 0, fmt.Errorf("error counting embeddings: %v", err)
	}

	return total, withEmbeddings, nil
}

// SearchByVector performs a vector similarity search using pgvector.
func SearchByVector(embedding []float64, limit int) ([]CommandRecord, error) {
	if IsMCP() {
		return searchByVectorMCP(embedding, limit)
	}
	if !IsPostgreSQL() {
		return searchByVectorSQLite(embedding, limit)
	}
	tableName := config.TableName()
	embeddingStr := FormatEmbedding(embedding)

	query := fmt.Sprintf(`
		SELECT id, key, data
		FROM %s
		WHERE embedding IS NOT NULL
		ORDER BY embedding <-> $1::vector
		LIMIT $2`, tableName)

	rows, err := db.Query(query, embeddingStr, limit)
	if err != nil {
		return nil, fmt.Errorf("vector search error: %v", err)
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
	return results, nil
}

// AuthenticateUser validates email and API key against the database.
func AuthenticateUser(email, apiKey string) (bool, error) {
	if IsMCP() {
		return false, fmt.Errorf("authentication not supported with MCP backend")
	}
	if !IsPostgreSQL() {
		return authenticateUserSQLite(email, apiKey)
	}
	email = strings.TrimSpace(email)
	apiKey = strings.TrimSpace(apiKey)

	if email == "" || apiKey == "" {
		return false, fmt.Errorf("email and API key are required")
	}

	if db == nil {
		return false, fmt.Errorf("database not connected")
	}

	accessTbl := os.Getenv("ACCESS_TB")
	if accessTbl == "" {
		accessTbl = "access"
	}

	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE email = $1 AND api_key = $2", accessTbl)
	err := db.QueryRow(query, email, apiKey).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("authentication query failed: %v", err)
	}

	return count > 0, nil
}

// FormatEmbedding converts a float64 slice to PostgreSQL vector format.
func FormatEmbedding(embedding []float64) string {
	embeddingStr := "["
	for i, val := range embedding {
		if i > 0 {
			embeddingStr += ","
		}
		embeddingStr += fmt.Sprintf("%f", val)
	}
	embeddingStr += "]"
	return embeddingStr
}
