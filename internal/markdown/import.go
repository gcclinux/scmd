package markdown

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gcclinux/scmd/internal/database"
)

// IsMarkdownFile checks if the file path has a .md extension.
func IsMarkdownFile(filePath string) bool {
	ext := filepath.Ext(filePath)
	return strings.EqualFold(ext, ".md")
}

// ExtractTitle parses the markdown content and returns the first top-level heading.
func ExtractTitle(content string, filePath string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# ") {
			title := strings.TrimPrefix(line, "# ")
			title = strings.TrimSpace(title)
			if title != "" {
				return title
			}
		}
	}
	base := filepath.Base(filePath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// ImportMarkdown reads a markdown file, validates it, extracts the title,
// checks for duplicates, and stores it via database.AddCommand.
// embeddingFn is an optional callback to generate embeddings.
func ImportMarkdown(filePath string, embeddingFn func(string) ([]float64, error)) (string, error) {
	if !IsMarkdownFile(filePath) {
		return "", fmt.Errorf("only markdown files (.md) are supported")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	content := string(data)
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("file is empty: %s", filePath)
	}

	title := ExtractTitle(content, filePath)

	exists, err := database.CheckCommandExists(title)
	if err != nil {
		return "", fmt.Errorf("error checking for duplicate: %v", err)
	}
	if exists {
		return "", fmt.Errorf("document already exists with title: %s", title)
	}

	success, err := database.AddCommand(title, content, embeddingFn)
	if err != nil {
		return "", fmt.Errorf("error storing document: %v", err)
	}
	if !success {
		return "", fmt.Errorf("failed to store document")
	}

	return title, nil
}
