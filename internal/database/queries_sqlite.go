package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/gcclinux/scmd/internal/config"
)

// searchCommandsSQLite searches for commands in SQLite using LIKE (case-insensitive via COLLATE NOCASE).
func searchCommandsSQLite(pattern string) ([]byte, error) {
	tableName := sqliteTableName()

	var query string
	var args []interface{}

	if pattern == "" {
		query = fmt.Sprintf("SELECT id, key, data FROM %s ORDER BY id", tableName)
	} else {
		if strings.Contains(pattern, ",") {
			patterns := strings.Split(pattern, ",")
			var conditions []string
			for _, p := range patterns {
				p = strings.TrimSpace(p)
				if p != "" {
					if strings.Contains(p, " ") {
						words := strings.Fields(p)
						var wordConditions []string
						for _, word := range words {
							searchPattern := "%" + word + "%"
							wordConditions = append(wordConditions,
								"(key LIKE ? COLLATE NOCASE OR data LIKE ? COLLATE NOCASE)")
							args = append(args, searchPattern, searchPattern)
						}
						conditions = append(conditions, "("+strings.Join(wordConditions, " AND ")+")")
					} else {
						searchPattern := "%" + p + "%"
						conditions = append(conditions, "(key LIKE ? COLLATE NOCASE OR data LIKE ? COLLATE NOCASE)")
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
			words := strings.Fields(pattern)
			var conditions []string
			for _, word := range words {
				word = strings.TrimSpace(word)
				if word != "" {
					searchPattern := "%" + word + "%"
					conditions = append(conditions, "(key LIKE ? COLLATE NOCASE OR data LIKE ? COLLATE NOCASE)")
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

	return json.Marshal(results)
}

// addCommandSQLite adds a new command to the SQLite database.
func addCommandSQLite(command, description string, embeddingFn func(string) ([]float64, error)) (bool, error) {
	tableName := sqliteTableName()

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

	if hasEmbedding {
		embeddingJSON, err := json.Marshal(embedding)
		if err != nil {
			return false, fmt.Errorf("error marshaling embedding: %v", err)
		}
		query := fmt.Sprintf("INSERT INTO %s (key, data, embedding) VALUES (?, ?, ?)", tableName)
		_, err = db.Exec(query, command, description, string(embeddingJSON))
		if err != nil {
			return false, fmt.Errorf("error inserting command with embedding: %v", err)
		}
		log.Println("✓ Generated embedding for new command")
	} else {
		log.Println("⚠ No embedding provider available, saving without vector")
		query := fmt.Sprintf("INSERT INTO %s (key, data) VALUES (?, ?)", tableName)
		_, err := db.Exec(query, command, description)
		if err != nil {
			return false, fmt.Errorf("error inserting command: %v", err)
		}
	}

	return true, nil
}

// checkCommandExistsSQLite checks if a command exists in SQLite.
func checkCommandExistsSQLite(command string) (bool, error) {
	tableName := sqliteTableName()
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE key = ?", tableName)
	var count int
	err := db.QueryRow(query, command).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking command existence: %v", err)
	}
	return count > 0, nil
}

// deleteCommandSQLite deletes a command from SQLite by ID.
func deleteCommandSQLite(id int) (bool, error) {
	tableName := sqliteTableName()
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)
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

// getCommandByIDSQLite retrieves a single command record by ID from SQLite.
func getCommandByIDSQLite(id int) (*CommandRecord, error) {
	tableName := sqliteTableName()
	query := fmt.Sprintf("SELECT id, key, data FROM %s WHERE id = ?", tableName)
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

// getCommandsWithoutEmbeddingsSQLite returns commands without embeddings from SQLite.
func getCommandsWithoutEmbeddingsSQLite() ([]CommandRecord, error) {
	tableName := sqliteTableName()
	query := fmt.Sprintf("SELECT id, key, data FROM %s WHERE embedding IS NULL OR embedding = ''", tableName)
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

// updateEmbeddingSQLite updates the embedding for a command in SQLite.
func updateEmbeddingSQLite(id int, embedding []float64) error {
	tableName := sqliteTableName()
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("error marshaling embedding: %v", err)
	}
	query := fmt.Sprintf("UPDATE %s SET embedding = ? WHERE id = ?", tableName)
	_, err = db.Exec(query, string(embeddingJSON), id)
	return err
}

// getEmbeddingStatsSQLite returns total commands and count with embeddings for SQLite.
func getEmbeddingStatsSQLite() (total int, withEmbeddings int, err error) {
	tableName := sqliteTableName()

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	if err = db.QueryRow(query).Scan(&total); err != nil {
		return 0, 0, fmt.Errorf("error counting commands: %v", err)
	}

	query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE embedding IS NOT NULL AND embedding != ''", tableName)
	if err = db.QueryRow(query).Scan(&withEmbeddings); err != nil {
		return 0, 0, fmt.Errorf("error counting embeddings: %v", err)
	}

	return total, withEmbeddings, nil
}

// searchByVectorSQLite performs cosine similarity search in SQLite.
func searchByVectorSQLite(embedding []float64, limit int) ([]CommandRecord, error) {
	tableName := sqliteTableName()

	// Fetch all rows with embeddings and compute similarity in Go
	query := fmt.Sprintf("SELECT id, key, data, embedding FROM %s WHERE embedding IS NOT NULL AND embedding != ''", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("vector search error: %v", err)
	}
	defer rows.Close()

	type scored struct {
		record CommandRecord
		score  float64
	}
	var scored_results []scored

	for rows.Next() {
		var record CommandRecord
		var embStr string
		if err := rows.Scan(&record.Id, &record.Key, &record.Data, &embStr); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		var storedEmb []float64
		if err := json.Unmarshal([]byte(embStr), &storedEmb); err != nil {
			continue
		}
		sim := cosineSimilarity(embedding, storedEmb)
		scored_results = append(scored_results, scored{record: record, score: sim})
	}

	// Sort by similarity descending
	for i := 0; i < len(scored_results); i++ {
		for j := i + 1; j < len(scored_results); j++ {
			if scored_results[j].score > scored_results[i].score {
				scored_results[i], scored_results[j] = scored_results[j], scored_results[i]
			}
		}
	}

	var results []CommandRecord
	for i, s := range scored_results {
		if i >= limit {
			break
		}
		results = append(results, s.record)
	}
	return results, nil
}

// authenticateUserSQLite validates email and API key in SQLite.
func authenticateUserSQLite(email, apiKey string) (bool, error) {
	email = strings.TrimSpace(email)
	apiKey = strings.TrimSpace(apiKey)

	if email == "" || apiKey == "" {
		return false, fmt.Errorf("email and API key are required")
	}
	if db == nil {
		return false, fmt.Errorf("database not connected")
	}

	accessTbl := sqliteAccessTable()
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE email = ? AND api_key = ?", accessTbl)
	err := db.QueryRow(query, email, apiKey).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("authentication query failed: %v", err)
	}
	return count > 0, nil
}

// cosineSimilarity computes cosine similarity between two vectors.
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		minLen := len(a)
		if len(b) < minLen {
			minLen = len(b)
		}
		a = a[:minLen]
		b = b[:minLen]
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func sqliteTableName() string {
	tbl := config.GetEnv("TB_NAME", "data")
	return tbl
}

func sqliteAccessTable() string {
	tbl := config.GetEnv("ACCESS_TB", "access")
	return tbl
}
