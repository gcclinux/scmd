package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const workbookTimestampLayout = "2006-01-02_15-04-05"
const workbookPrefix = "scmd-workbook-"

// WorkbookStatus represents the overall state of a workbook.
type WorkbookStatus string

const (
	WorkbookStarted  WorkbookStatus = "STARTED"
	WorkbookProgress WorkbookStatus = "PROGRESS"
	WorkbookFinished WorkbookStatus = "FINISHED"
)

// StepStatus represents the state of an individual workbook step.
type StepStatus string

const (
	StepTodo StepStatus = "TODO"
	StepDone StepStatus = "DONE"
)

// WorkbookStep represents a single action item in a workbook.
type WorkbookStep struct {
	Number      int
	Description string
	Status      StepStatus
}

// Workbook represents a parsed workbook file.
type Workbook struct {
	Title    string
	Status   WorkbookStatus
	Steps    []WorkbookStep
	FilePath string
}

// buildWorkbookFilename returns a filename of the form
// scmd-workbook-YYYY-MM-DD_HH-mm-SS.md for the given time.
func buildWorkbookFilename(t time.Time) string {
	return workbookPrefix + t.Format(workbookTimestampLayout) + ".md"
}

// computeWorkbookStatus computes the correct overall status based on step states.
// Returns STARTED when all steps are TODO (or no steps exist),
// FINISHED when all steps are DONE, and PROGRESS otherwise.
func computeWorkbookStatus(steps []WorkbookStep) WorkbookStatus {
	if len(steps) == 0 {
		return WorkbookStarted
	}
	allTodo := true
	allDone := true
	for _, s := range steps {
		if s.Status != StepTodo {
			allTodo = false
		}
		if s.Status != StepDone {
			allDone = false
		}
	}
	if allTodo {
		return WorkbookStarted
	}
	if allDone {
		return WorkbookFinished
	}
	return WorkbookProgress
}

// serializeWorkbook converts a *Workbook to the markdown format:
//
//	# Workbook: <title>
//
//	**Status:** <STATUS>
//	**Created:** YYYY-MM-DD HH:mm:SS
//
//	## Steps
//
//	- [TODO] 1. First step description
//	- [DONE] 2. Second step description
func serializeWorkbook(wb *Workbook) string {
	var b strings.Builder
	b.WriteString("# Workbook: ")
	b.WriteString(wb.Title)
	b.WriteString("\n\n")
	b.WriteString("**Status:** ")
	b.WriteString(string(wb.Status))
	b.WriteString("\n")
	b.WriteString("**Created:** ")
	b.WriteString(time.Now().Format("2006-01-02 15:04:05"))
	b.WriteString("\n\n")
	b.WriteString("## Steps\n\n")
	for _, step := range wb.Steps {
		b.WriteString(fmt.Sprintf("- [%s] %d. %s\n", step.Status, step.Number, step.Description))
	}
	return b.String()
}

// parseWorkbook parses a workbook markdown file into a Workbook struct.
// It extracts the title from the "# Workbook: " heading, the status from
// the "**Status:**" line, and each step from lines matching
// "- [TODO] N. description" or "- [DONE] N. description".
// Returns an error if the format is invalid.
func parseWorkbook(content string) (*Workbook, error) {
	lines := strings.Split(content, "\n")
	wb := &Workbook{}

	foundTitle := false
	foundStatus := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Parse title
		if strings.HasPrefix(trimmed, "# Workbook: ") {
			wb.Title = strings.TrimPrefix(trimmed, "# Workbook: ")
			foundTitle = true
			continue
		}

		// Parse status
		if strings.HasPrefix(trimmed, "**Status:**") {
			statusStr := strings.TrimSpace(strings.TrimPrefix(trimmed, "**Status:**"))
			switch WorkbookStatus(statusStr) {
			case WorkbookStarted, WorkbookProgress, WorkbookFinished:
				wb.Status = WorkbookStatus(statusStr)
				foundStatus = true
			default:
				return nil, fmt.Errorf("invalid workbook status: %q", statusStr)
			}
			continue
		}

		// Parse steps: - [TODO] N. description or - [DONE] N. description
		if strings.HasPrefix(trimmed, "- [TODO] ") || strings.HasPrefix(trimmed, "- [DONE] ") {
			step, err := parseStepLine(trimmed)
			if err != nil {
				return nil, err
			}
			wb.Steps = append(wb.Steps, step)
		}
	}

	if !foundTitle {
		return nil, errors.New("workbook missing title (expected '# Workbook: <title>')")
	}
	if !foundStatus {
		return nil, errors.New("workbook missing status (expected '**Status:** <STATUS>')")
	}

	return wb, nil
}

