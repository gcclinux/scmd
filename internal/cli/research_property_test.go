package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"pgregory.net/rapid"
)

// uuidV4Pattern matches a standard UUID v4 string:
// 8-4-4-4-12 hex digits, with version nibble = 4 and variant nibble in [89ab].
var uuidV4Pattern = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
)

// idHeaderPattern matches the full formatted ID header line with an embedded UUID v4.
var idHeaderPattern = regexp.MustCompile(
	`^<!-- SCMD-ID: [0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12} -->$`,
)

// drawUUIDv4 generates a random valid UUID v4 string using rapid generators.
func drawUUIDv4(rt *rapid.T) string {
	// Build each section of a UUID v4 with the correct constraints.
	hexChar := rapid.StringMatching(`[0-9a-f]`)

	section1 := ""
	for i := 0; i < 8; i++ {
		section1 += hexChar.Draw(rt, "s1")
	}

	section2 := ""
	for i := 0; i < 4; i++ {
		section2 += hexChar.Draw(rt, "s2")
	}

	// Section 3: version nibble is always '4', followed by 3 hex chars.
	section3 := "4"
	for i := 0; i < 3; i++ {
		section3 += hexChar.Draw(rt, "s3")
	}

	// Section 4: variant nibble is one of [89ab], followed by 3 hex chars.
	variantNibble := rapid.SampledFrom([]string{"8", "9", "a", "b"}).Draw(rt, "variant")
	section4 := variantNibble
	for i := 0; i < 3; i++ {
		section4 += hexChar.Draw(rt, "s4")
	}

	section5 := ""
	for i := 0; i < 12; i++ {
		section5 += hexChar.Draw(rt, "s5")
	}

	return section1 + "-" + section2 + "-" + section3 + "-" + section4 + "-" + section5
}

// Feature: research-persona, Property 2: UUID header formatting
// **Validates: Requirements 4.1, 4.2**
func TestProperty_UUIDHeaderFormat(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		uid := drawUUIDv4(rt)

		header := formatIDHeader(uid)

		// Property: the header must match the exact pattern <!-- SCMD-ID: <uuid-v4> -->
		if !idHeaderPattern.MatchString(header) {
			rt.Fatalf("formatIDHeader(%q) = %q does not match expected pattern %s",
				uid, header, idHeaderPattern.String())
		}

		// Property: the header must contain the original UUID verbatim.
		expected := "<!-- SCMD-ID: " + uid + " -->"
		if header != expected {
			rt.Fatalf("formatIDHeader(%q) = %q, want %q", uid, header, expected)
		}
	})
}

// Feature: research-persona, Property 3: UUID uniqueness
// **Validates: Requirements 4.3**
func TestProperty_UUIDUniqueness(t *testing.T) {
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

// timestampFilenamePattern matches the expected recommendation filename format:
// scmd-YYYY-MM-DD_HH-mm-SS.md with zero-padded 24-hour time components.
var timestampFilenamePattern = regexp.MustCompile(
	`^scmd-\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}\.md$`,
)

// Feature: research-persona, Property 4: Timestamp filename format
// **Validates: Requirements 6.1, 9.1**
func TestProperty_TimestampFilenameFormat(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random time.Time by drawing a random Unix timestamp.
		// Range: 0 (1970-01-01) to 253402300799 (9999-12-31 23:59:59 UTC).
		// This covers all 4-digit years, which is the practical range.
		unixSec := rapid.Int64Range(0, 253402300799).Draw(rt, "unixSec")
		ts := time.Unix(unixSec, 0)

		filename := buildRecommendationFilename(ts)

		// Property: the filename must match the strict pattern scmd-YYYY-MM-DD_HH-mm-SS.md
		if !timestampFilenamePattern.MatchString(filename) {
			rt.Fatalf("buildRecommendationFilename(%v) = %q does not match pattern %s",
				ts, filename, timestampFilenamePattern.String())
		}
	})
}

// Feature: research-persona, Property 5: Timestamp round-trip
// **Validates: Requirements 9.2**
func TestProperty_TimestampRoundTrip(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random Unix timestamp in the range [0, 9999-12-31 23:59:59 UTC].
		unixSec := rapid.Int64Range(0, 253402300799).Draw(rt, "unixSec")
		ts := time.Unix(unixSec, 0)

		// Build the recommendation filename from the timestamp.
		filename := buildRecommendationFilename(ts)

		// Strip the "scmd-" prefix and ".md" suffix to extract the timestamp portion.
		timestampStr := filename[len("scmd-") : len(filename)-len(".md")]

		// Parse the timestamp back using the same location as the original time.
		// buildRecommendationFilename formats in the local timezone of ts, so we
		// must parse in the same location to get an equivalent time value.
		parsed, err := time.ParseInLocation(researchTimestampLayout, timestampStr, ts.Location())
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

// Feature: research-persona, Property 6: Recommendation file content structure
// **Validates: Requirements 6.2**
func TestProperty_RecommendationContentStructure(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random UUID v4 using the existing helper.
		uid := drawUUIDv4(rt)

		// Generate a random non-empty query string (no newlines).
		query := rapid.StringMatching(`[a-zA-Z0-9 ]+`).Draw(rt, "query")

		// Generate a random non-empty recommendation string (may contain newlines).
		recommendation := rapid.StringMatching(`[a-zA-Z0-9 \n]+`).Draw(rt, "recommendation")

		// Assemble the full recommendation file content.
		content := assembleRecommendationContent(uid, query, recommendation)

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

		// Remainder: the recommendation string.
		if parts[4] != recommendation {
			rt.Fatalf("remainder: got %q, want %q", parts[4], recommendation)
		}
	})
}

