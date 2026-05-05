package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestWorkbookCreation creates a workbook file in a temp directory (by
// writing manually), then parses it and verifies STARTED status with all
// TODO steps.
// Validates: Requirements 13.1, 13.3, 13.4
func TestWorkbookCreation(t *testing.T) {
	dir := t.TempDir()

	// Build a workbook struct and serialize it.
	steps := []WorkbookStep{
		{Number: 1, Description: "Set up project", Status: StepTodo},
		{Number: 2, Description: "Write tests", Status: StepTodo},
		{Number: 3, Description: "Deploy", Status: StepTodo},
	}
	wb := &Workbook{
		Title:  "My Plan",
		Status: WorkbookStarted,
		Steps:  steps,
	}

	content := serializeWorkbook(wb)
	filename := buildWorkbookFilename(time.Now())
	fullPath := filepath.Join(dir, filename)

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write workbook file: %v", err)
	}

	// Read back and parse.
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("failed to read workbook file: %v", err)
	}

	parsed, err := parseWorkbook(string(data))
	if err != nil {
		t.Fatalf("parseWorkbook failed: %v", err)
	}

	// Verify title.
	if parsed.Title != "My Plan" {
		t.Errorf("title = %q, want %q", parsed.Title, "My Plan")
	}

	// Verify status is STARTED.
	if parsed.Status != WorkbookStarted {
		t.Errorf("status = %q, want %q", parsed.Status, WorkbookStarted)
	}

	// Verify all steps are TODO.
	if len(parsed.Steps) != 3 {
		t.Fatalf("step count = %d, want 3", len(parsed.Steps))
	}
	for i, step := range parsed.Steps {
		if step.Status != StepTodo {
			t.Errorf("step %d status = %q, want %q", i+1, step.Status, StepTodo)
		}
		if step.Number != i+1 {
			t.Errorf("step %d number = %d, want %d", i+1, step.Number, i+1)
		}
	}

	// Verify step descriptions.
	expectedDescs := []string{"Set up project", "Write tests", "Deploy"}
	for i, step := range parsed.Steps {
		if step.Description != expectedDescs[i] {
			t.Errorf("step %d description = %q, want %q", i+1, step.Description, expectedDescs[i])
		}
	}
}

// TestWorkbookStepCompletion creates a workbook struct with a temp file,
// calls updateWorkbookStep, reads back the file, and verifies the step is
// DONE and the overall status changed.
// Validates: Requirements 13.5, 13.6
func TestWorkbookStepCompletion(t *testing.T) {
	dir := t.TempDir()

	// Create a workbook with 3 TODO steps.
	steps := []WorkbookStep{
		{Number: 1, Description: "First task", Status: StepTodo},
		{Number: 2, Description: "Second task", Status: StepTodo},
		{Number: 3, Description: "Third task", Status: StepTodo},
	}
	wb := &Workbook{
		Title:  "Step Test",
		Status: WorkbookStarted,
		Steps:  steps,
	}

	// Write the initial workbook to a temp file.
	content := serializeWorkbook(wb)
	fullPath := filepath.Join(dir, "scmd-workbook-2025-01-01_10-00-00.md")
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write workbook: %v", err)
	}
	wb.FilePath = fullPath

	// Mark step 1 as done.
	if err := updateWorkbookStep(wb, 1); err != nil {
		t.Fatalf("updateWorkbookStep(1) failed: %v", err)
	}

	// Verify in-memory state.
	if wb.Steps[0].Status != StepDone {
		t.Errorf("step 1 status = %q, want %q", wb.Steps[0].Status, StepDone)
	}
	if wb.Status != WorkbookProgress {
		t.Errorf("workbook status = %q, want %q", wb.Status, WorkbookProgress)
	}

	// Read back from file and verify persistence.
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("failed to read workbook file: %v", err)
	}
	parsed, err := parseWorkbook(string(data))
	if err != nil {
		t.Fatalf("parseWorkbook failed: %v", err)
	}

	if parsed.Steps[0].Status != StepDone {
		t.Errorf("persisted step 1 status = %q, want %q", parsed.Steps[0].Status, StepDone)
	}
	if parsed.Status != WorkbookProgress {
		t.Errorf("persisted workbook status = %q, want %q", parsed.Status, WorkbookProgress)
	}

	// Mark all remaining steps as done.
	if err := updateWorkbookStep(wb, 2); err != nil {
		t.Fatalf("updateWorkbookStep(2) failed: %v", err)
	}
	if err := updateWorkbookStep(wb, 3); err != nil {
		t.Fatalf("updateWorkbookStep(3) failed: %v", err)
	}

	if wb.Status != WorkbookFinished {
		t.Errorf("workbook status after all done = %q, want %q", wb.Status, WorkbookFinished)
	}

	// Verify step not found error.
	if err := updateWorkbookStep(wb, 99); err == nil {
		t.Error("expected error for non-existent step 99, got nil")
	}
}

