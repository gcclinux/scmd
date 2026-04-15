# Walkthrough: SCMD MCP Integration

I have successfully integrated a **Model Context Protocol (MCP)** server into `scmd`. This allows AI assistants to directly interact with your personal command knowledge base.

## Changes Made

### 1. New MCP Package
- Created `internal/mcp` which uses the official Go MCP SDK.
- Implemented three core tools:
    - **`search_commands`**: Hybrid search (keyword + vector).
    - **`add_command`**: Save new commands to the database.
    - **`get_stats`**: Overview of entries and embedding coverage.

### 2. CLI Integration
- Added the `--mcp` flag to the main command line interface.
- Updated the help menu to include the new capability.

### 3. Build & Dependency Update
- Upgraded project to Go 1.25.0 to support the MCP SDK.
- Added `github.com/modelcontextprotocol/go-sdk` dependency.

---

## How to use the MCP Server

### 1. Build the Binary
To register it as an MCP server, you first need a compiled binary:
```powershell
go build -o scmd.exe ./cmd/scmd/
```

### 2. Register in your Assistant (Claude Desktop / Cursor)
Add the following to your `mcp_config.json`:

```json
{
  "mcpServers": {
    "scmd": {
      "command": "C:/Users/ricardo/Programming/scmd/scmd.exe",
      "args": ["--mcp"]
    }
  }
}
```

> [!IMPORTANT]
> The MCP server uses the same `~/.scmd/config.json` as your CLI. Ensure your database connection and AI provider settings are correct there.

---

## Technical Details

- **Transport**: Communicates via standard input/output (stdio), the standard for local MCP servers.
- **Safety**: The server runs with the same permissions as your current user.
- **AI Integration**: When using `add_command` via MCP, it will automatically attempt to generate embeddings if you have an AI provider (Ollama or Gemini) configured.

> [!TIP]
> You can now ask an AI assistant: *"Search my scmd database for docker network commands"* or *"Save this command to scmd: docker system prune -a"* and it will work!
