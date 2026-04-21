package cli

import (
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// proseGen generates a random prose string that does NOT contain any line
// starting with triple backticks. This ensures the prose cannot be confused
// with a code fence when constructing test input.
func proseGen() *rapid.Generator[string] {
	return rapid.Custom[string](func(t *rapid.T) string {
		numLines := rapid.IntRange(0, 5).Draw(t, "proseLines")
		lines := make([]string, numLines)
		for i := range lines {
			line := rapid.StringMatching(`[a-zA-Z0-9 .,!?]{0,60}`).Draw(t, "proseLine")
			// Ensure no line starts with triple backticks
			line = strings.TrimLeft(line, "`")
			lines[i] = line
		}
		return strings.Join(lines, "\n")
	})
}

// codeContentGen generates random code block content lines where no line
// starts with triple backticks.
func codeContentGen() *rapid.Generator[string] {
	return rapid.Custom[string](func(t *rapid.T) string {
		numLines := rapid.IntRange(0, 5).Draw(t, "contentLines")
		lines := make([]string, numLines)
		for i := range lines {
			line := rapid.StringMatching(`[a-zA-Z0-9 _\-=/$(){};:'",.!?]{0,80}`).Draw(t, "contentLine")
			// Ensure no line starts with triple backticks
			line = strings.TrimLeft(line, "`")
			lines[i] = line
		}
		return strings.Join(lines, "\n")
	})
}

// langTagGen generates a simple alphanumeric language tag (or empty string).
func langTagGen() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{
		"", "bash", "python", "go", "js", "sh", "ruby", "sql", "yaml", "json",
	})
}

// Feature: execute-command-option, Property 1: Code block extraction round-trip
// **Validates: Requirements 2.1, 2.2, 2.3**
func TestProperty_CodeBlockExtractionRoundTrip(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random number of code blocks (1 to 5).
		numBlocks := rapid.IntRange(1, 5).Draw(rt, "numBlocks")

		// Generate the code block contents and language tags.
		contents := make([]string, numBlocks)
		langTags := make([]string, numBlocks)
		for i := 0; i < numBlocks; i++ {
			contents[i] = codeContentGen().Draw(rt, "content")
			langTags[i] = langTagGen().Draw(rt, "langTag")
		}

		// Build the full text by interleaving prose with fenced code blocks.
		var builder strings.Builder

		// Leading prose before the first block.
		leadingProse := proseGen().Draw(rt, "leadingProse")
		if leadingProse != "" {
			builder.WriteString(leadingProse)
			builder.WriteString("\n")
		}

		for i := 0; i < numBlocks; i++ {
			// Opening fence with optional language tag.
			builder.WriteString("```")
			builder.WriteString(langTags[i])
			builder.WriteString("\n")

			// Code block content.
			if contents[i] != "" {
				builder.WriteString(contents[i])
				builder.WriteString("\n")
			}

			// Closing fence.
			builder.WriteString("```")
			builder.WriteString("\n")

			// Trailing prose after this block.
			trailingProse := proseGen().Draw(rt, "trailingProse")
			if trailingProse != "" {
				builder.WriteString(trailingProse)
				builder.WriteString("\n")
			}
		}

		fullText := builder.String()

		// Extract code blocks.
		got := ExtractCodeBlocks(fullText)

		// Verify the number of extracted blocks matches.
		if len(got) != numBlocks {
			rt.Fatalf("expected %d blocks, got %d\ntext:\n%s", numBlocks, len(got), fullText)
		}

		// Verify each extracted block matches the original content exactly.
		for i := 0; i < numBlocks; i++ {
			if got[i] != contents[i] {
				rt.Fatalf("block %d mismatch:\n  expected: %q\n  got:      %q\ntext:\n%s",
					i, contents[i], got[i], fullText)
			}
		}
	})
}
