package main

import (
	"bufio"
	"fmt"
	"strings"
)

// ANSI escape codes for terminal styling.
const (
	ansiBold   = "\033[1m"
	ansiCyan   = "\033[36m"
	ansiYellow = "\033[33m"
	ansiGray   = "\033[90m"
	ansiReset  = "\033[0m"
)

// isMarkdownContent detects whether a string contains markdown formatting
// (fenced code blocks, headers, or markdown links) vs plain command text.
func isMarkdownContent(content string) bool {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Detect fenced code blocks
		if strings.HasPrefix(trimmed, "```") {
			return true
		}

		// Detect markdown headers (# , ## , ### , etc.)
		if strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "### ") {
			return true
		}
	}

	// Detect markdown links [text](url) anywhere in content
	if containsMarkdownLink(content) {
		return true
	}

	return false
}

// containsMarkdownLink checks if the content contains a markdown link pattern [text](url).
func containsMarkdownLink(content string) bool {
	i := 0
	for i < len(content) {
		// Find opening bracket
		openBracket := strings.Index(content[i:], "[")
		if openBracket == -1 {
			break
		}
		openBracket += i

		// Find closing bracket after opening
		closeBracket := strings.Index(content[openBracket+1:], "]")
		if closeBracket == -1 {
			break
		}
		closeBracket += openBracket + 1

		// Check that text between brackets is non-empty
		linkText := content[openBracket+1 : closeBracket]
		if len(linkText) == 0 {
			i = closeBracket + 1
			continue
		}

		// Check for immediate opening paren after closing bracket
		if closeBracket+1 < len(content) && content[closeBracket+1] == '(' {
			closeParen := strings.Index(content[closeBracket+2:], ")")
			if closeParen != -1 {
				url := content[closeBracket+2 : closeBracket+2+closeParen]
				if len(url) > 0 {
					return true
				}
			}
		}

		i = closeBracket + 1
	}
	return false
}

// RenderMarkdown formats markdown content for terminal display.
// Renders fenced code blocks with delimiters/labels and headers with ANSI styling.
func RenderMarkdown(content string) string {
	var out strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(content))

	inCodeBlock := false
	var codeBlockLang string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Handle fenced code block boundaries
		if strings.HasPrefix(trimmed, "```") {
			if !inCodeBlock {
				// Opening fence — extract language label
				codeBlockLang = strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
				if codeBlockLang != "" {
					out.WriteString(fmt.Sprintf("%s┌─ %s ─%s\n", ansiYellow, codeBlockLang, ansiReset))
				} else {
					out.WriteString(fmt.Sprintf("%s┌──────%s\n", ansiYellow, ansiReset))
				}
				inCodeBlock = true
			} else {
				// Closing fence
				if codeBlockLang != "" {
					out.WriteString(fmt.Sprintf("%s└─ %s ─%s\n", ansiYellow, codeBlockLang, ansiReset))
				} else {
					out.WriteString(fmt.Sprintf("%s└──────%s\n", ansiYellow, ansiReset))
				}
				inCodeBlock = false
				codeBlockLang = ""
			}
			continue
		}

		// Inside a code block — render with gray styling
		if inCodeBlock {
			out.WriteString(fmt.Sprintf("%s│%s %s\n", ansiYellow, ansiReset, line))
			continue
		}

		// Render headers with ANSI bold + cyan
		if strings.HasPrefix(trimmed, "### ") {
			headerText := strings.TrimPrefix(trimmed, "### ")
			out.WriteString(fmt.Sprintf("%s%s%s%s\n", ansiBold, ansiCyan, headerText, ansiReset))
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			headerText := strings.TrimPrefix(trimmed, "## ")
			out.WriteString(fmt.Sprintf("%s%s%s%s\n", ansiBold, ansiCyan, headerText, ansiReset))
			continue
		}
		if strings.HasPrefix(trimmed, "# ") {
			headerText := strings.TrimPrefix(trimmed, "# ")
			out.WriteString(fmt.Sprintf("%s%s%s%s\n", ansiBold, ansiCyan, headerText, ansiReset))
			continue
		}

		// Regular line — pass through as-is
		out.WriteString(line + "\n")
	}

	// If we ended inside an unclosed code block, close it
	if inCodeBlock {
		out.WriteString(fmt.Sprintf("%s└──────%s\n", ansiYellow, ansiReset))
	}

	return out.String()
}
