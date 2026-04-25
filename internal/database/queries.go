package database

import (
	"fmt"
)

// SearchCommands searches for commands matching the pattern.
func SearchCommands(pattern string, format string) ([]byte, error) {
	if IsMCP() {
		return searchCommandsMCP(pattern)
	}
	return searchCommandsSQLite(pattern)
}

// AddCommand adds a new command to the database.
// embeddingFn is an optional callback to generate embeddings.
func AddCommand(command, description string, embeddingFn func(string) ([]float64, error)) (bool, error) {
	if IsMCP() {
		return addCommandMCP(command, description, embeddingFn)
	}
	return addCommandSQLite(command, description, embeddingFn)
}

// CheckCommandExists checks if a command already exists in the database.
func CheckCommandExists(command string) (bool, error) {
	if IsMCP() {
		return checkCommandExistsMCP(command)
	}
	return checkCommandExistsSQLite(command)
}

// DeleteCommand deletes a command from the database by ID.
func DeleteCommand(id int) (bool, error) {
	if IsMCP() {
		return deleteCommandMCP(id)
	}
	return deleteCommandSQLite(id)
}

// GetCommandByID retrieves a single command record by its ID.
func GetCommandByID(id int) (*CommandRecord, error) {
	if IsMCP() {
		return getCommandByIDMCP(id)
	}
	return getCommandByIDSQLite(id)
}

// GetCommandsWithoutEmbeddings returns all commands that have no embedding.
func GetCommandsWithoutEmbeddings() ([]CommandRecord, error) {
	if IsMCP() {
		return getCommandsWithoutEmbeddingsMCP()
	}
	return getCommandsWithoutEmbeddingsSQLite()
}

// UpdateEmbedding updates the embedding for a command by ID.
func UpdateEmbedding(id int, embedding []float64) error {
	if IsMCP() {
		return updateEmbeddingMCP(id, embedding)
	}
	return updateEmbeddingSQLite(id, embedding)
}

// GetEmbeddingStats returns total commands and count with embeddings.
func GetEmbeddingStats() (total int, withEmbeddings int, err error) {
	if IsMCP() {
		return getEmbeddingStatsMCP()
	}
	return getEmbeddingStatsSQLite()
}

// SearchByVector performs a vector similarity search.
func SearchByVector(embedding []float64, limit int) ([]CommandRecord, error) {
	if IsMCP() {
		return searchByVectorMCP(embedding, limit)
	}
	return searchByVectorSQLite(embedding, limit)
}

// AuthenticateUser validates email and API key against the database.
func AuthenticateUser(email, apiKey string) (bool, error) {
	if IsMCP() {
		return false, fmt.Errorf("authentication not supported with MCP backend")
	}
	return authenticateUserSQLite(email, apiKey)
}

// FormatEmbedding converts a float64 slice to a string representation.
// This is used by some backends or for logging.
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
