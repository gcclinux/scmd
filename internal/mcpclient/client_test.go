package mcpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// toolCall captures the arguments passed to a mock toolCaller.
type toolCall struct {
	toolName string
	args     map[string]any
}

// newTestClient creates a Client with an injectable toolCaller for testing.
// The namespace is set to "test-ns" by default.
func newTestClient(fn ToolCallerFunc) *Client {
	return &Client{
		namespace:  "test-ns",
		serverName: "test-server",
		IDMap:      NewIDMapper(),
		toolCaller: fn,
	}
}

// --- StoreData parameter construction tests ---

func TestStoreData_ConstructsCorrectParameters(t *testing.T) {
	var captured toolCall
	c := newTestClient(func(ctx context.Context, toolName string, args map[string]any) (string, error) {
		captured = toolCall{toolName: toolName, args: args}
		return "", nil
	})

	embedding := []float64{0.1, 0.2, 0.3}
	metadata := map[string]string{"source": "scmd"}

	err := c.StoreData("docker ps", "list containers", embedding, metadata)
	if err != nil {
		t.Fatalf("StoreData: %v", err)
	}

	if captured.toolName != "store_data" {
		t.Errorf("toolName = %q, want %q", captured.toolName, "store_data")
	}
	if captured.args["key"] != "docker ps" {
		t.Errorf("key = %v, want %q", captured.args["key"], "docker ps")
	}
	if captured.args["content"] != "list containers" {
		t.Errorf("content = %v, want %q", captured.args["content"], "list containers")
	}
	if captured.args["namespace"] != "test-ns" {
		t.Errorf("namespace = %v, want %q", captured.args["namespace"], "test-ns")
	}

	// Verify embedding is passed through.
	emb, ok := captured.args["embedding"].([]float64)
	if !ok {
		t.Fatalf("embedding type = %T, want []float64", captured.args["embedding"])
	}
	if len(emb) != 3 || emb[0] != 0.1 || emb[1] != 0.2 || emb[2] != 0.3 {
		t.Errorf("embedding = %v, want [0.1 0.2 0.3]", emb)
	}

	// Verify metadata is passed through (converted to map[string]any for MCP server).
	meta, ok := captured.args["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("metadata type = %T, want map[string]any", captured.args["metadata"])
	}
	if meta["source"] != "scmd" {
		t.Errorf("metadata[source] = %v, want %q", meta["source"], "scmd")
	}
}

func TestStoreData_NilMetadataDefaultsToSource(t *testing.T) {
	var captured toolCall
	c := newTestClient(func(ctx context.Context, toolName string, args map[string]any) (string, error) {
		captured = toolCall{toolName: toolName, args: args}
		return "", nil
	})

	err := c.StoreData("ls -la", "list files", []float64{1.0}, nil)
	if err != nil {
		t.Fatalf("StoreData: %v", err)
	}

	meta, ok := captured.args["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("metadata type = %T, want map[string]any", captured.args["metadata"])
	}
	if meta["source"] != "scmd" {
		t.Errorf("metadata[source] = %v, want %q", meta["source"], "scmd")
	}
}

func TestStoreData_EmptyEmbeddingOmitted(t *testing.T) {
	var captured toolCall
	c := newTestClient(func(ctx context.Context, toolName string, args map[string]any) (string, error) {
		captured = toolCall{toolName: toolName, args: args}
		return "", nil
	})

	err := c.StoreData("pwd", "print working directory", nil, nil)
	if err != nil {
		t.Fatalf("StoreData: %v", err)
	}

	if _, exists := captured.args["embedding"]; exists {
		t.Error("embedding should be omitted when nil/empty, but was present")
	}
}

// --- QuerySimilar tests ---

