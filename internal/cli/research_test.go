package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestParseResearchQuery_NoFile verifies that when no token is a valid file
// path, the entire args string is returned as the query.
func TestParseResearchQuery_NoFile(t *testing.T) {
	query, fileContent, filePath, err := parseResearchQuery("explain docker networking")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if query != "explain docker networking" {
		t.Errorf("query = %q, want %q", query, "explain docker networking")
	}
	if fileContent != "" {
		t.Errorf("fileContent = %q, want empty", fileContent)
	}
	if filePath != "" {
		t.Errorf("filePath = %q, want empty", filePath)
	}
}

// TestParseResearchQuery_EmptyArgs verifies that empty input returns all
// empty values with no error.
func TestParseResearchQuery_EmptyArgs(t *testing.T) {
	query, fileContent, filePath, err := parseResearchQuery("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if query != "" || fileContent != "" || filePath != "" {
		t.Errorf("expected all empty, got query=%q fileContent=%q filePath=%q", query, fileContent, filePath)
	}
}

// TestParseResearchQuery_ValidFile verifies that a valid file path token is
// detected, its contents are read, and the remaining tokens form the query.
func TestParseResearchQuery_ValidFile(t *testing.T) {
	// Create a temp file with known content.
	dir := t.TempDir()
	fp := filepath.Join(dir, "test.txt")
	content := "hello world\nline two"
	if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	query, fileContent, filePath, err := parseResearchQuery("analyse this " + fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if query != "analyse this" {
		t.Errorf("query = %q, want %q", query, "analyse this")
	}
	if fileContent != content {
		t.Errorf("fileContent = %q, want %q", fileContent, content)
	}
	if filePath != fp {
		t.Errorf("filePath = %q, want %q", filePath, fp)
	}
}

// TestParseResearchQuery_FileOnly verifies that when the file path is the
// only token, the query is empty.
func TestParseResearchQuery_FileOnly(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "script.sh")
	if err := os.WriteFile(fp, []byte("#!/bin/bash\necho hi"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	query, fileContent, filePath, err := parseResearchQuery(fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if query != "" {
		t.Errorf("query = %q, want empty", query)
	}
	if fileContent != "#!/bin/bash\necho hi" {
		t.Errorf("fileContent = %q, want %q", fileContent, "#!/bin/bash\necho hi")
	}
	if filePath != fp {
		t.Errorf("filePath = %q, want %q", filePath, fp)
	}
}

// TestParseResearchQuery_NonExistentFile verifies that a non-existent path
// is not treated as a file — the entire args is returned as the query.
func TestParseResearchQuery_NonExistentFile(t *testing.T) {
	query, fileContent, filePath, err := parseResearchQuery("review /no/such/file.txt please")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if query != "review /no/such/file.txt please" {
		t.Errorf("query = %q, want %q", query, "review /no/such/file.txt please")
	}
	if fileContent != "" {
		t.Errorf("fileContent = %q, want empty", fileContent)
	}
	if filePath != "" {
		t.Errorf("filePath = %q, want empty", filePath)
	}
}

// TestParseResearchQuery_UnreadableFile verifies that a file that exists but
// cannot be read returns an error with the OS error message.
func TestParseResearchQuery_UnreadableFile(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping unreadable file test when running as root")
	}

	dir := t.TempDir()
	fp := filepath.Join(dir, "secret.txt")
	if err := os.WriteFile(fp, []byte("secret"), 0000); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, _, _, err := parseResearchQuery("check " + fp)
	if err == nil {
		t.Fatal("expected error for unreadable file, got nil")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("error = %q, expected it to contain 'permission denied'", err.Error())
	}
}

// TestParseResearchQuery_Directory verifies that a directory path is not
// treated as a file — it is skipped.
func TestParseResearchQuery_Directory(t *testing.T) {
	dir := t.TempDir()

	query, fileContent, filePath, err := parseResearchQuery("review " + dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fileContent != "" {
		t.Errorf("fileContent = %q, want empty (directories should be skipped)", fileContent)
	}
	if filePath != "" {
		t.Errorf("filePath = %q, want empty", filePath)
	}
	if query != "review "+dir {
		t.Errorf("query = %q, want %q", query, "review "+dir)
	}
}

// TestResearchFileReadValid creates a temp file and verifies that
// parseResearchQuery reads its contents and includes them in the result.
// Validates: Requirements 3.1
func TestResearchFileReadValid(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "sample.txt")
	content := "line one\nline two\nline three"
	if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	query, fileContent, filePath, err := parseResearchQuery("summarise " + fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if query != "summarise" {
		t.Errorf("query = %q, want %q", query, "summarise")
	}
	if fileContent != content {
		t.Errorf("fileContent = %q, want %q", fileContent, content)
	}
	if filePath != fp {
		t.Errorf("filePath = %q, want %q", filePath, fp)
	}
}

// TestResearchFileReadInvalid passes a non-existent file path and verifies
// that the path is not treated as a file — the entire args string is
// returned as the query with no error (non-existent paths are skipped).
// Validates: Requirements 3.2
func TestResearchFileReadInvalid(t *testing.T) {
	nonExistent := "/tmp/scmd_test_nonexistent_" + t.Name() + ".txt"
	// Ensure the file truly does not exist.
	os.Remove(nonExistent)

	query, fileContent, filePath, err := parseResearchQuery("review " + nonExistent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fileContent != "" {
		t.Errorf("fileContent = %q, want empty for non-existent path", fileContent)
	}
	if filePath != "" {
		t.Errorf("filePath = %q, want empty for non-existent path", filePath)
	}
	expectedQuery := "review " + nonExistent
	if query != expectedQuery {
		t.Errorf("query = %q, want %q", query, expectedQuery)
	}
}

// TestEnsureCacheDirCreates verifies that ensureCacheDir creates the
// ~/.scmd/cache/ directory with 0755 permissions and returns a path
// ending in ".scmd/cache".
// Validates: Requirements 5.1
func TestEnsureCacheDirCreates(t *testing.T) {
	dir, err := ensureCacheDir()
	if err != nil {
		t.Fatalf("ensureCacheDir() returned error: %v", err)
	}

	// The returned path must end with .scmd/cache.
	if !strings.HasSuffix(dir, filepath.Join(".scmd", "cache")) {
		t.Errorf("dir = %q, want suffix %q", dir, filepath.Join(".scmd", "cache"))
	}

	// The directory must exist.
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("os.Stat(%q) returned error: %v", dir, err)
	}
	if !info.IsDir() {
		t.Fatalf("%q is not a directory", dir)
	}

	// Permissions must be 0755 (on Unix systems the permission bits are
	// meaningful; on Windows this check is a no-op since Stat always
	// returns 0777).
	perm := info.Mode().Perm()
	if perm != 0755 {
		t.Errorf("permissions = %o, want 0755", perm)
	}
}

// TestEnsureCacheDirExists verifies that calling ensureCacheDir twice
// returns the same path without error — the second call is a no-op on
// an already-existing directory.
// Validates: Requirements 5.2
func TestEnsureCacheDirExists(t *testing.T) {
	dir1, err1 := ensureCacheDir()
	if err1 != nil {
		t.Fatalf("first ensureCacheDir() returned error: %v", err1)
	}

	dir2, err2 := ensureCacheDir()
	if err2 != nil {
		t.Fatalf("second ensureCacheDir() returned error: %v", err2)
	}

	if dir1 != dir2 {
		t.Errorf("paths differ: first = %q, second = %q", dir1, dir2)
	}
}

// TestResearchUsageMessage verifies the empty-args path that underpins the
// usage message behaviour. When args is empty, parseResearchQuery returns
// all empty strings and no error, which is the condition the /research case
// in handleSlashCommand uses to display the usage message instead of
// calling handleResearchCommand.
// Validates: Requirements 1.2
func TestResearchUsageMessage(t *testing.T) {
	query, fileContent, filePath, err := parseResearchQuery("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if query != "" {
		t.Errorf("query = %q, want empty for no-args invocation", query)
	}
	if fileContent != "" {
		t.Errorf("fileContent = %q, want empty for no-args invocation", fileContent)
	}
	if filePath != "" {
		t.Errorf("filePath = %q, want empty for no-args invocation", filePath)
	}
}

// TestResearchCommandRouting verifies that the `/research` routing logic
// correctly identifies valid and invalid research feedback inputs via
// isResearchFeedbackInput. We cannot call handleSlashCommand directly
// because it requires database/AI providers, but we can verify the routing
// logic that determines whether input enters the research feedback path.
// Validates: Requirements 1.1
func TestResearchCommandRouting(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"s", true},   // save — valid research feedback
		{"n", true},   // regenerate — valid research feedback
		{"x", false},  // execute — NOT valid for research
		{"1", false},  // numeric block selector — NOT valid for research
		{"anything else", false},
	}

	for _, tc := range tests {
		got := isResearchFeedbackInput(tc.input)
		if got != tc.want {
			t.Errorf("isResearchFeedbackInput(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

// TestResearchSaveToDBFeedback verifies the feedback input validation for
// the save path: "s" is accepted as valid research feedback, and the
// research feedback prompt contains the "[s]" option so the user knows
// saving is available.
// Validates: Requirements 8.3
func TestResearchSaveToDBFeedback(t *testing.T) {
	// "s" must be recognised as valid research feedback input.
	if !isResearchFeedbackInput("s") {
		t.Error("isResearchFeedbackInput(\"s\") = false, want true")
	}

	// The research feedback prompt must advertise the save option.
	prompt := buildResearchFeedbackPrompt()
	if !strings.Contains(prompt, "[s]") {
		t.Errorf("buildResearchFeedbackPrompt() = %q, want it to contain \"[s]\"", prompt)
	}
}

// TestResearchRegenerateFeedback verifies the feedback input validation for
// the regenerate path: "n" is accepted as valid research feedback, the
// prompt contains "[n]", successive UUID generation produces distinct IDs,
// and different timestamps produce different filenames — confirming that
// regeneration would result in a new UUID and a new file.
// Validates: Requirements 8.4
func TestResearchRegenerateFeedback(t *testing.T) {
	// "n" must be recognised as valid research feedback input.
	if !isResearchFeedbackInput("n") {
		t.Error("isResearchFeedbackInput(\"n\") = false, want true")
	}

	// The research feedback prompt must advertise the regenerate option.
	prompt := buildResearchFeedbackPrompt()
	if !strings.Contains(prompt, "[n]") {
		t.Errorf("buildResearchFeedbackPrompt() = %q, want it to contain \"[n]\"", prompt)
	}

	// Successive calls to generateRecommendationID must produce different UUIDs,
	// confirming that regeneration would get a new unique ID.
	uid1 := generateRecommendationID()
	uid2 := generateRecommendationID()
	if uid1 == uid2 {
		t.Errorf("generateRecommendationID() produced duplicate UUIDs: %q", uid1)
	}

	// Different timestamps must produce different filenames, confirming that
	// regeneration at a different time would write to a new file.
	t1 := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	t2 := time.Date(2025, 6, 15, 10, 30, 1, 0, time.UTC) // one second later
	f1 := buildRecommendationFilename(t1)
	f2 := buildRecommendationFilename(t2)
	if f1 == f2 {
		t.Errorf("buildRecommendationFilename produced identical filenames for different times: %q", f1)
	}
}
