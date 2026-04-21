package cli

import "strings"

// ExtractCodeBlocks extracts all fenced code blocks from the given text.
// It looks for triple-backtick delimiters, strips the opening fence line
// (including any language tag) and the closing fence, and returns the
// inner content of each block as a string slice.
// Returns nil when no code blocks are found.
func ExtractCodeBlocks(text string) []string {
	lines := strings.Split(text, "\n")
	var blocks []string
	var current []string
	inside := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if inside {
				// Closing fence — finish the current block.
				blocks = append(blocks, strings.Join(current, "\n"))
				current = nil
				inside = false
			} else {
				// Opening fence — start collecting a new block.
				inside = true
				current = nil
			}
			continue
		}
		if inside {
			current = append(current, line)
		}
	}

	// If we're still inside a block when the text ends, the fence was
	// never closed. Discard the incomplete block.
	if len(blocks) == 0 {
		return nil
	}
	return blocks
}
