package cli

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"testing"
	"time"

	"pgregory.net/rapid"
)

// workbookFilenamePattern matches the expected workbook filename format:
// scmd-workbook-YYYY-MM-DD_HH-mm-SS.md with zero-padded 24-hour time components.
var workbookFilenamePattern = regexp.MustCompile(
	`^scmd-workbook-\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}\.md$`,
)

// Feature: code-persona, Property 11: Workbook filename format
// **Validates: Requirements 13.2, 16.1**
func TestProperty_WorkbookFilenameFormat(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random Unix timestamp in range [0, 253402300799] (1970-01-01 to 9999-12-31).
		unixSec := rapid.Int64Range(0, 253402300799).Draw(rt, "unixSec")
		ts := time.Unix(unixSec, 0)

		filename := buildWorkbookFilename(ts)

		// Property: the filename must match the strict pattern scmd-workbook-YYYY-MM-DD_HH-mm-SS.md
		if !workbookFilenamePattern.MatchString(filename) {
			rt.Fatalf("buildWorkbookFilename(%v) = %q does not match pattern %s",
				ts, filename, workbookFilenamePattern.String())
		}
	})
}

// Feature: code-persona, Property 12: Workbook filename round-trip
// **Validates: Requirements 16.2**
func TestProperty_WorkbookFilenameRoundTrip(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random Unix timestamp in range [0, 253402300799].
		unixSec := rapid.Int64Range(0, 253402300799).Draw(rt, "unixSec")
		ts := time.Unix(unixSec, 0)

		// Build the workbook filename from the timestamp.
		filename := buildWorkbookFilename(ts)

		// Strip the "scmd-workbook-" prefix and ".md" suffix to extract the timestamp portion.
		timestampStr := filename[len("scmd-workbook-") : len(filename)-len(".md")]

		// Parse the timestamp back using the same location as the original time.
		parsed, err := time.ParseInLocation(workbookTimestampLayout, timestampStr, ts.Location())
		if err != nil {
			rt.Fatalf("failed to parse timestamp %q back: %v", timestampStr, err)
		}

		// The original time truncated to the second should equal the parsed time.
		expected := ts.Truncate(time.Second)
		if !parsed.Equal(expected) {
			rt.Fatalf("round-trip mismatch: original truncated = %v, parsed = %v (filename = %q)",
				expected, parsed, filename)
		}
	})
}

// Feature: code-persona, Property 13: Workbook status invariant
// **Validates: Requirements 13.5, 13.6, 13.7**
func TestProperty_WorkbookStatusInvariant(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random non-empty list of steps with random statuses.
		n := rapid.IntRange(1, 50).Draw(rt, "numSteps")
		steps := make([]WorkbookStep, n)
		for i := 0; i < n; i++ {
			status := rapid.SampledFrom([]StepStatus{StepTodo, StepDone}).Draw(rt, fmt.Sprintf("status_%d", i))
			steps[i] = WorkbookStep{
				Number:      i + 1,
				Description: fmt.Sprintf("Step %d", i+1),
				Status:      status,
			}
		}

		result := computeWorkbookStatus(steps)

		// Count TODO and DONE steps.
		todoCount := 0
		doneCount := 0
		for _, s := range steps {
			if s.Status == StepTodo {
				todoCount++
			} else {
				doneCount++
			}
		}

		// Property: all TODO → STARTED, all DONE → FINISHED, mix → PROGRESS.
		if todoCount == n && doneCount == 0 {
			if result != WorkbookStarted {
				rt.Fatalf("all TODO steps: expected STARTED, got %s", result)
			}
		} else if doneCount == n && todoCount == 0 {
			if result != WorkbookFinished {
				rt.Fatalf("all DONE steps: expected FINISHED, got %s", result)
			}
		} else {
			if result != WorkbookProgress {
				rt.Fatalf("mixed steps (TODO=%d, DONE=%d): expected PROGRESS, got %s",
					todoCount, doneCount, result)
			}
		}
	})
}

// Feature: code-persona, Property 14: Workbook parse/serialize round-trip
// **Validates: Requirements 13.9**
func TestProperty_WorkbookRoundTrip(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random title (simple alphanumeric, no newlines).
		title := rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9]{0,29}`).Draw(rt, "title")

		// Generate random steps with sequential numbers.
		numSteps := rapid.IntRange(1, 20).Draw(rt, "numSteps")
		steps := make([]WorkbookStep, numSteps)
		for i := 0; i < numSteps; i++ {
			status := rapid.SampledFrom([]StepStatus{StepTodo, StepDone}).Draw(rt, fmt.Sprintf("stepStatus_%d", i))
			// Generate simple alphanumeric description with spaces (no newlines, no brackets,
			// no dots at start, no trailing spaces — trailing spaces are trimmed by parseWorkbook).
			desc := rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9]*( [a-zA-Z0-9]+){0,5}`).Draw(rt, fmt.Sprintf("desc_%d", i))
			steps[i] = WorkbookStep{
				Number:      i + 1,
				Description: desc,
				Status:      status,
			}
		}

		// Compute the expected status from the steps.
		expectedStatus := computeWorkbookStatus(steps)

		// Build the workbook.
		wb := &Workbook{
			Title:  title,
			Status: expectedStatus,
			Steps:  steps,
		}

		// Serialize and parse back.
		serialized := serializeWorkbook(wb)
		parsed, err := parseWorkbook(serialized)
		if err != nil {
			rt.Fatalf("parseWorkbook failed on serialized content: %v\nContent:\n%s", err, serialized)
		}

		// Verify title matches.
		if parsed.Title != title {
			rt.Fatalf("title mismatch: got %q, want %q", parsed.Title, title)
		}

		// Verify status matches (the serialized status is what parseWorkbook reads back).
		if parsed.Status != expectedStatus {
			rt.Fatalf("status mismatch: got %q, want %q", parsed.Status, expectedStatus)
		}

		// Verify step count matches.
		if len(parsed.Steps) != numSteps {
			rt.Fatalf("step count mismatch: got %d, want %d", len(parsed.Steps), numSteps)
		}

		// Verify each step's number, description, and status.
		for i := 0; i < numSteps; i++ {
			if parsed.Steps[i].Number != steps[i].Number {
				rt.Fatalf("step %d number mismatch: got %d, want %d",
					i, parsed.Steps[i].Number, steps[i].Number)
			}
			if parsed.Steps[i].Description != steps[i].Description {
				rt.Fatalf("step %d description mismatch: got %q, want %q",
					i, parsed.Steps[i].Description, steps[i].Description)
			}
			if parsed.Steps[i].Status != steps[i].Status {
				rt.Fatalf("step %d status mismatch: got %q, want %q",
					i, parsed.Steps[i].Status, steps[i].Status)
			}
		}
	})
}

