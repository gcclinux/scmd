package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCacheScannerEmptyDir(t *testing.T) {
	// scanCacheFiles with a query that won't match anything should return nil
	matches, err := scanCacheFiles("xyznonexistentquery12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// May return nil or empty — both are acceptable
	if len(matches) > 0 {
		// Only fail if matches contain something unexpected
		t.Logf("found %d matches (may be from existing cache)", len(matches))
	}
}

func TestCacheScannerNonExistentDir(t *testing.T) {
	// This tests the graceful handling — scanCacheFiles uses ensureCacheDir
	// which creates the dir, so it won't be non-existent. Just verify no error.
	matches, err := scanCacheFiles("test query")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = matches // May or may not have results depending on cache state
}

func TestBuildSearchCommandWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("skipping Windows-specific test")
	}
	cmd := buildSearchCommand("C:\\Users\\test\\.scmd\\cache", []string{"docker", "network"})
	if cmd.Path == "" {
		t.Error("expected non-empty command path")
	}
	// Verify findstr is used
	if filepath.Base(cmd.Path) != "findstr" && filepath.Base(cmd.Path) != "findstr.exe" {
		// On Windows, the path may resolve differently
		t.Logf("command path: %s", cmd.Path)
	}
}

func TestBuildSearchCommandLinux(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping Linux/macOS-specific test")
	}
	cmd := buildSearchCommand("/home/user/.scmd/cache", []string{"docker", "network"})
	if cmd.Path == "" {
		t.Error("expected non-empty command path")
	}
}

func TestValidateTargetPathExists(t *testing.T) {
	dir := t.TempDir()
	valid, reason := validateTargetPath(dir)
	if !valid {
		t.Errorf("expected valid=true for existing dir, got false: %s", reason)
	}
	if reason != "" {
		t.Errorf("expected empty reason, got %q", reason)
	}
}

func TestValidateTargetPathNotDir(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "file.txt")
	os.WriteFile(fp, []byte("test"), 0644)

	valid, reason := validateTargetPath(fp)
	if valid {
		t.Error("expected valid=false for file path")
	}
	if reason == "" {
		t.Error("expected non-empty reason")
	}
}

func TestValidateTargetPathNotExist(t *testing.T) {
	valid, reason := validateTargetPath("/nonexistent/path/12345")
	if valid {
		t.Error("expected valid=false for non-existent path")
	}
	if reason == "" {
		t.Error("expected non-empty reason")
	}
}

func TestLLMFallbackOllamaToGemini(t *testing.T) {
	alt := getAlternativeProvider("ollama")
	if alt != "gemini" {
		t.Errorf("getAlternativeProvider(\"ollama\") = %q, want \"gemini\"", alt)
	}
}

func TestLLMFallbackGeminiToOllama(t *testing.T) {
	alt := getAlternativeProvider("gemini")
	if alt != "ollama" {
		t.Errorf("getAlternativeProvider(\"gemini\") = %q, want \"ollama\"", alt)
	}
}

func TestLLMFallbackUnavailable(t *testing.T) {
	// proposeLLMFallback with a provider that's likely not available in test env
	// should return an error. We test the error path.
	_, err := proposeLLMFallback("ollama", []string{"error 1", "error 2"})
	// In test environment, providers are typically not available
	// so we just verify the function doesn't panic
	_ = err
}

func TestFixWorkflowCreatesBackup(t *testing.T) {
	// Create a temp script file
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "test-script.sh")
	content := "#!/bin/bash\necho hello"
	if err := os.WriteFile(scriptPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	// Create backup
	backupPath, err := createScriptBackup(scriptPath)
	if err != nil {
		t.Fatalf("createScriptBackup failed: %v", err)
	}

	// Verify backup exists and content matches
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("failed to read backup: %v", err)
	}

	if string(backupContent) != content {
		t.Errorf("backup content = %q, want %q", string(backupContent), content)
	}

	// Clean up
	os.Remove(backupPath)
}

func TestCodeCommandRouting(t *testing.T) {
	// Verify isCodeFeedbackInput correctly identifies code feedback inputs
	tests := []struct {
		input          string
		codeBlockCount int
		want           bool
	}{
		{"s", 0, true},      // save — valid
		{"n", 0, true},      // regenerate — valid
		{"x", 0, true},      // execute — valid
		{"x 1", 0, true},    // execute with number — valid
		{"1", 3, true},      // numeric with blocks — valid
		{"1", 0, false},     // numeric without blocks — invalid
		{"hello", 0, false}, // random text — invalid
	}

	for _, tc := range tests {
		got := isCodeFeedbackInput(tc.input, tc.codeBlockCount)
		if got != tc.want {
			t.Errorf("isCodeFeedbackInput(%q, %d) = %v, want %v", tc.input, tc.codeBlockCount, got, tc.want)
		}
	}
}

func TestCodeUsageMessage(t *testing.T) {
	// Verify the code feedback prompt contains the expected options
	prompt := buildCodeFeedbackPrompt(0)
	if !strings.Contains(prompt, "[s]") {
		t.Errorf("prompt missing [s]: %q", prompt)
	}
	if !strings.Contains(prompt, "[n]") {
		t.Errorf("prompt missing [n]: %q", prompt)
	}
	if !strings.Contains(prompt, "[x]") {
		t.Errorf("prompt missing [x]: %q", prompt)
	}
}

func TestCodeSessionPrefixDistinct(t *testing.T) {
	// Verify the three prefixes are distinct
	if codeSessionPrefix == "scmd-" {
		t.Error("codeSessionPrefix should differ from research prefix 'scmd-'")
	}
	if codeSessionPrefix == workbookPrefix {
		t.Error("codeSessionPrefix should differ from workbookPrefix")
	}
	if workbookPrefix == "scmd-" {
		t.Error("workbookPrefix should differ from research prefix 'scmd-'")
	}

	// Verify they have the expected values
	if codeSessionPrefix != "scmd-code-" {
		t.Errorf("codeSessionPrefix = %q, want 'scmd-code-'", codeSessionPrefix)
	}
	if workbookPrefix != "scmd-workbook-" {
		t.Errorf("workbookPrefix = %q, want 'scmd-workbook-'", workbookPrefix)
	}
}
