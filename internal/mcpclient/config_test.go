package mcpclient

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseConfig_ValidSingleEntry(t *testing.T) {
	dir := t.TempDir()
	cfg := MCPServerConfig{
		MCPServers: map[string]MCPServerEntry{
			"my-server": {URL: "http://localhost:3001/sse"},
		},
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	path := filepath.Join(dir, "mcp_server.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	name, endpoint, err := ParseConfig(path)
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if name != "my-server" {
		t.Errorf("serverName = %q, want %q", name, "my-server")
	}
	if endpoint != "http://localhost:3001/sse" {
		t.Errorf("endpoint = %q, want %q", endpoint, "http://localhost:3001/sse")
	}
}

func TestParseConfig_FileNotFound(t *testing.T) {
	_, _, err := ParseConfig("/nonexistent/path/mcp_server.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !strings.Contains(err.Error(), "MCP server config not found") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "MCP server config not found")
	}
	if !strings.Contains(err.Error(), "/nonexistent/path/mcp_server.json") {
		t.Errorf("error = %q, want it to contain the file path", err.Error())
	}
}

func TestParseConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mcp_server.json")
	if err := os.WriteFile(path, []byte("not valid json{{{"), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, _, err := ParseConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse MCP server config") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "failed to parse MCP server config")
	}
}

func TestParseConfig_EmptyMCPServersMap(t *testing.T) {
	dir := t.TempDir()
	cfg := MCPServerConfig{
		MCPServers: map[string]MCPServerEntry{},
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	path := filepath.Join(dir, "mcp_server.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, _, err = ParseConfig(path)
	if err == nil {
		t.Fatal("expected error for empty mcpServers map, got nil")
	}
	if !strings.Contains(err.Error(), "no valid MCP server entry found") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "no valid MCP server entry found")
	}
}

func TestParseConfig_MissingURL(t *testing.T) {
	dir := t.TempDir()
	cfg := MCPServerConfig{
		MCPServers: map[string]MCPServerEntry{
			"my-server": {URL: ""},
		},
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	path := filepath.Join(dir, "mcp_server.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, _, err = ParseConfig(path)
	if err == nil {
		t.Fatal("expected error for missing URL, got nil")
	}
	if !strings.Contains(err.Error(), "no valid MCP server entry found") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "no valid MCP server entry found")
	}
}

func TestParseConfig_NilMCPServersMap(t *testing.T) {
	dir := t.TempDir()
	// JSON with no mcpServers field at all.
	path := filepath.Join(dir, "mcp_server.json")
	if err := os.WriteFile(path, []byte(`{}`), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, _, err := ParseConfig(path)
	if err == nil {
		t.Fatal("expected error for nil mcpServers map, got nil")
	}
	if !strings.Contains(err.Error(), "no valid MCP server entry found") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "no valid MCP server entry found")
	}
}
