package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/gcclinux/scmd/internal/ai"
	"github.com/gcclinux/scmd/internal/database"
)

// StartServer starts the MCP server over stdio.
func StartServer() error {
	// Initialize database first
	if err := database.InitDB(); err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	// Initialize AI providers for embeddings
	ai.InitProviders()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "scmd-mcp",
		Version: "1.0.0",
	}, nil)

	// Search Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_commands",
		Description: "Search for commands in the SCMD database. Supports keyword search.",
	}, handleSearch)

	// Add Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_command",
		Description: "Add a new command to the SCMD database.",
	}, handleAdd)

	// Stats Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_stats",
		Description: "Get database and embedding statistics.",
	}, handleStats)

	return server.Run(context.Background(), &mcp.StdioTransport{})
}

func handleSearch(ctx context.Context, req *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, any, error) {
	received, err := database.SearchCommands(input.Query, "json")
	if err != nil {
		return nil, nil, fmt.Errorf("search error: %v", err)
	}

	var records []database.CommandRecord
	if err := json.Unmarshal(received, &records); err != nil {
		return nil, nil, fmt.Errorf("parse error: %v", err)
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 5
	}
	if len(records) > limit {
		records = records[:limit]
	}

	results := make([]CommandResult, len(records))
	for i, r := range records {
		results[i] = CommandResult{
			Id:          r.Id,
			Command:     r.Key,
			Description: r.Data,
		}
	}

	return nil, results, nil
}

func handleAdd(ctx context.Context, req *mcp.CallToolRequest, input AddCommandInput) (*mcp.CallToolResult, any, error) {
	success, err := database.AddCommand(input.Command, input.Description, ai.GetBestEmbedding)
	if err != nil {
		return nil, nil, fmt.Errorf("add error: %v", err)
	}

	if !success {
		return nil, "Failed to add command", nil
	}

	return nil, fmt.Sprintf("✓ Successfully added command: %s", input.Command), nil
}

func handleStats(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	total, withEmb, err := database.GetEmbeddingStats()
	if err != nil {
		return nil, nil, fmt.Errorf("stats error: %v", err)
	}

	return nil, StatsResult{
		TotalEntries:   total,
		WithEmbeddings: withEmb,
	}, nil
}
