package mcpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolCallerFunc is the signature for the function that invokes an MCP tool.
// It is used internally to allow test injection.
type ToolCallerFunc func(ctx context.Context, toolName string, args map[string]any) (string, error)

// Client wraps the MCP SDK client and provides tool invocation methods
// for interacting with an external MCP server.
type Client struct {
	mcpClient  *mcp.Client
	session    *mcp.ClientSession
	serverName string
	namespace  string
	IDMap      *IDMapper
	toolCaller ToolCallerFunc // injectable for testing; defaults to callToolReal
}

// Init connects to the MCP server described in configPath and returns a ready Client.
func Init(configPath string) (*Client, error) {
	serverName, endpoint, err := ParseConfig(configPath)
	if err != nil {
		return nil, err
	}

	namespace := os.Getenv("TB_NAME")
	if namespace == "" {
		namespace = "scmd"
	}

	mcpClient := mcp.NewClient(&mcp.Implementation{
		Name:    "scmd-mcp-client",
		Version: "1.0.0",
	}, nil)

	transport := &mcp.SSEClientTransport{
		Endpoint: endpoint,
	}

	session, err := mcpClient.Connect(context.Background(), transport, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to MCP server at %s: %v", endpoint, err)
	}

	c := &Client{
		mcpClient:  mcpClient,
		session:    session,
		serverName: serverName,
		namespace:  namespace,
		IDMap:      NewIDMapper(),
	}
	c.toolCaller = c.callToolReal
	return c, nil
}

// Close terminates the connection to the MCP server.
func (c *Client) Close() {
	if c.session != nil {
		c.session.Close()
	}
}

// callTool delegates to the toolCaller function, which defaults to callToolReal
// but can be overridden in tests.
func (c *Client) callTool(ctx context.Context, toolName string, args map[string]any) (string, error) {
	return c.toolCaller(ctx, toolName, args)
}

// callToolReal is the real implementation that invokes a named tool via the MCP session
// and returns the text content from the result.
func (c *Client) callToolReal(ctx context.Context, toolName string, args map[string]any) (string, error) {
	result, err := c.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	})
	if err != nil {
		return "", fmt.Errorf("%s failed: %v", toolName, err)
	}

	if result.IsError {
		// Extract error text from content
		for _, content := range result.Content {
			if tc, ok := content.(*mcp.TextContent); ok {
				return "", fmt.Errorf("%s error: %s", toolName, tc.Text)
			}
		}
		return "", fmt.Errorf("%s returned an error", toolName)
	}

	// Extract text content from the result
	for _, content := range result.Content {
		if tc, ok := content.(*mcp.TextContent); ok {
			return tc.Text, nil
		}
	}

	return "", nil
}

// StoreData invokes the store_data tool on the MCP server.
func (c *Client) StoreData(key, content string, embedding []float64, metadata map[string]string) error {
	args := map[string]any{
		"key":       key,
		"content":   content,
		"namespace": c.namespace,
	}
	if len(embedding) > 0 {
		args["embedding"] = embedding
	}
	// Convert map[string]string to map[string]any for the MCP server.
	metaAny := make(map[string]any)
	if metadata == nil {
		metaAny["source"] = "scmd"
	} else {
		for k, v := range metadata {
			metaAny[k] = v
		}
	}
	args["metadata"] = metaAny

	_, err := c.callTool(context.Background(), "store_data", args)
	return err
}

// QuerySimilar invokes the query_similar tool for vector similarity search.
// The server returns []DocumentResult which embeds DocumentSummary + similarity.
func (c *Client) QuerySimilar(embedding []float64, namespace string, limit int) ([]MCPRecord, error) {
	args := map[string]any{
		"embedding": embedding,
		"namespace": namespace,
		"limit":     limit,
	}

	text, err := c.callTool(context.Background(), "query_similar", args)
	if err != nil {
		return nil, err
	}

	if text == "" {
		return []MCPRecord{}, nil
	}

	// Server returns []DocumentResult (MCPRecord fields + similarity).
	var results []MCPDocumentResult
	if err := json.Unmarshal([]byte(text), &results); err != nil {
		return nil, fmt.Errorf("query_similar: failed to parse response: %v", err)
	}

	records := make([]MCPRecord, len(results))
	for i := range results {
		records[i] = results[i].MCPRecord
	}

	return records, nil
}

// GetData invokes the get_data tool to retrieve a single record by UUID.
func (c *Client) GetData(uuid string) (*MCPRecord, error) {
	args := map[string]any{
		"id": uuid,
	}

	text, err := c.callTool(context.Background(), "get_data", args)
	if err != nil {
		return nil, err
	}

	if text == "" {
		return nil, fmt.Errorf("no command found with ID %s", uuid)
	}

	var record MCPRecord
	if err := json.Unmarshal([]byte(text), &record); err != nil {
		return nil, fmt.Errorf("get_data: failed to parse response: %v", err)
	}

	return &record, nil
}

// ListData invokes the list_data tool with optional pagination.
// The server returns {"records": [...], "total_count": N}.
func (c *Client) ListData(namespace string, limit, offset int) ([]MCPRecord, error) {
	args := map[string]any{
		"namespace": namespace,
		"limit":     limit,
		"offset":    offset,
	}

	text, err := c.callTool(context.Background(), "list_data", args)
	if err != nil {
		return nil, err
	}

	if text == "" {
		return []MCPRecord{}, nil
	}

	// The server wraps the array in {"records": [...], "total_count": N}.
	var resp ListResponse
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		// Fallback: try bare array for forward compatibility.
		var records []MCPRecord
		if err2 := json.Unmarshal([]byte(text), &records); err2 != nil {
			return nil, fmt.Errorf("list_data: failed to parse response: %v", err)
		}
		return records, nil
	}

	return resp.Records, nil
}

// UpdateData invokes the update_data tool to update a record's embedding.
func (c *Client) UpdateData(uuid string, embedding []float64) error {
	args := map[string]any{
		"id":        uuid,
		"embedding": embedding,
	}

	_, err := c.callTool(context.Background(), "update_data", args)
	return err
}

// DeleteData invokes the delete_data tool to remove a record by UUID.
func (c *Client) DeleteData(uuid string) (bool, error) {
	args := map[string]any{
		"id": uuid,
	}

	text, err := c.callTool(context.Background(), "delete_data", args)
	if err != nil {
		return false, err
	}

	// Parse the response to determine success
	var result struct {
		Deleted bool `json:"deleted"`
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		// If we can't parse, assume success if no error was returned
		return true, nil
	}

	return result.Deleted, nil
}

// CheckCommandExists checks if any record in the slice has a Key that
// exactly matches the given command string.
func CheckCommandExists(records []MCPRecord, command string) bool {
	for _, r := range records {
		if r.Key == command {
			return true
		}
	}
	return false
}
