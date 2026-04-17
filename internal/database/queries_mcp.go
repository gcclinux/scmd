package database

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gcclinux/scmd/internal/config"
)

// ---------------------------------------------------------------------------
// MCP bridge types and functions.
//
// The mcpclient package imports database (for CommandRecord), so database
// cannot import mcpclient without creating a cycle. Instead, queries_mcp.go
// works through bridge function variables that the main package wires to the
// concrete mcpclient.Client methods at startup. Record data crosses the
// boundary as JSON-encoded bytes and is decoded into the local mcpRecord
// type which mirrors mcpclient.MCPRecord.
// ---------------------------------------------------------------------------

// mcpRecord mirrors mcpclient.MCPRecord without importing the package.
type mcpRecord struct {
	ID        string            `json:"id"`
	Key       string            `json:"key"`
	Content   string            `json:"content"`
	Metadata  map[string]string `json:"metadata"`
	Embedding []float64         `json:"embedding,omitempty"`
}

// Bridge function variables — set by the main package at startup.
var (
	// MCPListDataFn lists records. Returns JSON-encoded []mcpRecord.
	MCPListDataFn func(namespace string, limit, offset int) ([]byte, error)

	// MCPStoreDataFn stores a record on the MCP server.
	MCPStoreDataFn func(key, content string, embedding []float64, metadata map[string]string) error

	// MCPGetDataFn retrieves a single record by UUID. Returns JSON-encoded mcpRecord.
	MCPGetDataFn func(uuid string) ([]byte, error)

	// MCPUpdateDataFn updates a record's embedding.
	MCPUpdateDataFn func(uuid string, embedding []float64) error

	// MCPDeleteDataFn deletes a record by UUID.
	MCPDeleteDataFn func(uuid string) (bool, error)

	// MCPQuerySimilarFn performs vector similarity search. Returns JSON-encoded []mcpRecord.
	MCPQuerySimilarFn func(embedding []float64, namespace string, limit int) ([]byte, error)

	// MCPIDMapToUUIDFn resolves an integer ID to a UUID.
	MCPIDMapToUUIDFn func(id int) (string, error)

	// MCPIDMapAssignFn assigns sequential integer IDs to UUIDs.
	MCPIDMapAssignFn func(uuids []string) []int

	// MCPCheckCommandExistsFn checks if any record key matches the command.
	// Accepts JSON-encoded []mcpRecord and the command string.
	MCPCheckCommandExistsFn func(recordsJSON []byte, command string) (bool, error)
)

// toCommandRecord converts an mcpRecord to a CommandRecord, assigning an
// integer ID via the session mapper.
func (r *mcpRecord) toCommandRecord() CommandRecord {
	ids := MCPIDMapAssignFn([]string{r.ID})
	return CommandRecord{
		Id:   ids[0],
		Key:  r.Key,
		Data: r.Content,
	}
}

// parseMCPRecords deserializes JSON-encoded MCP records.
func parseMCPRecords(data []byte) ([]mcpRecord, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var records []mcpRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("error parsing MCP records: %v", err)
	}
	return records, nil
}

// parseSingleMCPRecord deserializes a single JSON-encoded MCP record.
func parseSingleMCPRecord(data []byte) (*mcpRecord, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var record mcpRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("error parsing MCP record: %v", err)
	}
	return &record, nil
}

// ---------------------------------------------------------------------------
// MCP backend query functions
// ---------------------------------------------------------------------------

// searchCommandsMCP searches for commands via the MCP backend with client-side
// keyword filtering. If pattern is empty, all records are returned. Commas in
// the pattern trigger OR logic (match any word); spaces trigger AND logic
// (match all words). Matching is case-insensitive against both Key and Content.
func searchCommandsMCP(pattern string) ([]byte, error) {
	namespace := config.TableName()

	data, err := MCPListDataFn(namespace, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("error listing MCP data: %v", err)
	}

	records, err := parseMCPRecords(data)
	if err != nil {
		return nil, err
	}

	var filtered []mcpRecord

	if pattern == "" {
		filtered = records
	} else if strings.Contains(pattern, ",") {
		// Comma-separated: OR logic — match any word
		parts := strings.Split(pattern, ",")
		var words []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				words = append(words, strings.ToLower(p))
			}
		}
		for _, r := range records {
			key := strings.ToLower(r.Key)
			content := strings.ToLower(r.Content)
			for _, w := range words {
				if strings.Contains(key, w) || strings.Contains(content, w) {
					filtered = append(filtered, r)
					break
				}
			}
		}
	} else {
		// Space-separated: AND logic — match all words
		words := strings.Fields(pattern)
		for _, r := range records {
			key := strings.ToLower(r.Key)
			content := strings.ToLower(r.Content)
			allMatch := true
			for _, w := range words {
				w = strings.ToLower(w)
				if !strings.Contains(key, w) && !strings.Contains(content, w) {
					allMatch = false
					break
				}
			}
			if allMatch {
				filtered = append(filtered, r)
			}
		}
	}

	// Map MCP records to CommandRecords via the ID mapper
	var results []CommandRecord
	for i := range filtered {
		results = append(results, filtered[i].toCommandRecord())
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		return nil, fmt.Errorf("error marshaling to JSON: %v", err)
	}
	return jsonData, nil
}

