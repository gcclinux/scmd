package markdown

import (
	"bufio"
	"fmt"
	"strings"
)

// ANSI escape codes for terminal styling.
const (
	ansiBold   = "\033[1m"
	ansiCyan   = "\033[36m"
	ansiGreen  = "\033[32m"
	ansiReset  = "\033[0m"
)

// IsMarkdownContent detects whether a string contains markdown formatting.
func IsMarkdownContent(content string) bool {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			return true
		}
		if strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "### ") {
			return true
		}
	}

	if ContainsMarkdownLink(content) {
		return true
	}

	return false
}

// ContainsMarkdownLink checks if the content contains a markdown link pattern [text](url).
func ContainsMarkdownLink(content string) bool {
	i := 0
	for i < len(content) {
		openBracket := strings.Index(content[i:], "[")
		if openBracket == -1 {
			break
		}
		openBracket += i

		closeBracket := strings.Index(content[openBracket+1:], "]")
		if closeBracket == -1 {
			break
		}
		closeBracket += openBracket + 1

		linkText := content[openBracket+1 : closeBracket]
		if len(linkText) == 0 {
			i = closeBracket + 1
			continue
		}

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

// Render formats markdown content for terminal display.
func Render(content string) string {
	var out strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(content))

	inCodeBlock := false
	var codeBlockLang string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			if !inCodeBlock {
				codeBlockLang = strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
				if codeBlockLang != "" {
					out.WriteString(fmt.Sprintf("%s┌─ %s ─%s\n", ansiGreen, codeBlockLang, ansiReset))
				} else {
					out.WriteString(fmt.Sprintf("%s┌──────%s\n", ansiGreen, ansiReset))
				}
				inCodeBlock = true
			} else {
				if codeBlockLang != "" {
					out.WriteString(fmt.Sprintf("%s└─ %s ─%s\n", ansiGreen, codeBlockLang, ansiReset))
				} else {
					out.WriteString(fmt.Sprintf("%s└──────%s\n", ansiGreen, ansiReset))
				}
				inCodeBlock = false
				codeBlockLang = ""
			}
			continue
		}

		if inCodeBlock {
			out.WriteString(fmt.Sprintf("%s│%s %s\n", ansiGreen, ansiReset, line))
			continue
		}

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

		out.WriteString(line + "\n")
	}

	if inCodeBlock {
		out.WriteString(fmt.Sprintf("%s└──────%s\n", ansiGreen, ansiReset))
	}

	return out.String()
}
