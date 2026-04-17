package mcpclient

import (
	"time"

	"github.com/gcclinux/scmd/internal/database"
)

// MCPRecord represents a record as returned by the MCP server.
// Field types match the go-mcp-postgres-server's DocumentSummary / Document models.
type MCPRecord struct {
	ID        string         `json:"id"`
	Namespace string         `json:"namespace,omitempty"`
	Key       string         `json:"key"`
	Content   string         `json:"content"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Embedding []float64      `json:"embedding,omitempty"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
	UpdatedAt time.Time      `json:"updated_at,omitempty"`
}

// MCPDocumentResult extends MCPRecord with a similarity score,
// matching the server's DocumentResult from query_similar.
type MCPDocumentResult struct {
	MCPRecord
	Similarity float64 `json:"similarity,omitempty"`
}

// ListResponse matches the server's list_data response wrapper.
type ListResponse struct {
	Records    []MCPRecord `json:"records"`
	TotalCount int64       `json:"total_count"`
}

// ToCommandRecord converts an MCPRecord to a database.CommandRecord using the
// ID mapper. The UUID is mapped to a sequential integer ID via the mapper,
// Key maps to Key, and Content maps to Data.
func (r *MCPRecord) ToCommandRecord(mapper *IDMapper) database.CommandRecord {
	ids := mapper.Assign([]string{r.ID})
	return database.CommandRecord{
		Id:   ids[0],
		Key:  r.Key,
		Data: r.Content,
	}
}