// numericBlockSelectorPattern matches any numeric block selector like [0], [1], [2], etc.
var numericBlockSelectorPattern = regexp.MustCompile(`\[\d\]`)

// Feature: research-persona, Property 8: Research feedback prompt correctness
// **Validates: Requirements 8.1, 8.2**
func TestProperty_ResearchFeedbackPrompt(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		prompt := buildResearchFeedbackPrompt()

		// Property: the prompt must contain [s] and [n].
		if !strings.Contains(prompt, "[s]") {
			rt.Fatalf("buildResearchFeedbackPrompt() = %q does not contain [s]", prompt)
		}
		if !strings.Contains(prompt, "[n]") {
			rt.Fatalf("buildResearchFeedbackPrompt() = %q does not contain [n]", prompt)
		}

		// Property: the prompt must NOT contain [x].
		if strings.Contains(prompt, "[x]") {
			rt.Fatalf("buildResearchFeedbackPrompt() = %q must not contain [x]", prompt)
		}

		// Property: the prompt must NOT contain "Execute".
		if strings.Contains(prompt, "Execute") {
			rt.Fatalf("buildResearchFeedbackPrompt() = %q must not contain 'Execute'", prompt)
		}

		// Property: the prompt must NOT contain any numeric block selector [<digit>].
		if numericBlockSelectorPattern.MatchString(prompt) {
			rt.Fatalf("buildResearchFeedbackPrompt() = %q must not contain numeric block selectors like [0], [1], etc.", prompt)
		}
	})
}

// Feature: research-persona, Property 9: Non-feedback input rejection
// **Validates: Requirements 8.5**
func TestProperty_ResearchNonFeedbackRejection(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Draw an arbitrary string and filter out the two valid feedback inputs.
		input := rapid.String().Filter(func(s string) bool {
			return s != "s" && s != "n"
		}).Draw(rt, "input")

		// Property: any string that is not "s" or "n" must be rejected.
		if isResearchFeedbackInput(input) {
			rt.Fatalf("isResearchFeedbackInput(%q) = true, want false for non-feedback input", input)
		}
	})
}

// Feature: research-persona, Property 1: Query parsing separates file path from text
// **Validates: Requirements 3.3**
func TestProperty_ResearchQueryParsing(t *testing.T) {
	// Create a temp directory once using the outer *testing.T (rapid.T has no TempDir).
	tmpDir := t.TempDir()

	rapid.Check(t, func(rt *rapid.T) {
		// Generate random file content (non-empty, alphanumeric with spaces).
		fileContent := rapid.StringMatching("[a-zA-Z0-9 ]+").Draw(rt, "fileContent")

		// Write the content to a temp file with a simple name.
		tmpFile := filepath.Join(tmpDir, "testfile.txt")
		err := os.WriteFile(tmpFile, []byte(fileContent), 0644)
		if err != nil {
			rt.Fatalf("failed to write temp file: %v", err)
		}

		// Generate random query text (starts with a letter, no path-like characters).
		queryText := rapid.StringMatching("[a-zA-Z][a-zA-Z0-9 ]*").Draw(rt, "queryText")

		// Combine query text and file path as args.
		args := queryText + " " + tmpFile

		// Call parseResearchQuery.
		query, content, fpath, parseErr := parseResearchQuery(args)

		// Assert: no error.
		if parseErr != nil {
			rt.Fatalf("parseResearchQuery(%q) returned error: %v", args, parseErr)
		}

		// Assert: filePath matches the temp file path.
		if fpath != tmpFile {
			rt.Fatalf("parseResearchQuery(%q): filePath = %q, want %q", args, fpath, tmpFile)
		}

		// Assert: fileContent matches the written content.
		if content != fileContent {
			rt.Fatalf("parseResearchQuery(%q): fileContent = %q, want %q", args, content, fileContent)
		}

		// Assert: query matches the query text after whitespace normalization.
		// parseResearchQuery uses strings.Fields internally, which collapses
		// multiple spaces into one, so we normalize both sides for comparison.
		normalizeWS := func(s string) string {
			return strings.Join(strings.Fields(s), " ")
		}
		if normalizeWS(query) != normalizeWS(queryText) {
			rt.Fatalf("parseResearchQuery(%q): query = %q, want %q", args, query, queryText)
		}
	})
}

// Feature: research-persona, Property 7: Recommendation file round-trip
// **Validates: Requirements 6.5**
func TestProperty_RecommendationFileRoundTrip(t *testing.T) {
	// Create a temp directory once using the outer *testing.T.
	tmpDir := t.TempDir()

	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random UUID v4.
		uid := drawUUIDv4(rt)

		// Generate a random query string (no newlines).
		query := rapid.StringMatching(`[a-zA-Z0-9 ]+`).Draw(rt, "query")

		// Generate a random recommendation string (may contain newlines).
		recommendation := rapid.StringMatching(`[a-zA-Z0-9 \n]+`).Draw(rt, "recommendation")

		// Assemble the full recommendation file content.
		content := assembleRecommendationContent(uid, query, recommendation)

		// Write the content to a temp file via writeRecommendationFile.
		writtenPath, err := writeRecommendationFile(tmpDir, "test.md", content)
		if err != nil {
			rt.Fatalf("writeRecommendationFile failed: %v", err)
		}

		// Read the file back.
		readBack, err := os.ReadFile(writtenPath)
		if err != nil {
			rt.Fatalf("os.ReadFile(%q) failed: %v", writtenPath, err)
		}

		// Property: the read content must be identical to the assembled content.
		if string(readBack) != content {
			rt.Fatalf("round-trip mismatch:\n  written: %q\n  read:    %q", content, string(readBack))
		}
	})
}
