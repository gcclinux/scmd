package database

import (
	"encoding/json"
	"os"
	"testing"
)

// saveBridgeFns saves the current bridge function variables and returns a
// restore function. This lets each test install mock bridges without leaking
// state to other tests.
func saveBridgeFns() func() {
	origList := MCPListDataFn
	origStore := MCPStoreDataFn
	origGet := MCPGetDataFn
	origUpdate := MCPUpdateDataFn
	origDelete := MCPDeleteDataFn
	origQuery := MCPQuerySimilarFn
	origToUUID := MCPIDMapToUUIDFn
	origAssign := MCPIDMapAssignFn
	origCheck := MCPCheckCommandExistsFn
	return func() {
		MCPListDataFn = origList
		MCPStoreDataFn = origStore
		MCPGetDataFn = origGet
		MCPUpdateDataFn = origUpdate
		MCPDeleteDataFn = origDelete
		MCPQuerySimilarFn = origQuery
		MCPIDMapToUUIDFn = origToUUID
		MCPIDMapAssignFn = origAssign
		MCPCheckCommandExistsFn = origCheck
	}
}

// setDBType sets the DB_TYPE env var and returns a cleanup function.
func setDBType(t *testing.T, value string) {
	t.Helper()
	old, hadOld := os.LookupEnv("DB_TYPE")
	os.Setenv("DB_TYPE", value)
	t.Cleanup(func() {
		if hadOld {
			os.Setenv("DB_TYPE", old)
		} else {
			os.Unsetenv("DB_TYPE")
		}
	})
}

// --- IsMCP tests ---

func TestIsMCP_TrueWhenMCP(t *testing.T) {
	setDBType(t, "mcp")
	if !IsMCP() {
		t.Error("IsMCP() = false, want true when DB_TYPE=mcp")
	}
}

func TestIsMCP_TrueWhenMCPUpperCase(t *testing.T) {
	setDBType(t, "MCP")
	if !IsMCP() {
		t.Error("IsMCP() = false, want true when DB_TYPE=MCP (case-insensitive)")
	}
}

func TestIsMCP_FalseWhenSQLite(t *testing.T) {
	setDBType(t, "sqlite")
	if IsMCP() {
		t.Error("IsMCP() = true, want false when DB_TYPE=sqlite")
	}
}

func TestIsMCP_FalseWhenPostgreSQL(t *testing.T) {
	setDBType(t, "postgresql")
	if IsMCP() {
		t.Error("IsMCP() = true, want false when DB_TYPE=postgresql")
	}
}

func TestIsMCP_FalseWhenEmpty(t *testing.T) {
	setDBType(t, "")
	if IsMCP() {
		t.Error("IsMCP() = true, want false when DB_TYPE is empty")
	}
}



// --- Dispatch tests: verify each public function routes to MCP bridge ---

// installMockBridges sets up mock bridge functions that record whether they
// were called. It also sets TB_NAME for namespace resolution.
func installMockBridges(t *testing.T) map[string]*bool {
	t.Helper()

	// Set TB_NAME so config.TableName() returns a value.
	old, had := os.LookupEnv("TB_NAME")
	os.Setenv("TB_NAME", "testns")
	t.Cleanup(func() {
		if had {
			os.Setenv("TB_NAME", old)
		} else {
			os.Unsetenv("TB_NAME")
		}
	})

	called := map[string]*bool{
		"list":     new(bool),
		"store":    new(bool),
		"get":      new(bool),
		"update":   new(bool),
		"delete":   new(bool),
		"query":    new(bool),
		"toUUID":   new(bool),
		"assign":   new(bool),
		"checkCmd": new(bool),
	}

	// Return empty JSON arrays / objects for functions that return data.
	emptyArray, _ := json.Marshal([]mcpRecord{})
	singleRecord, _ := json.Marshal(mcpRecord{
		ID: "uuid-1", Key: "ls", Content: "list files",
	})

	MCPListDataFn = func(namespace string, limit, offset int) ([]byte, error) {
		*called["list"] = true
		return emptyArray, nil
	}
	MCPStoreDataFn = func(key, content string, embedding []float64, metadata map[string]string) error {
		*called["store"] = true
		return nil
	}
	MCPGetDataFn = func(uuid string) ([]byte, error) {
		*called["get"] = true
		return singleRecord, nil
	}
	MCPUpdateDataFn = func(uuid string, embedding []float64) error {
		*called["update"] = true
		return nil
	}
	MCPDeleteDataFn = func(uuid string) (bool, error) {
		*called["delete"] = true
		return true, nil
	}
	MCPQuerySimilarFn = func(embedding []float64, namespace string, limit int) ([]byte, error) {
		*called["query"] = true
		return emptyArray, nil
	}
	MCPIDMapToUUIDFn = func(id int) (string, error) {
		*called["toUUID"] = true
		return "uuid-1", nil
	}
	MCPIDMapAssignFn = func(uuids []string) []int {
		*called["assign"] = true
		ids := make([]int, len(uuids))
		for i := range uuids {
			ids[i] = i + 1
		}
		return ids
	}
	MCPCheckCommandExistsFn = func(recordsJSON []byte, command string) (bool, error) {
		*called["checkCmd"] = true
		return false, nil
	}

	return called
}