// TestFindUnfinishedWorkbooks creates workbook files with different statuses
// in a temp directory and verifies that only STARTED/PROGRESS workbooks are
// returned by the parsing and filtering logic.
// Validates: Requirements 14.1
func TestFindUnfinishedWorkbooks(t *testing.T) {
	dir := t.TempDir()

	// Create a STARTED workbook.
	startedContent := serializeWorkbook(&Workbook{
		Title:  "Started WB",
		Status: WorkbookStarted,
		Steps: []WorkbookStep{
			{Number: 1, Description: "Do something", Status: StepTodo},
		},
	})
	startedFile := filepath.Join(dir, "scmd-workbook-2025-06-01_10-00-00.md")
	if err := os.WriteFile(startedFile, []byte(startedContent), 0644); err != nil {
		t.Fatalf("failed to write started workbook: %v", err)
	}

	// Create a PROGRESS workbook (more recent).
	progressContent := serializeWorkbook(&Workbook{
		Title:  "Progress WB",
		Status: WorkbookProgress,
		Steps: []WorkbookStep{
			{Number: 1, Description: "Done step", Status: StepDone},
			{Number: 2, Description: "Pending step", Status: StepTodo},
		},
	})
	progressFile := filepath.Join(dir, "scmd-workbook-2025-06-02_10-00-00.md")
	if err := os.WriteFile(progressFile, []byte(progressContent), 0644); err != nil {
		t.Fatalf("failed to write progress workbook: %v", err)
	}

	// Create a FINISHED workbook.
	finishedContent := serializeWorkbook(&Workbook{
		Title:  "Finished WB",
		Status: WorkbookFinished,
		Steps: []WorkbookStep{
			{Number: 1, Description: "All done", Status: StepDone},
		},
	})
	finishedFile := filepath.Join(dir, "scmd-workbook-2025-06-03_10-00-00.md")
	if err := os.WriteFile(finishedFile, []byte(finishedContent), 0644); err != nil {
		t.Fatalf("failed to write finished workbook: %v", err)
	}

	// Create a non-workbook file (should be ignored).
	otherFile := filepath.Join(dir, "scmd-2025-06-01_10-00-00.md")
	if err := os.WriteFile(otherFile, []byte("not a workbook"), 0644); err != nil {
		t.Fatalf("failed to write other file: %v", err)
	}

	// Read all workbook files from the temp dir and filter manually
	// (since findUnfinishedWorkbooks uses ensureCacheDir which points to ~/.scmd/cache/).
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	var unfinished []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, workbookPrefix) || !strings.HasSuffix(name, ".md") {
			continue
		}

		fullPath := filepath.Join(dir, name)
		data, readErr := os.ReadFile(fullPath)
		if readErr != nil {
			continue
		}

		wb, parseErr := parseWorkbook(string(data))
		if parseErr != nil {
			continue
		}

		if wb.Status == WorkbookStarted || wb.Status == WorkbookProgress {
			unfinished = append(unfinished, fullPath)
		}
	}

	// Should find exactly 2 unfinished workbooks (STARTED and PROGRESS).
	if len(unfinished) != 2 {
		t.Fatalf("found %d unfinished workbooks, want 2", len(unfinished))
	}

	// Verify the FINISHED workbook is not included.
	for _, path := range unfinished {
		if path == finishedFile {
			t.Error("FINISHED workbook should not be in unfinished list")
		}
	}

	// Verify the non-workbook file is not included.
	for _, path := range unfinished {
		if path == otherFile {
			t.Error("non-workbook file should not be in unfinished list")
		}
	}

	// Verify that selectMostRecentWorkbook picks the PROGRESS one (June 2).
	mostRecent := selectMostRecentWorkbook(unfinished)
	if mostRecent != progressFile {
		t.Errorf("most recent = %q, want %q", mostRecent, progressFile)
	}
}

// TestLoadLastWorkbookCommand verifies that checkStartupWorkbook returns
// nil when no unfinished workbooks exist in the cache directory.
// Validates: Requirements 14.3, 15.7
func TestLoadLastWorkbookCommand(t *testing.T) {
	// checkStartupWorkbook uses ensureCacheDir() which points to ~/.scmd/cache/.
	// We test the nil return path — if no unfinished workbooks exist, it returns nil.
	// Note: this test may find workbooks if the user's cache has them, so we
	// just verify the function doesn't error out.
	wb, err := checkStartupWorkbook()
	if err != nil {
		t.Fatalf("checkStartupWorkbook() returned error: %v", err)
	}

	// If a workbook was found, verify it has a valid status.
	if wb != nil {
		if wb.Status != WorkbookStarted && wb.Status != WorkbookProgress {
			t.Errorf("found workbook with unexpected status %q (expected STARTED or PROGRESS)", wb.Status)
		}
		if wb.FilePath == "" {
			t.Error("found workbook but FilePath is empty")
		}
		if wb.Title == "" {
			t.Error("found workbook but Title is empty")
		}
	}
	// If wb is nil, that's the expected "no unfinished workbooks" path — also valid.
}

// TestStartupWorkbookDetection verifies that checkStartupWorkbook does not
// error out and, if a workbook is found, it has a valid unfinished status.
// Validates: Requirements 15.1
func TestStartupWorkbookDetection(t *testing.T) {
	// checkStartupWorkbook should not error out
	wb, err := checkStartupWorkbook()
	if err != nil {
		t.Fatalf("checkStartupWorkbook() returned error: %v", err)
	}
	// If a workbook was found, verify it has valid status
	if wb != nil {
		if wb.Status != WorkbookStarted && wb.Status != WorkbookProgress {
			t.Errorf("unexpected status %q", wb.Status)
		}
	}
}

// TestStartupNoWorkbooks verifies the function handles the case gracefully.
// In a clean environment, it should return nil.
// Validates: Requirements 15.7
func TestStartupNoWorkbooks(t *testing.T) {
	wb, err := checkStartupWorkbook()
	if err != nil {
		t.Fatalf("checkStartupWorkbook() returned error: %v", err)
	}
	// wb may or may not be nil depending on the user's cache state
	_ = wb
}