func TestQuerySimilar_ReturnsEmptySliceOnZeroResults(t *testing.T) {
	c := newTestClient(func(ctx context.Context, toolName string, args map[string]any) (string, error) {
		return "", nil // empty response
	})

	records, err := c.QuerySimilar([]float64{0.1, 0.2}, "test-ns", 10)
	if err != nil {
		t.Fatalf("QuerySimilar: %v", err)
	}
	if records == nil {
		t.Fatal("QuerySimilar returned nil, want empty slice")
	}
	if len(records) != 0 {
		t.Errorf("len(records) = %d, want 0", len(records))
	}
}

func TestQuerySimilar_ParsesValidResponse(t *testing.T) {
	// Server returns []DocumentResult (MCPRecord fields + similarity).
	expected := []MCPDocumentResult{
		{MCPRecord: MCPRecord{ID: "uuid-1", Key: "docker ps", Content: "list containers"}, Similarity: 0.95},
		{MCPRecord: MCPRecord{ID: "uuid-2", Key: "ls -la", Content: "list files"}, Similarity: 0.80},
	}
	respJSON, _ := json.Marshal(expected)

	c := newTestClient(func(ctx context.Context, toolName string, args map[string]any) (string, error) {
		return string(respJSON), nil
	})

	records, err := c.QuerySimilar([]float64{0.1}, "test-ns", 5)
	if err != nil {
		t.Fatalf("QuerySimilar: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("len(records) = %d, want 2", len(records))
	}
	if records[0].Key != "docker ps" {
		t.Errorf("records[0].Key = %q, want %q", records[0].Key, "docker ps")
	}
	if records[1].Key != "ls -la" {
		t.Errorf("records[1].Key = %q, want %q", records[1].Key, "ls -la")
	}
}

func TestQuerySimilar_PassesCorrectParameters(t *testing.T) {
	var captured toolCall
	c := newTestClient(func(ctx context.Context, toolName string, args map[string]any) (string, error) {
		captured = toolCall{toolName: toolName, args: args}
		return "[]", nil // empty DocumentResult array
	})

	embedding := []float64{0.5, 0.6, 0.7}
	_, err := c.QuerySimilar(embedding, "my-namespace", 25)
	if err != nil {
		t.Fatalf("QuerySimilar: %v", err)
	}

	if captured.toolName != "query_similar" {
		t.Errorf("toolName = %q, want %q", captured.toolName, "query_similar")
	}
	if captured.args["namespace"] != "my-namespace" {
		t.Errorf("namespace = %v, want %q", captured.args["namespace"], "my-namespace")
	}
	if captured.args["limit"] != 25 {
		t.Errorf("limit = %v, want 25", captured.args["limit"])
	}
}

// --- ListData parameter tests ---

func TestListData_PassesLimitAndOffsetCorrectly(t *testing.T) {
	var captured toolCall
	c := newTestClient(func(ctx context.Context, toolName string, args map[string]any) (string, error) {
		captured = toolCall{toolName: toolName, args: args}
		return `{"records":[],"total_count":0}`, nil
	})

	_, err := c.ListData("data", 50, 100)
	if err != nil {
		t.Fatalf("ListData: %v", err)
	}

	if captured.toolName != "list_data" {
		t.Errorf("toolName = %q, want %q", captured.toolName, "list_data")
	}
	if captured.args["namespace"] != "data" {
		t.Errorf("namespace = %v, want %q", captured.args["namespace"], "data")
	}
	if captured.args["limit"] != 50 {
		t.Errorf("limit = %v, want 50", captured.args["limit"])
	}
	if captured.args["offset"] != 100 {
		t.Errorf("offset = %v, want 100", captured.args["offset"])
	}
}

func TestListData_ZeroLimitAndOffset(t *testing.T) {
	var captured toolCall
	c := newTestClient(func(ctx context.Context, toolName string, args map[string]any) (string, error) {
		captured = toolCall{toolName: toolName, args: args}
		return `{"records":[],"total_count":0}`, nil
	})

	_, err := c.ListData("ns", 0, 0)
	if err != nil {
		t.Fatalf("ListData: %v", err)
	}

	if captured.args["limit"] != 0 {
		t.Errorf("limit = %v, want 0", captured.args["limit"])
	}
	if captured.args["offset"] != 0 {
		t.Errorf("offset = %v, want 0", captured.args["offset"])
	}
}

func TestListData_ReturnsEmptySliceOnEmptyResponse(t *testing.T) {
	c := newTestClient(func(ctx context.Context, toolName string, args map[string]any) (string, error) {
		return "", nil
	})

	records, err := c.ListData("ns", 10, 0)
	if err != nil {
		t.Fatalf("ListData: %v", err)
	}
	if records == nil {
		t.Fatal("ListData returned nil, want empty slice")
	}
	if len(records) != 0 {
		t.Errorf("len(records) = %d, want 0", len(records))
	}
}

// --- Error propagation tests ---

func TestStoreData_PropagatesError(t *testing.T) {
	c := newTestClient(func(ctx context.Context, toolName string, args map[string]any) (string, error) {
		return "", fmt.Errorf("store_data failed: connection refused")
	})

	err := c.StoreData("cmd", "desc", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "connection refused")
	}
}