// addCommandMCP adds a new command via the MCP backend. If embeddingFn is
// provided, an embedding is generated and included in the store call.
func addCommandMCP(command, description string, embeddingFn func(string) ([]float64, error)) (bool, error) {
	var embedding []float64
	if embeddingFn != nil {
		text := command + " " + description
		emb, err := embeddingFn(text)
		if err != nil {
			log.Printf("Warning: embedding generation failed: %v\n", err)
		} else {
			embedding = emb
			log.Println("✓ Generated embedding for new command")
		}
	}

	if len(embedding) == 0 {
		log.Println("⚠ No embedding provider available, saving without vector")
	}

	metadata := map[string]string{"source": "scmd"}
	err := MCPStoreDataFn(command, description, embedding, metadata)
	if err != nil {
		return false, fmt.Errorf("error storing command via MCP: %v", err)
	}

	return true, nil
}

// checkCommandExistsMCP checks if a command with the given key already exists
// in the MCP backend by listing all records and matching on the Key field.
func checkCommandExistsMCP(command string) (bool, error) {
	namespace := config.TableName()

	data, err := MCPListDataFn(namespace, 0, 0)
	if err != nil {
		return false, fmt.Errorf("error listing MCP data: %v", err)
	}

	return MCPCheckCommandExistsFn(data, command)
}

// deleteCommandMCP deletes a command by resolving the integer ID to a UUID
// via the ID mapper and calling DeleteData on the MCP server.
func deleteCommandMCP(id int) (bool, error) {
	uuid, err := MCPIDMapToUUIDFn(id)
	if err != nil {
		return false, err
	}

	return MCPDeleteDataFn(uuid)
}

// getCommandByIDMCP retrieves a single command by resolving the integer ID
// to a UUID and calling GetData on the MCP server.
func getCommandByIDMCP(id int) (*CommandRecord, error) {
	uuid, err := MCPIDMapToUUIDFn(id)
	if err != nil {
		return nil, err
	}

	data, err := MCPGetDataFn(uuid)
	if err != nil {
		return nil, err
	}

	record, err := parseSingleMCPRecord(data)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, fmt.Errorf("no command found with ID %d", id)
	}

	cr := record.toCommandRecord()
	return &cr, nil
}

// getCommandsWithoutEmbeddingsMCP returns all commands that have no embedding
// by listing all records and filtering for those with empty Embedding slices.
func getCommandsWithoutEmbeddingsMCP() ([]CommandRecord, error) {
	namespace := config.TableName()

	data, err := MCPListDataFn(namespace, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("error listing MCP data: %v", err)
	}

	records, err := parseMCPRecords(data)
	if err != nil {
		return nil, err
	}

	var results []CommandRecord
	for i := range records {
		if len(records[i].Embedding) == 0 {
			results = append(results, records[i].toCommandRecord())
		}
	}

	return results, nil
}

// updateEmbeddingMCP updates the embedding for a command by resolving the
// integer ID to a UUID and calling UpdateData on the MCP server.
func updateEmbeddingMCP(id int, embedding []float64) error {
	uuid, err := MCPIDMapToUUIDFn(id)
	if err != nil {
		return err
	}

	return MCPUpdateDataFn(uuid, embedding)
}

// getEmbeddingStatsMCP returns the total number of records and the count of
// records that have embeddings by listing all records from the MCP server.
func getEmbeddingStatsMCP() (total int, withEmbeddings int, err error) {
	namespace := config.TableName()

	data, err := MCPListDataFn(namespace, 0, 0)
	if err != nil {
		return 0, 0, fmt.Errorf("error listing MCP data: %v", err)
	}

	records, err := parseMCPRecords(data)
	if err != nil {
		return 0, 0, err
	}

	total = len(records)
	for _, r := range records {
		if len(r.Embedding) > 0 {
			withEmbeddings++
		}
	}

	return total, withEmbeddings, nil
}

// searchByVectorMCP performs a vector similarity search by calling
// QuerySimilar on the MCP server and mapping the results to CommandRecords.
func searchByVectorMCP(embedding []float64, limit int) ([]CommandRecord, error) {
	namespace := config.TableName()

	data, err := MCPQuerySimilarFn(embedding, namespace, limit)
	if err != nil {
		return nil, fmt.Errorf("vector search error: %v", err)
	}

	records, err := parseMCPRecords(data)
	if err != nil {
		return nil, err
	}

	var results []CommandRecord
	for i := range records {
		results = append(results, records[i].toCommandRecord())
	}

	return results, nil
}
