package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"pgregory.net/rapid"
)

// codeSessionFilenamePattern matches the expected code session filename format:
// scmd-code-YYYY-MM-DD_HH-mm-SS.md with zero-padded 24-hour time components.
var codeSessionFilenamePattern = regexp.MustCompile(
	`^scmd-code-\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}\.md$`,
)

// backupFilenameTimestampPattern matches the timestamp suffix portion of a backup filename.
var backupFilenameTimestampPattern = regexp.MustCompile(
	`\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}$`,
)

// Feature: code-persona, Property 1: Query matcher threshold correctness
// **Validates: Requirements 3.1, 3.2**
func TestProperty_QueryMatcherThreshold(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random query: 1-10 space-separated words.
		wordCount := rapid.IntRange(1, 10).Draw(rt, "wordCount")
		words := make([]string, wordCount)
		for i := 0; i < wordCount; i++ {
			words[i] = rapid.StringMatching(`[a-zA-Z]{1,8}`).Draw(rt, fmt.Sprintf("word%d", i))
		}
		query := strings.Join(words, " ")

		// Generate random content: 1-20 space-separated words.
		contentWordCount := rapid.IntRange(1, 20).Draw(rt, "contentWordCount")
		contentWords := make([]string, contentWordCount)
		for i := 0; i < contentWordCount; i++ {
			contentWords[i] = rapid.StringMatching(`[a-zA-Z]{1,8}`).Draw(rt, fmt.Sprintf("cword%d", i))
		}
		content := strings.Join(contentWords, " ")

		overlap := computeWordOverlap(query, content)
		relevant := isRelevantMatch(query, content)

		// Property: isRelevantMatch returns true iff computeWordOverlap >= 0.70
		if overlap >= queryMatchThreshold && !relevant {
			rt.Fatalf("computeWordOverlap(%q, %q) = %f >= %f but isRelevantMatch returned false",
				query, content, overlap, queryMatchThreshold)
		}
		if overlap < queryMatchThreshold && relevant {
			rt.Fatalf("computeWordOverlap(%q, %q) = %f < %f but isRelevantMatch returned true",
				query, content, overlap, queryMatchThreshold)
		}
	})
}

// Feature: code-persona, Property 2: Code session filename format
// **Validates: Requirements 6.2, 9.1**
func TestProperty_CodeSessionFilenameFormat(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random time.Time by drawing a random Unix timestamp.
		// Range: 0 (1970-01-01) to 253402300799 (9999-12-31 23:59:59 UTC).
		unixSec := rapid.Int64Range(0, 253402300799).Draw(rt, "unixSec")
		ts := time.Unix(unixSec, 0)

		filename := buildCodeSessionFilename(ts)

		// Property: the filename must match the strict pattern scmd-code-YYYY-MM-DD_HH-mm-SS.md
		if !codeSessionFilenamePattern.MatchString(filename) {
			rt.Fatalf("buildCodeSessionFilename(%v) = %q does not match pattern %s",
				ts, filename, codeSessionFilenamePattern.String())
		}
	})
}

// Feature: code-persona, Property 3: Code session filename round-trip
// **Validates: Requirements 9.2**
func TestProperty_CodeSessionFilenameRoundTrip(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random Unix timestamp in the range [0, 9999-12-31 23:59:59 UTC].
		unixSec := rapid.Int64Range(0, 253402300799).Draw(rt, "unixSec")
		ts := time.Unix(unixSec, 0)

		// Build the code session filename from the timestamp.
		filename := buildCodeSessionFilename(ts)

		// Strip the "scmd-code-" prefix and ".md" suffix to extract the timestamp portion.
		timestampStr := filename[len("scmd-code-") : len(filename)-len(".md")]

		// Parse the timestamp back using the same location as the original time.
		parsed, err := time.ParseInLocation(codeTimestampLayout, timestampStr, ts.Location())
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

// Feature: code-persona, Property 4: Code session file content structure
// **Validates: Requirements 6.3, 6.4, 6.5**
func TestProperty_CodeSessionContentStructure(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random UUID v4 using the existing helper.
		uid := drawUUIDv4(rt)

		// Generate a random non-empty query string (no newlines).
		query := rapid.StringMatching(`[a-zA-Z0-9 ]+`).Draw(rt, "query")

		// Generate a random non-empty response string (may contain newlines).
		response := rapid.StringMatching(`[a-zA-Z0-9 \n]+`).Draw(rt, "response")

		// Assemble the full code session file content.
		content := assembleCodeSessionContent(uid, query, response)

		// Split into exactly 5 parts: header, blank, heading, blank, remainder.
		parts := strings.SplitN(content, "\n", 5)
		if len(parts) != 5 {
			rt.Fatalf("expected 5 parts from SplitN, got %d for content:\n%s", len(parts), content)
		}

		// Line 1: the ID header.
		expectedHeader := "<!-- SCMD-ID: " + uid + " -->"
		if parts[0] != expectedHeader {
			rt.Fatalf("line 1: got %q, want %q", parts[0], expectedHeader)
		}

		// Line 2: blank.
		if parts[1] != "" {
			rt.Fatalf("line 2: expected blank, got %q", parts[1])
		}

		// Line 3: markdown heading with the query.
		expectedHeading := "# " + query
		if parts[2] != expectedHeading {
			rt.Fatalf("line 3: got %q, want %q", parts[2], expectedHeading)
		}

		// Line 4: blank.
		if parts[3] != "" {
			rt.Fatalf("line 4: expected blank, got %q", parts[3])
		}

		// Remainder: the response string.
		if parts[4] != response {
			rt.Fatalf("remainder: got %q, want %q", parts[4], response)
		}
	})
}

