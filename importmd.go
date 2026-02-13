package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// isMarkdownFile checks if the file path has a .md extension (case-insensitive).
func isMarkdownFile(filePath string) bool {
	ext := filepath.Ext(filePath)
	return strings.EqualFold(ext, ".md")
}

// extractTitle parses the markdown content and returns the first
// top-level heading (# Title). Falls back to the filename without extension.
func extractTitle(content string, filePath string) string {
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
	// Fallback: filename without extension
	base := filepath.Base(filePath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// ImportMarkdown reads a markdown file, validates it, extracts the title,
// checks for duplicates, and stores it via AddCommand.
// Returns the document title and an error if any step fails.
func ImportMarkdown(filePath string) (string, error) {
	// Validate .md extension
	if !isMarkdownFile(filePath) {
		return "", fmt.Errorf("only markdown files (.md) are supported")
	}

	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	content := string(data)

	// Check for empty file
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("file is empty: %s", filePath)
	}

	// Extract title from content
	title := extractTitle(content, filePath)

	// Check for duplicate key
	exists, err := CheckCommandExists(title)
	if err != nil {
		return "", fmt.Errorf("error checking for duplicate: %v", err)
	}
	if exists {
		return "", fmt.Errorf("document already exists with title: %s", title)
	}

	// Store via AddCommand (title as key, full content as data)
	success, err := AddCommand(title, content)
	if err != nil {
		return "", fmt.Errorf("error storing document: %v", err)
	}
	if !success {
		return "", fmt.Errorf("failed to store document")
	}

	return title, nil
}