func TestQuerySimilar_PropagatesError(t *testing.T) {
	c := newTestClient(func(ctx context.Context, toolName string, args map[string]any) (string, error) {
		return "", fmt.Errorf("query_similar failed: timeout")
	})

	_, err := c.QuerySimilar([]float64{0.1}, "ns", 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "timeout")
	}
}

func TestListData_PropagatesError(t *testing.T) {
	c := newTestClient(func(ctx context.Context, toolName string, args map[string]any) (string, error) {
		return "", fmt.Errorf("list_data failed: server error")
	})

	_, err := c.ListData("ns", 10, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "server error") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "server error")
	}
}

func TestGetData_PropagatesError(t *testing.T) {
	c := newTestClient(func(ctx context.Context, toolName string, args map[string]any) (string, error) {
		return "", fmt.Errorf("get_data failed: not found")
	})

	_, err := c.GetData("some-uuid")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "not found")
	}
}

func TestDeleteData_PropagatesError(t *testing.T) {
	c := newTestClient(func(ctx context.Context, toolName string, args map[string]any) (string, error) {
		return "", fmt.Errorf("delete_data failed: permission denied")
	})

	_, err := c.DeleteData("some-uuid")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "permission denied")
	}
}

func TestUpdateData_PropagatesError(t *testing.T) {
	c := newTestClient(func(ctx context.Context, toolName string, args map[string]any) (string, error) {
		return "", fmt.Errorf("update_data failed: invalid embedding")
	})

	err := c.UpdateData("some-uuid", []float64{0.1})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid embedding") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "invalid embedding")
	}
}

// --- CheckCommandExists tests ---

func TestCheckCommandExists_MatchFound(t *testing.T) {
	records := []MCPRecord{
		{ID: "uuid-1", Key: "docker ps", Content: "list containers"},
		{ID: "uuid-2", Key: "ls -la", Content: "list files"},
	}

	if !CheckCommandExists(records, "docker ps") {
		t.Error("CheckCommandExists(docker ps) = false, want true")
	}
}

func TestCheckCommandExists_NoMatch(t *testing.T) {
	records := []MCPRecord{
		{ID: "uuid-1", Key: "docker ps", Content: "list containers"},
	}

	if CheckCommandExists(records, "docker run") {
		t.Error("CheckCommandExists(docker run) = true, want false")
	}
}

func TestCheckCommandExists_EmptyRecords(t *testing.T) {
	if CheckCommandExists([]MCPRecord{}, "anything") {
		t.Error("CheckCommandExists on empty slice = true, want false")
	}
}

func TestCheckCommandExists_PartialMatchNotAccepted(t *testing.T) {
	records := []MCPRecord{
		{ID: "uuid-1", Key: "docker ps -a", Content: "list all containers"},
	}

	if CheckCommandExists(records, "docker ps") {
		t.Error("CheckCommandExists should not match partial key")
	}
}