// Feature: code-persona, Property 5: UUID uniqueness
// **Validates: Requirements 6.1**
func TestProperty_CodeUUIDUniqueness(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Draw a random batch size between 1 and 1000.
		n := rapid.IntRange(1, 1000).Draw(rt, "batchSize")

		seen := make(map[string]struct{}, n)
		for i := 0; i < n; i++ {
			uid := generateRecommendationID()
			if _, exists := seen[uid]; exists {
				rt.Fatalf("duplicate UUID detected at index %d: %s (batch size %d)", i, uid, n)
			}
			seen[uid] = struct{}{}
		}
	})
}

// Feature: code-persona, Property 6: Code feedback prompt correctness
// **Validates: Requirements 7.1, 7.6**
func TestProperty_CodeFeedbackPrompt(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Draw a random non-negative code block count.
		codeBlockCount := rapid.IntRange(0, 50).Draw(rt, "codeBlockCount")

		prompt := buildCodeFeedbackPrompt(codeBlockCount)

		// Property: the prompt must always contain [s], [n], and [x].
		if !strings.Contains(prompt, "[s]") {
			rt.Fatalf("buildCodeFeedbackPrompt(%d) = %q does not contain [s]", codeBlockCount, prompt)
		}
		if !strings.Contains(prompt, "[n]") {
			rt.Fatalf("buildCodeFeedbackPrompt(%d) = %q does not contain [n]", codeBlockCount, prompt)
		}
		if !strings.Contains(prompt, "[x]") {
			rt.Fatalf("buildCodeFeedbackPrompt(%d) = %q does not contain [x]", codeBlockCount, prompt)
		}

		// Property: when codeBlockCount > 1, numbered selectors [1]..[N] must be present.
		if codeBlockCount > 1 {
			for i := 1; i <= codeBlockCount; i++ {
				selector := fmt.Sprintf("[%d]", i)
				if !strings.Contains(prompt, selector) {
					rt.Fatalf("buildCodeFeedbackPrompt(%d) = %q does not contain %s",
						codeBlockCount, prompt, selector)
				}
			}
		}
	})
}

// Feature: code-persona, Property 7: Non-feedback input rejection
// **Validates: Requirements 7.5**
func TestProperty_CodeNonFeedbackRejection(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Draw a random code block count (0 means no numbered selectors are valid).
		codeBlockCount := rapid.IntRange(0, 20).Draw(rt, "codeBlockCount")

		// Draw an arbitrary string and filter out all valid feedback inputs.
		input := rapid.String().Filter(func(s string) bool {
			// Exclude "s" and "n".
			if s == "s" || s == "n" {
				return false
			}
			// Exclude anything starting with "x".
			if strings.HasPrefix(s, "x") {
				return false
			}
			// Exclude valid integers when code blocks exist.
			if codeBlockCount > 0 {
				if _, err := strconv.Atoi(s); err == nil {
					return false
				}
			}
			return true
		}).Draw(rt, "input")

		// Property: any string that is not a valid feedback input must be rejected.
		if isCodeFeedbackInput(input, codeBlockCount) {
			rt.Fatalf("isCodeFeedbackInput(%q, %d) = true, want false for non-feedback input",
				input, codeBlockCount)
		}
	})
}