func TestSearchCommands_DispatchesToMCP(t *testing.T) {
	setDBType(t, "mcp")
	restore := saveBridgeFns()
	defer restore()
	called := installMockBridges(t)

	_, err := SearchCommands("docker", "json")
	if err != nil {
		t.Fatalf("SearchCommands error: %v", err)
	}
	if !*called["list"] {
		t.Error("SearchCommands did not dispatch to MCP list bridge")
	}
}

func TestAddCommand_DispatchesToMCP(t *testing.T) {
	setDBType(t, "mcp")
	restore := saveBridgeFns()
	defer restore()
	called := installMockBridges(t)

	_, err := AddCommand("docker ps", "list containers", nil)
	if err != nil {
		t.Fatalf("AddCommand error: %v", err)
	}
	if !*called["store"] {
		t.Error("AddCommand did not dispatch to MCP store bridge")
	}
}

func TestCheckCommandExists_DispatchesToMCP(t *testing.T) {
	setDBType(t, "mcp")
	restore := saveBridgeFns()
	defer restore()
	called := installMockBridges(t)

	_, err := CheckCommandExists("docker ps")
	if err != nil {
		t.Fatalf("CheckCommandExists error: %v", err)
	}
	if !*called["list"] {
		t.Error("CheckCommandExists did not dispatch to MCP list bridge")
	}
	if !*called["checkCmd"] {
		t.Error("CheckCommandExists did not dispatch to MCP checkCmd bridge")
	}
}

func TestDeleteCommand_DispatchesToMCP(t *testing.T) {
	setDBType(t, "mcp")
	restore := saveBridgeFns()
	defer restore()
	called := installMockBridges(t)

	_, err := DeleteCommand(1)
	if err != nil {
		t.Fatalf("DeleteCommand error: %v", err)
	}
	if !*called["toUUID"] {
		t.Error("DeleteCommand did not resolve integer ID via MCP toUUID bridge")
	}
	if !*called["delete"] {
		t.Error("DeleteCommand did not dispatch to MCP delete bridge")
	}
}

func TestGetCommandByID_DispatchesToMCP(t *testing.T) {
	setDBType(t, "mcp")
	restore := saveBridgeFns()
	defer restore()
	called := installMockBridges(t)

	_, err := GetCommandByID(1)
	if err != nil {
		t.Fatalf("GetCommandByID error: %v", err)
	}
	if !*called["toUUID"] {
		t.Error("GetCommandByID did not resolve integer ID via MCP toUUID bridge")
	}
	if !*called["get"] {
		t.Error("GetCommandByID did not dispatch to MCP get bridge")
	}
}

func TestGetCommandsWithoutEmbeddings_DispatchesToMCP(t *testing.T) {
	setDBType(t, "mcp")
	restore := saveBridgeFns()
	defer restore()
	called := installMockBridges(t)

	_, err := GetCommandsWithoutEmbeddings()
	if err != nil {
		t.Fatalf("GetCommandsWithoutEmbeddings error: %v", err)
	}
	if !*called["list"] {
		t.Error("GetCommandsWithoutEmbeddings did not dispatch to MCP list bridge")
	}
}

func TestUpdateEmbedding_DispatchesToMCP(t *testing.T) {
	setDBType(t, "mcp")
	restore := saveBridgeFns()
	defer restore()
	called := installMockBridges(t)

	err := UpdateEmbedding(1, []float64{0.1, 0.2})
	if err != nil {
		t.Fatalf("UpdateEmbedding error: %v", err)
	}
	if !*called["toUUID"] {
		t.Error("UpdateEmbedding did not resolve integer ID via MCP toUUID bridge")
	}
	if !*called["update"] {
		t.Error("UpdateEmbedding did not dispatch to MCP update bridge")
	}
}

func TestGetEmbeddingStats_DispatchesToMCP(t *testing.T) {
	setDBType(t, "mcp")
	restore := saveBridgeFns()
	defer restore()
	called := installMockBridges(t)

	_, _, err := GetEmbeddingStats()
	if err != nil {
		t.Fatalf("GetEmbeddingStats error: %v", err)
	}
	if !*called["list"] {
		t.Error("GetEmbeddingStats did not dispatch to MCP list bridge")
	}
}

func TestSearchByVector_DispatchesToMCP(t *testing.T) {
	setDBType(t, "mcp")
	restore := saveBridgeFns()
	defer restore()
	called := installMockBridges(t)

	_, err := SearchByVector([]float64{0.1, 0.2}, 10)
	if err != nil {
		t.Fatalf("SearchByVector error: %v", err)
	}
	if !*called["query"] {
		t.Error("SearchByVector did not dispatch to MCP query bridge")
	}
}

func TestAuthenticateUser_ReturnsErrorWhenMCP(t *testing.T) {
	setDBType(t, "mcp")

	ok, err := AuthenticateUser("user@example.com", "key123")
	if ok {
		t.Error("AuthenticateUser returned true, want false for MCP backend")
	}
	if err == nil {
		t.Fatal("AuthenticateUser returned nil error, want error for MCP backend")
	}
	if err.Error() != "authentication not supported with MCP backend" {
		t.Errorf("unexpected error message: %v", err)
	}
}
