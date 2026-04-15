package mcp

// SearchInput defines the input for the search_commands tool.
type SearchInput struct {
	Query string `json:"query" jsonschema:"The search pattern (e.g., 'postgresql backup')"`
	Limit int    `json:"limit,omitempty" jsonschema:"Maximum number of results to return (default 5)"`
}

// AddCommandInput defines the input for the add_command tool.
type AddCommandInput struct {
	Command     string `json:"command" jsonschema:"The actual CLI command string"`
	Description string `json:"description" jsonschema:"What the command does"`
}

// GetCommandInput defines the input for the get_command tool.
type GetCommandInput struct {
	Id int `json:"id" jsonschema:"The ID of the command to retrieve"`
}

// CommandResult represents a command record returned to the MCP client.
type CommandResult struct {
	Id          int    `json:"id"`
	Command     string `json:"command"`
	Description string `json:"description"`
}

// StatsResult represents database statistics.
type StatsResult struct {
	TotalEntries   int `json:"total_entries"`
	WithEmbeddings int `json:"with_embeddings"`
}