// Feature: code-persona, Property 9: Backup filename format
// **Validates: Requirements 11.3**
func TestProperty_BackupFilenameFormat(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a simple alphanumeric filename with extension.
		baseName := rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9]{0,7}`).Draw(rt, "baseName")
		ext := rapid.SampledFrom([]string{".sh", ".py", ".go", ".js", ".rb", ".txt"}).Draw(rt, "ext")
		originalName := baseName + ext

		// Generate a random time.Time by drawing a random Unix timestamp.
		unixSec := rapid.Int64Range(0, 253402300799).Draw(rt, "unixSec")
		ts := time.Unix(unixSec, 0)

		filename := buildBackupFilename(originalName, ts)

		// Property: the filename must match backup-<originalName>-YYYY-MM-DD_HH-mm-SS
		expectedPattern := regexp.MustCompile(
			`^backup-` + regexp.QuoteMeta(originalName) + `-\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}$`,
		)
		if !expectedPattern.MatchString(filename) {
			rt.Fatalf("buildBackupFilename(%q, %v) = %q does not match pattern %s",
				originalName, ts, filename, expectedPattern.String())
		}

		// Also verify the timestamp suffix is present and well-formed.
		if !backupFilenameTimestampPattern.MatchString(filename) {
			rt.Fatalf("buildBackupFilename(%q, %v) = %q missing valid timestamp suffix",
				originalName, ts, filename)
		}
	})
}

// Feature: code-persona, Property 8: Script backup round-trip
// **Validates: Requirements 11.2, 11.9**
func TestProperty_ScriptBackupRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	rapid.Check(t, func(rt *rapid.T) {
		// Generate random file content.
		content := rapid.StringMatching(`[a-zA-Z0-9 \n!@#$%^&*()]{1,500}`).Draw(rt, "content")

		// Write to a temp file.
		scriptName := "test-script.sh"
		scriptPath := filepath.Join(tmpDir, scriptName)
		if err := os.WriteFile(scriptPath, []byte(content), 0644); err != nil {
			rt.Fatalf("failed to write script: %v", err)
		}

		// Create backup.
		backupPath, err := createScriptBackup(scriptPath)
		if err != nil {
			rt.Fatalf("createScriptBackup failed: %v", err)
		}

		// Read backup and verify content matches.
		backupContent, err := os.ReadFile(backupPath)
		if err != nil {
			rt.Fatalf("failed to read backup: %v", err)
		}

		if string(backupContent) != content {
			rt.Fatalf("backup content mismatch:\n  original: %q\n  backup:   %q", content, string(backupContent))
		}

		// Clean up backup file.
		os.Remove(backupPath)
	})
}

// Feature: code-persona, Property 10: Fix attempt counter behavior
// **Validates: Requirements 12.1, 12.2, 12.8**
func TestProperty_FixAttemptCounter(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Reset counter at start of each test (simulates new fix session).
		resetFixAttemptCounter()

		// Verify counter starts at 0.
		if fixAttemptCounter != 0 {
			rt.Fatalf("counter should be 0 after reset, got %d", fixAttemptCounter)
		}

		// Generate a random number of failures (1-10).
		numFailures := rapid.IntRange(1, 10).Draw(rt, "numFailures")

		for i := 0; i < numFailures; i++ {
			count := incrementFixAttemptCounter()

			// Property: counter equals number of increments.
			if count != i+1 {
				rt.Fatalf("after %d increments, counter = %d, want %d", i+1, count, i+1)
			}

			// Property: shouldProposeFallback is true exactly when counter >= maxFixAttempts.
			shouldFallback := shouldProposeFallback()
			if count >= maxFixAttempts && !shouldFallback {
				rt.Fatalf("counter = %d >= %d but shouldProposeFallback = false", count, maxFixAttempts)
			}
			if count < maxFixAttempts && shouldFallback {
				rt.Fatalf("counter = %d < %d but shouldProposeFallback = true", count, maxFixAttempts)
			}
		}

		// Property: reset brings counter back to 0.
		resetFixAttemptCounter()
		if fixAttemptCounter != 0 {
			rt.Fatalf("counter should be 0 after second reset, got %d", fixAttemptCounter)
		}
		if shouldProposeFallback() {
			rt.Fatal("shouldProposeFallback should be false after reset")
		}
	})
}