// parseStepLine parses a single step line like "- [TODO] 1. Fix the bug"
// into a WorkbookStep.
func parseStepLine(line string) (WorkbookStep, error) {
	var status StepStatus
	var rest string

	if strings.HasPrefix(line, "- [TODO] ") {
		status = StepTodo
		rest = strings.TrimPrefix(line, "- [TODO] ")
	} else if strings.HasPrefix(line, "- [DONE] ") {
		status = StepDone
		rest = strings.TrimPrefix(line, "- [DONE] ")
	} else {
		return WorkbookStep{}, fmt.Errorf("invalid step line: %q", line)
	}

	// Parse "N. description"
	dotIdx := strings.Index(rest, ". ")
	if dotIdx < 1 {
		return WorkbookStep{}, fmt.Errorf("invalid step format (expected 'N. description'): %q", line)
	}

	var num int
	_, err := fmt.Sscanf(rest[:dotIdx], "%d", &num)
	if err != nil {
		return WorkbookStep{}, fmt.Errorf("invalid step number in: %q", line)
	}

	description := rest[dotIdx+2:]

	return WorkbookStep{
		Number:      num,
		Description: description,
		Status:      status,
	}, nil
}

// findFirstTodoStep returns the index of the first step with TODO status,
// or an error if no TODO steps are found.
func findFirstTodoStep(wb *Workbook) (int, error) {
	for i, step := range wb.Steps {
		if step.Status == StepTodo {
			return i, nil
		}
	}
	return -1, errors.New("no TODO steps found")
}

// selectMostRecentWorkbook returns the path with the most recent timestamp
// from a list of workbook file paths. If the list is empty, returns an
// empty string. Timestamps are parsed from the filename by stripping the
// "scmd-workbook-" prefix and ".md" suffix.
func selectMostRecentWorkbook(paths []string) string {
	if len(paths) == 0 {
		return ""
	}

	type entry struct {
		path string
		ts   time.Time
	}

	var entries []entry
	for _, p := range paths {
		base := filepath.Base(p)
		tsStr := strings.TrimPrefix(base, workbookPrefix)
		tsStr = strings.TrimSuffix(tsStr, ".md")
		t, err := time.Parse(workbookTimestampLayout, tsStr)
		if err != nil {
			continue
		}
		entries = append(entries, entry{path: p, ts: t})
	}

	if len(entries) == 0 {
		return ""
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ts.After(entries[j].ts)
	})

	return entries[0].path
}

// createWorkbook creates a new workbook file in the Cache_Directory with
// STARTED status and all steps as TODO. It returns the populated Workbook
// struct (including FilePath) or an error if the file cannot be written.
func createWorkbook(title string, steps []string) (*Workbook, error) {
	dir, err := ensureCacheDir()
	if err != nil {
		return nil, err
	}

	wbSteps := make([]WorkbookStep, len(steps))
	for i, desc := range steps {
		wbSteps[i] = WorkbookStep{
			Number:      i + 1,
			Description: desc,
			Status:      StepTodo,
		}
	}

	wb := &Workbook{
		Title:  title,
		Status: WorkbookStarted,
		Steps:  wbSteps,
	}

	filename := buildWorkbookFilename(time.Now())
	content := serializeWorkbook(wb)
	fullPath := filepath.Join(dir, filename)

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return nil, err
	}

	wb.FilePath = fullPath
	return wb, nil
}

// updateWorkbookStep marks the specified step as DONE, recomputes the
// overall workbook status, and writes the updated workbook back to its file.
func updateWorkbookStep(wb *Workbook, stepNumber int) error {
	found := false
	for i := range wb.Steps {
		if wb.Steps[i].Number == stepNumber {
			wb.Steps[i].Status = StepDone
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("step %d not found", stepNumber)
	}

	wb.Status = computeWorkbookStatus(wb.Steps)
	content := serializeWorkbook(wb)

	if wb.FilePath == "" {
		return errors.New("workbook has no file path")
	}

	return os.WriteFile(wb.FilePath, []byte(content), 0644)
}

// findUnfinishedWorkbooks scans the Cache_Directory for workbook files
// with STARTED or PROGRESS status. It returns their full paths sorted
// by timestamp (most recent first).
func findUnfinishedWorkbooks() ([]string, error) {
	dir, err := ensureCacheDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
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
		data, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		wb, err := parseWorkbook(string(data))
		if err != nil {
			continue
		}

		if wb.Status == WorkbookStarted || wb.Status == WorkbookProgress {
			unfinished = append(unfinished, fullPath)
		}
	}

	// Sort by timestamp (most recent first).
	sort.Slice(unfinished, func(i, j int) bool {
		tsI := extractWorkbookTimestamp(unfinished[i])
		tsJ := extractWorkbookTimestamp(unfinished[j])
		return tsI.After(tsJ)
	})

	return unfinished, nil
}

// extractWorkbookTimestamp parses the timestamp from a workbook file path.
func extractWorkbookTimestamp(path string) time.Time {
	base := filepath.Base(path)
	tsStr := strings.TrimPrefix(base, workbookPrefix)
	tsStr = strings.TrimSuffix(tsStr, ".md")
	t, err := time.Parse(workbookTimestampLayout, tsStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

// checkStartupWorkbook finds and parses the most recent unfinished workbook.
// Returns nil (with no error) if no unfinished workbooks exist.
func checkStartupWorkbook() (*Workbook, error) {
	paths, err := findUnfinishedWorkbooks()
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, nil
	}

	mostRecent := selectMostRecentWorkbook(paths)
	if mostRecent == "" {
		return nil, nil
	}

	data, err := os.ReadFile(mostRecent)
	if err != nil {
		return nil, err
	}

	wb, err := parseWorkbook(string(data))
	if err != nil {
		return nil, err
	}

	wb.FilePath = mostRecent
	return wb, nil
}
