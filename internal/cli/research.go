package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gcclinux/scmd/internal/ai"
	"github.com/gcclinux/scmd/internal/markdown"
	"github.com/google/uuid"
)

// researchTimestampLayout is the Go reference-time layout used for
// recommendation filenames: scmd-YYYY-MM-DD_HH-mm-SS.md
const researchTimestampLayout = "2006-01-02_15-04-05"

// generateRecommendationID returns a new UUID v4 string for tagging a
// recommendation file.
func generateRecommendationID() string {
	return uuid.New().String()
}

// formatIDHeader wraps a UUID string in the HTML comment format used as
// the first line of every recommendation file.
func formatIDHeader(uid string) string {
	return fmt.Sprintf("<!-- SCMD-ID: %s -->", uid)
}

// buildRecommendationFilename returns a filename of the form
// scmd-YYYY-MM-DD_HH-mm-SS.md for the given time.
func buildRecommendationFilename(t time.Time) string {
	return "scmd-" + t.Format(researchTimestampLayout) + ".md"
}

// assembleRecommendationContent produces the full markdown file content:
//
//	<!-- SCMD-ID: <uid> -->
//
//	# <query>
//
//	<recommendation body>
func assembleRecommendationContent(uid, query, recommendation string) string {
	return formatIDHeader(uid) + "\n\n# " + query + "\n\n" + recommendation
}

// buildResearchFeedbackPrompt returns the fixed feedback prompt shown
// after a research recommendation. It offers only [s] save and [n] new.
func buildResearchFeedbackPrompt() string {
	return "[s] - Good answer (saves to db)  |  [n] - New answer (discards)"
}

// isResearchFeedbackInput returns true only for the two valid research
// feedback inputs: "s" (save) and "n" (new/regenerate).
func isResearchFeedbackInput(input string) bool {
	return input == "s" || input == "n"
}

// parseResearchQuery splits args into tokens and checks whether any token
// is a valid, readable file path. If a file is found, its contents are
// returned in fileContent and the remaining tokens form the query string.
// If no file token is detected, the entire args string is returned as the
// query with empty fileContent and filePath.
func parseResearchQuery(args string) (query, fileContent, filePath string, err error) {
	tokens := strings.Fields(args)
	if len(tokens) == 0 {
		return "", "", "", nil
	}

	fileIndex := -1
	for i, token := range tokens {
		info, statErr := os.Stat(token)
		if statErr != nil {
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		fileIndex = i
		break
	}

	if fileIndex == -1 {
		// No file path found — return the entire args as the query.
		return args, "", "", nil
	}

	filePath = tokens[fileIndex]

	data, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return "", "", "", readErr
	}
	fileContent = string(data)

	// Build the query from the remaining tokens (excluding the file path token).
	remaining := make([]string, 0, len(tokens)-1)
	for i, token := range tokens {
		if i != fileIndex {
			remaining = append(remaining, token)
		}
	}
	query = strings.Join(remaining, " ")

	return query, fileContent, filePath, nil
}

// ensureCacheDir returns the path ~/.scmd/cache/ after creating it (and any
// parent directories) with 0755 permissions if it does not already exist.
// If the directory cannot be created, the OS error is returned.
func ensureCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".scmd", "cache")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// writeRecommendationFile writes content to dir/filename and returns the
// full path on success, or an error if the write fails.
func writeRecommendationFile(dir, filename, content string) (string, error) {
	fullPath := filepath.Join(dir, filename)
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return "", err
	}
	return fullPath, nil
}

// handleResearchCommand is the main entry point for the /research command.
// It parses the query (optionally reading a file), calls the research AI
// persona, generates a UUID, writes the recommendation to the cache, and
// renders the output in the CLI. It returns the AI response, UUID, and
// saved file path for use by the feedback loop.
func handleResearchCommand(args string) (string, string, string) {
	fmt.Println("🔬 Processing with research persona...")

	// 1. Parse the query to separate file path from query text.
	query, fileContent, filePath, err := parseResearchQuery(args)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return "", "", ""
	}

	// 2. Build the AI prompt with file context if present.
	var prompt string
	if fileContent != "" {
		if query != "" {
			prompt = fmt.Sprintf("File: %s\n\nFile Contents:\n%s\n\nQuestion: %s", filePath, fileContent, query)
		} else {
			prompt = fmt.Sprintf("File: %s\n\nFile Contents:\n%s", filePath, fileContent)
		}
	} else {
		prompt = query
	}

	// 3. Perform a smart search to get context for the persona.
	results, _, _, searchErr := ai.SmartSearch(prompt, true)
	if searchErr != nil {
		fmt.Printf("Error searching: %v\n", searchErr)
	}

	// 4. Call the research persona.
	aiResp, _, aiErr := ai.AskAIPersona("research", prompt, results)
	if aiErr != nil {
		fmt.Printf("Error from AI: %v\n", aiErr)
		return "", "", ""
	}

	// 5. Handle empty AI response.
	if aiResp == "" {
		fmt.Println("Failed to generate recommendation")
		return "", "", ""
	}

	// 6. Generate UUID.
	uid := generateRecommendationID()

	// 7. Assemble file content (use original args as the query in the file).
	content := assembleRecommendationContent(uid, args, aiResp)

	// 8. Ensure cache directory exists.
	var savedFilePath string
	dir, dirErr := ensureCacheDir()
	if dirErr != nil {
		fmt.Printf("⚠ Could not create cache directory: %v\n", dirErr)
	} else {
		// 9. Write the recommendation file.
		written, writeErr := writeRecommendationFile(dir, buildRecommendationFilename(time.Now()), content)
		if writeErr != nil {
			fmt.Printf("⚠ Could not write recommendation file: %v\n", writeErr)
		} else {
			savedFilePath = written
		}
	}

	// 10. Display the recommendation in the CLI.
	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Printf("  ID: %s\n", uid)
	if savedFilePath != "" {
		fmt.Printf("  File: %s\n", savedFilePath)
	}
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Print(markdown.Render(aiResp))
	if !strings.HasSuffix(aiResp, "\n") {
		fmt.Println()
	}
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println()

	// 11. Return the AI response, UUID, and file path.
	return aiResp, uid, savedFilePath
}