// Feature: code-persona, Property 15: Most recent workbook selection
// **Validates: Requirements 14.2**
func TestProperty_MostRecentWorkbookSelection(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a non-empty set of distinct timestamps.
		n := rapid.IntRange(1, 50).Draw(rt, "numTimestamps")

		// Use a map to ensure distinct timestamps (second-level granularity).
		seen := make(map[int64]bool)
		var timestamps []time.Time
		for len(timestamps) < n {
			unixSec := rapid.Int64Range(0, 253402300799).Draw(rt, fmt.Sprintf("ts_%d", len(timestamps)))
			if seen[unixSec] {
				continue
			}
			seen[unixSec] = true
			timestamps = append(timestamps, time.Unix(unixSec, 0))
		}

		// Build workbook filenames from the timestamps.
		paths := make([]string, n)
		for i, ts := range timestamps {
			paths[i] = buildWorkbookFilename(ts)
		}

		// Shuffle the list to ensure order doesn't matter.
		rng := rand.New(rand.NewSource(rapid.Int64().Draw(rt, "seed")))
		rng.Shuffle(len(paths), func(i, j int) {
			paths[i], paths[j] = paths[j], paths[i]
		})

		// Find the expected latest timestamp.
		sortedTimestamps := make([]time.Time, len(timestamps))
		copy(sortedTimestamps, timestamps)
		sort.Slice(sortedTimestamps, func(i, j int) bool {
			return sortedTimestamps[i].After(sortedTimestamps[j])
		})
		expectedPath := buildWorkbookFilename(sortedTimestamps[0])

		// Call selectMostRecentWorkbook.
		result := selectMostRecentWorkbook(paths)

		// Property: the returned path must have the latest timestamp.
		if result != expectedPath {
			rt.Fatalf("selectMostRecentWorkbook returned %q, want %q (from %d paths)",
				result, expectedPath, n)
		}
	})
}

// Feature: code-persona, Property 16: First TODO step selection
// **Validates: Requirements 14.4**
func TestProperty_FirstTodoStepSelection(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a workbook with at least one TODO step.
		numSteps := rapid.IntRange(1, 30).Draw(rt, "numSteps")
		steps := make([]WorkbookStep, numSteps)

		// Ensure at least one TODO step exists.
		// First, generate random statuses.
		for i := 0; i < numSteps; i++ {
			status := rapid.SampledFrom([]StepStatus{StepTodo, StepDone}).Draw(rt, fmt.Sprintf("status_%d", i))
			steps[i] = WorkbookStep{
				Number:      i + 1,
				Description: fmt.Sprintf("Step %d", i+1),
				Status:      status,
			}
		}

		// Ensure at least one TODO step by forcing a random step to TODO.
		hasTodo := false
		for _, s := range steps {
			if s.Status == StepTodo {
				hasTodo = true
				break
			}
		}
		if !hasTodo {
			forceIdx := rapid.IntRange(0, numSteps-1).Draw(rt, "forceIdx")
			steps[forceIdx].Status = StepTodo
		}

		wb := &Workbook{
			Title:  "Test Workbook",
			Status: computeWorkbookStatus(steps),
			Steps:  steps,
		}

		// Call findFirstTodoStep.
		idx, err := findFirstTodoStep(wb)

		// Property: no error returned.
		if err != nil {
			rt.Fatalf("findFirstTodoStep returned error: %v", err)
		}

		// Property: the returned index points to a TODO step.
		if idx < 0 || idx >= len(steps) {
			rt.Fatalf("findFirstTodoStep returned index %d out of range [0, %d)", idx, len(steps))
		}
		if steps[idx].Status != StepTodo {
			rt.Fatalf("findFirstTodoStep returned index %d which has status %q, want TODO",
				idx, steps[idx].Status)
		}

		// Property: all steps before that index are DONE.
		for i := 0; i < idx; i++ {
			if steps[i].Status != StepDone {
				rt.Fatalf("step %d before first TODO (index %d) has status %q, want DONE",
					i, idx, steps[i].Status)
			}
		}
	})
}
