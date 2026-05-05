package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gcclinux/scmd/internal/ai"
	"github.com/gcclinux/scmd/internal/ai/gemini"
	"github.com/gcclinux/scmd/internal/ai/ollama"
	"github.com/gcclinux/scmd/internal/database"
	"github.com/gcclinux/scmd/internal/markdown"
)

const codeTimestampLayout = "2006-01-02_15-04-05"
const codeSessionPrefix = "scmd-code-"
const backupPrefix = "backup-"
const queryMatchThreshold = 0.70
const maxFixAttempts = 3

// buildCodeSessionFilename returns scmd-code-YYYY-MM-DD_HH-mm-SS.md for the given time.
func buildCodeSessionFilename(t time.Time) string {
	return codeSessionPrefix + t.Format(codeTimestampLayout) + ".md"
}

// buildBackupFilename returns backup-<originalName>-YYYY-MM-DD_HH-mm-SS for the given time.
func buildBackupFilename(originalName string, t time.Time) string {
	return backupPrefix + originalName + "-" + t.Format(codeTimestampLayout)
}

// assembleCodeSessionContent produces the full session file content:
//
//	<!-- SCMD-ID: <uid> -->
//
//	# <query>
//
//	<response body>
func assembleCodeSessionContent(uid, query, response string) string {
	return formatIDHeader(uid) + "\n\n# " + query + "\n\n" + response
}

// computeWordOverlap computes the word-overlap ratio between query and content.
// Both are lowercased and split on whitespace. Returns (matched query words / total query words).
func computeWordOverlap(query, content string) float64 {
	queryWords := strings.Fields(strings.ToLower(query))
	if len(queryWords) == 0 {
		return 0.0
	}

	contentLower := strings.ToLower(content)
	matched := 0
	for _, word := range queryWords {
		if strings.Contains(contentLower, word) {
			matched++
		}
	}

	return float64(matched) / float64(len(queryWords))
}

// isRelevantMatch returns true if the word-overlap ratio meets or exceeds the threshold (70%).
func isRelevantMatch(query, content string) bool {
	return computeWordOverlap(query, content) >= queryMatchThreshold
}

// buildCodeFeedbackPrompt returns the feedback prompt with [s], [n], [x] options.
// When codeBlockCount > 1, numbered block selectors [1]..[N] are added.
func buildCodeFeedbackPrompt(codeBlockCount int) string {
	prompt := "[s] - Good answer (saves to db)  |  [n] - New answer (discards)  |  [x] Execute"
	if codeBlockCount > 1 {
		prompt += ":"
		for i := 1; i <= codeBlockCount; i++ {
			prompt += fmt.Sprintf(" [%d]", i)
		}
	}
	return prompt
}

// isCodeFeedbackInput returns true for valid code feedback commands:
// "s", "n", any input starting with "x", or a bare number when code blocks exist.
func isCodeFeedbackInput(input string, codeBlockCount int) bool {
	if input == "s" || input == "n" {
		return true
	}
	if strings.HasPrefix(input, "x") {
		return true
	}
	if codeBlockCount > 0 {
		if _, err := strconv.Atoi(input); err == nil {
			return true
		}
	}
	return false
}

// buildSearchCommand constructs the appropriate exec.Cmd for searching cache files
// based on the current OS. Uses findstr on Windows, grep on Linux/macOS.
func buildSearchCommand(cacheDir string, queryTerms []string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		// findstr /S /I /M /C:"term1" /C:"term2" cacheDir\*.md
		args := []string{"/S", "/I", "/M"}
		for _, term := range queryTerms {
			args = append(args, "/C:"+term)
		}
		args = append(args, filepath.Join(cacheDir, "*.md"))
		return exec.Command("findstr", args...)
	}
	// grep -r -l -i -e "term1" -e "term2" cacheDir/
	args := []string{"-r", "-l", "-i"}
	for _, term := range queryTerms {
		args = append(args, "-e", term)
	}
	args = append(args, cacheDir)
	return exec.Command("grep", args...)
}

// scanCacheFiles searches Cache_Directory for files matching query terms using OS-native commands.
// Returns list of matching file paths. Handles non-existent or empty cache directory gracefully.
func scanCacheFiles(query string) ([]string, error) {
	dir, err := ensureCacheDir()
	if err != nil {
		return nil, nil // Skip READ phase without error
	}

	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		return nil, nil // Skip READ phase
	}

	queryTerms := strings.Fields(query)
	if len(queryTerms) == 0 {
		return nil, nil
	}

	cmd := buildSearchCommand(dir, queryTerms)
	output, err := cmd.Output()
	if err != nil {
		// grep/findstr returns exit code 1 when no matches found — not an error
		return nil, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var matches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			matches = append(matches, line)
		}
	}

	return matches, nil
}

// validateTargetPath checks that path exists, is a directory, and is writable.
// Returns (valid bool, reason string).
func validateTargetPath(path string) (bool, string) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "path does not exist"
		}
		return false, fmt.Sprintf("cannot access path: %v", err)
	}

	if !info.IsDir() {
		return false, "path is not a directory"
	}

	// Check write permission by attempting to create a temp file
	testFile := filepath.Join(path, ".scmd-write-test")
	f, err := os.Create(testFile)
	if err != nil {
		return false, "permission denied: cannot write to directory"
	}
	f.Close()
	os.Remove(testFile)

	return true, ""
}

// createScriptBackup copies the original script to Cache_Directory with a backup filename.
// Returns the backup path.
func createScriptBackup(scriptPath string) (string, error) {
	dir, err := ensureCacheDir()
	if err != nil {
		return "", err
	}

	originalName := filepath.Base(scriptPath)
	backupName := buildBackupFilename(originalName, time.Now())
	backupPath := filepath.Join(dir, backupName)

	src, err := os.Open(scriptPath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return backupPath, nil
}

// detectCurrentProvider returns "ollama" or "gemini" based on which provider is available.
// Checks the AGENT env var first, then falls back to availability checks.
func detectCurrentProvider() string {
	preferred := strings.ToLower(os.Getenv("AGENT"))
	if preferred == "ollama" && ollama.IsAvailable() {
		return "ollama"
	}
	if preferred == "gemini" && gemini.IsAvailable() {
		return "gemini"
	}
	if ollama.IsAvailable() {
		return "ollama"
	}
	if gemini.IsAvailable() {
		return "gemini"
	}
	return ""
}

// getAlternativeProvider returns the opposite provider name.
func getAlternativeProvider(current string) string {
	if current == "ollama" {
		return "gemini"
	}
	return "ollama"
}

// proposeLLMFallback calls the alternative provider with full error history.
// Returns the new AI response or an error if the alternative is unavailable.
func proposeLLMFallback(current string, errorHistory []string) (string, error) {
	alt := getAlternativeProvider(current)

	prompt := "Previous fix attempts failed with these errors:\n"
	for i, e := range errorHistory {
		prompt += fmt.Sprintf("%d. %s\n", i+1, e)
	}
	prompt += "\nPlease provide a fresh analysis and fix."

	var results []database.CommandRecord

	if alt == "ollama" {
		if !ollama.IsAvailable() {
			return "", fmt.Errorf("alternative provider '%s' is not available", alt)
		}
		resp, _, err := ollama.Ask(prompt, results)
		if err != nil {
			return "", err
		}
		return resp, nil
	}

	if alt == "gemini" {
		if !gemini.IsAvailable() {
			return "", fmt.Errorf("alternative provider '%s' is not available", alt)
		}
		resp, _, err := gemini.Ask(prompt, results)
		if err != nil {
			return "", err
		}
		return resp, nil
	}

	return "", fmt.Errorf("no alternative provider available")
}

// fixAttemptCounter tracks the number of fix attempts in the current session.
var fixAttemptCounter int

// resetFixAttemptCounter resets the fix attempt counter for a new fix session.
func resetFixAttemptCounter() {
	fixAttemptCounter = 0
}

// incrementFixAttemptCounter increments the counter and returns the new value.
func incrementFixAttemptCounter() int {
	fixAttemptCounter++
	return fixAttemptCounter
}

// shouldProposeFallback returns true when the counter reaches maxFixAttempts.
func shouldProposeFallback() bool {
	return fixAttemptCounter >= maxFixAttempts
}

// handleFixWorkflow orchestrates the fix loop: backup → AI fix → user test → iterate.
func handleFixWorkflow(scriptPath string, errorDesc string, reader *bufio.Reader) string {
	resetFixAttemptCounter()

	// Verify script exists
	if _, err := os.Stat(scriptPath); err != nil {
		fmt.Printf("Error: cannot access script %s: %v\n", scriptPath, err)
		return ""
	}

	// Create backup
	backupPath, err := createScriptBackup(scriptPath)
	if err != nil {
		fmt.Printf("Error: cannot create backup: %v\n", err)
		fmt.Println("Fix aborted — will not modify file without backup.")
		return ""
	}
	fmt.Printf("📁 Backup created: %s\n", backupPath)

	// Read the original script content
	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		fmt.Printf("Error reading script: %v\n", err)
		return ""
	}

	var errorHistory []string
	errorHistory = append(errorHistory, errorDesc)

	for {
		// Build the fix prompt
		prompt := fmt.Sprintf("Fix this script:\n\nFile: %s\n\nScript content:\n%s\n\nError:\n%s",
			scriptPath, string(scriptContent), strings.Join(errorHistory, "\n"))

		// Call AI for fix
		results, _, _, _ := ai.SmartSearch(prompt, true)
		aiResp, _, aiErr := ai.AskAIPersona("code", prompt, results)
		if aiErr != nil {
			fmt.Printf("Error from AI: %v\n", aiErr)
			return ""
		}

		if aiResp == "" {
			fmt.Println("Failed to generate fix proposal")
			return ""
		}

		// Display the fix proposal
		fmt.Println()
		fmt.Println("══════════════════════════════════════════════════════════════")
		fmt.Print(markdown.Render(aiResp))
		fmt.Println("══════════════════════════════════════════════════════════════")
		fmt.Println()

		// Ask user what to do
		fmt.Println("Options:")
		fmt.Println("  [t] - Manually test the script")
		fmt.Println("  [r] - Run the script within SCMD for testing")
		fmt.Println("  [a] - Abandon fix")
		fmt.Print("Choice: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))

		switch choice {
		case "a":
			fmt.Println("Fix abandoned. Original backup at:", backupPath)
			return aiResp
		case "t":
			fmt.Println("Test the script manually, then tell me the result.")
			fmt.Print("Did the fix work? (y/n): ")
			result, _ := reader.ReadString('\n')
			result = strings.TrimSpace(strings.ToLower(result))

			if result == "y" || result == "yes" {
				fmt.Println("✓ Fix successful!")
				return aiResp
			}

			// Fix failed
			count := incrementFixAttemptCounter()
			fmt.Print("What error did you get? ")
			newError, _ := reader.ReadString('\n')
			newError = strings.TrimSpace(newError)
			errorHistory = append(errorHistory, newError)

			if shouldProposeFallback() {
				current := detectCurrentProvider()
				alt := getAlternativeProvider(current)
				fmt.Printf("\n⚠ %d fix attempts failed. Would you like to try %s for a fresh perspective? (y/n): ", count, alt)
				fallbackChoice, _ := reader.ReadString('\n')
				fallbackChoice = strings.TrimSpace(strings.ToLower(fallbackChoice))

				if fallbackChoice == "y" || fallbackChoice == "yes" {
					fallbackResp, fallbackErr := proposeLLMFallback(current, errorHistory)
					if fallbackErr != nil {
						fmt.Printf("⚠ %v\n", fallbackErr)
						fmt.Println("Suggest manual review of the script.")
						return aiResp
					}
					aiResp = fallbackResp
					fmt.Println()
					fmt.Println("══════════════════════════════════════════════════════════════")
					fmt.Printf("🔄 Response from %s:\n", alt)
					fmt.Print(markdown.Render(fallbackResp))
					fmt.Println("══════════════════════════════════════════════════════════════")
					fmt.Println()
					continue
				}

				fmt.Println("Returning to interactive prompt.")
				return aiResp
			}

		case "r":
			fmt.Println("Running script via SCMD...")
			handleRunCommand(scriptPath)

			fmt.Print("Did the fix work? (y/n): ")
			result, _ := reader.ReadString('\n')
			result = strings.TrimSpace(strings.ToLower(result))

			if result == "y" || result == "yes" {
				fmt.Println("✓ Fix successful!")
				return aiResp
			}

			count := incrementFixAttemptCounter()
			fmt.Print("What error did you get? ")
			newError, _ := reader.ReadString('\n')
			newError = strings.TrimSpace(newError)
			errorHistory = append(errorHistory, newError)

			if shouldProposeFallback() {
				current := detectCurrentProvider()
				alt := getAlternativeProvider(current)
				fmt.Printf("\n⚠ %d fix attempts failed. Would you like to try %s for a fresh perspective? (y/n): ", count, alt)
				fallbackChoice, _ := reader.ReadString('\n')
				fallbackChoice = strings.TrimSpace(strings.ToLower(fallbackChoice))

				if fallbackChoice == "y" || fallbackChoice == "yes" {
					fallbackResp, fallbackErr := proposeLLMFallback(current, errorHistory)
					if fallbackErr != nil {
						fmt.Printf("⚠ %v\n", fallbackErr)
						fmt.Println("Suggest manual review of the script.")
						return aiResp
					}
					aiResp = fallbackResp
					fmt.Println()
					fmt.Println("══════════════════════════════════════════════════════════════")
					fmt.Printf("🔄 Response from %s:\n", alt)
					fmt.Print(markdown.Render(fallbackResp))
					fmt.Println("══════════════════════════════════════════════════════════════")
					fmt.Println()
					continue
				}

				fmt.Println("Returning to interactive prompt.")
				return aiResp
			}

		default:
			fmt.Println("Invalid choice. Please enter [t], [r], or [a].")
		}
	}
}

// CodeBlock represents an extracted code block with its language tag and content.
type CodeBlock struct {
	Language string
	Content  string
}

// extractCodeBlocksWithLang extracts fenced code blocks preserving the language tag.
func extractCodeBlocksWithLang(text string) []CodeBlock {
	lines := strings.Split(text, "\n")
	var blocks []CodeBlock
	var current []string
	var lang string
	inside := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if inside {
				blocks = append(blocks, CodeBlock{
					Language: lang,
					Content:  strings.Join(current, "\n"),
				})
				current = nil
				lang = ""
				inside = false
			} else {
				inside = true
				lang = strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
				current = nil
			}
			continue
		}
		if inside {
			current = append(current, line)
		}
	}

	if len(blocks) == 0 {
		return nil
	}
	return blocks
}

// inferFilenameFromBlock tries to determine a filename for a code block
// based on its language tag and content. Returns a suggested filename.
func inferFilenameFromBlock(block CodeBlock, index int) string {
	// Map common language tags to extensions
	extMap := map[string]string{
		"python": ".py", "py": ".py",
		"bash": ".sh", "sh": ".sh", "shell": ".sh",
		"go": ".go", "golang": ".go",
		"javascript": ".js", "js": ".js",
		"typescript": ".ts", "ts": ".ts",
		"ruby": ".rb", "rb": ".rb",
		"rust": ".rs", "rs": ".rs",
		"java": ".java",
		"c":    ".c", "cpp": ".cpp", "c++": ".cpp",
		"yaml": ".yaml", "yml": ".yaml",
		"json":       ".json",
		"toml":       ".toml",
		"sql":        ".sql",
		"html":       ".html",
		"css":        ".css",
		"powershell": ".ps1", "ps1": ".ps1",
		"bat": ".bat", "batch": ".bat", "cmd": ".cmd",
		"dockerfile": "Dockerfile",
		"makefile":   "Makefile",
		"txt":        ".txt", "text": ".txt",
		"ini": ".ini", "conf": ".conf",
	}

	lang := strings.ToLower(block.Language)

	// Check for special filenames
	if lang == "dockerfile" {
		return "Dockerfile"
	}
	if lang == "makefile" {
		return "Makefile"
	}

	// Check for requirements.txt pattern
	firstLine := ""
	if lines := strings.SplitN(block.Content, "\n", 2); len(lines) > 0 {
		firstLine = strings.TrimSpace(lines[0])
	}
	if lang == "txt" || lang == "text" || lang == "" {
		if strings.Contains(firstLine, "==") || strings.Contains(firstLine, ">=") {
			return "requirements.txt"
		}
	}

	ext, ok := extMap[lang]
	if ok {
		return fmt.Sprintf("script_%d%s", index+1, ext)
	}

	// Fallback: use .txt
	return fmt.Sprintf("code_block_%d.txt", index+1)
}

// handleCodeExecute implements the EXECUTE phase of the /code workflow.
// It asks the user for a target directory, validates it, shows the files
// that will be created, asks for confirmation, and writes them.
func handleCodeExecute(aiResponse string, codeBlocks []string, blockIndex int, reader *bufio.Reader) {
	// Extract blocks with language info from the original AI response
	richBlocks := extractCodeBlocksWithLang(aiResponse)

	// Determine which blocks to write
	var blocksToWrite []CodeBlock
	if blockIndex >= 0 && blockIndex < len(richBlocks) {
		// User selected a specific block
		blocksToWrite = []CodeBlock{richBlocks[blockIndex]}
	} else if len(richBlocks) == 1 {
		blocksToWrite = richBlocks
	} else {
		// All blocks
		blocksToWrite = richBlocks
	}

	if len(blocksToWrite) == 0 {
		fmt.Println("No code blocks found to create files from.")
		return
	}

	// Infer filenames
	type fileEntry struct {
		filename string
		content  string
	}
	files := make([]fileEntry, len(blocksToWrite))
	for i, block := range blocksToWrite {
		files[i] = fileEntry{
			filename: inferFilenameFromBlock(block, i),
			content:  block.Content,
		}
	}

	// Show what will be created
	fmt.Println()
	fmt.Println("📝 Files to create:")
	for _, f := range files {
		lineCount := strings.Count(f.content, "\n") + 1
		fmt.Printf("   • %s (%d lines)\n", f.filename, lineCount)
	}
	fmt.Println()

	// Ask for target directory
	fmt.Print("Target directory (or press Enter for current dir): ")
	targetDir, _ := reader.ReadString('\n')
	targetDir = strings.TrimSpace(targetDir)
	if targetDir == "" {
		var err error
		targetDir, err = os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}
	}

	// Expand ~ to home directory
	if strings.HasPrefix(targetDir, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			targetDir = filepath.Join(home, targetDir[1:])
		}
	}

	// Validate the target path
	valid, reason := validateTargetPath(targetDir)
	if !valid {
		fmt.Printf("⚠ Invalid target: %s\n", reason)
		fmt.Print("Enter a valid directory path: ")
		targetDir, _ = reader.ReadString('\n')
		targetDir = strings.TrimSpace(targetDir)
		if targetDir == "" {
			fmt.Println("Cancelled.")
			return
		}
		if strings.HasPrefix(targetDir, "~") {
			home, err := os.UserHomeDir()
			if err == nil {
				targetDir = filepath.Join(home, targetDir[1:])
			}
		}
		valid, reason = validateTargetPath(targetDir)
		if !valid {
			fmt.Printf("⚠ Still invalid: %s. Aborting.\n", reason)
			return
		}
	}

	// Allow user to rename files
	fmt.Println()
	fmt.Println("You can rename files before creation. Press Enter to keep the suggested name.")
	for i := range files {
		fmt.Printf("   [%d] %s → ", i+1, files[i].filename)
		newName, _ := reader.ReadString('\n')
		newName = strings.TrimSpace(newName)
		if newName != "" {
			files[i].filename = newName
		}
	}

	// Show final plan and confirm
	fmt.Println()
	fmt.Printf("📁 Will create in: %s\n", targetDir)
	for _, f := range files {
		fmt.Printf("   • %s\n", f.filename)
	}
	fmt.Print("\nProceed? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))
	if confirm != "y" && confirm != "yes" {
		fmt.Println("Cancelled. No files were created.")
		return
	}

	// Write files
	fmt.Println()
	for _, f := range files {
		fullPath := filepath.Join(targetDir, f.filename)

		// Check if file already exists
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("   ⚠ %s already exists. Overwrite? (y/n): ", f.filename)
			overwrite, _ := reader.ReadString('\n')
			overwrite = strings.TrimSpace(strings.ToLower(overwrite))
			if overwrite != "y" && overwrite != "yes" {
				fmt.Printf("   ⏭ Skipped %s\n", f.filename)
				continue
			}
		}

		if err := os.WriteFile(fullPath, []byte(f.content+"\n"), 0644); err != nil {
			fmt.Printf("   ✗ Failed to create %s: %v\n", f.filename, err)
		} else {
			fmt.Printf("   ✓ Created %s\n", fullPath)
		}
	}
	fmt.Println()
}

// handleCodeCommand is the main entry point for the /code command.
// It orchestrates READ→THINK→EXECUTE phases and returns the AI response,
// UUID, and session file path for the feedback loop.
func handleCodeCommand(args string) (string, string, string) {
	// Handle special case: /code load last workbook
	if strings.ToLower(strings.TrimSpace(args)) == "load last workbook" {
		wb, err := checkStartupWorkbook()
		if err != nil {
			fmt.Printf("Error loading workbook: %v\n", err)
			return "", "", ""
		}
		if wb == nil {
			fmt.Println("No unfinished workbooks found.")
			return "", "", ""
		}

		fmt.Printf("📋 Loaded workbook: %s\n", wb.Title)
		fmt.Printf("   Status: %s\n", wb.Status)
		fmt.Printf("   File: %s\n", wb.FilePath)
		fmt.Println()
		for _, step := range wb.Steps {
			fmt.Printf("   [%s] %d. %s\n", step.Status, step.Number, step.Description)
		}
		fmt.Println()

		idx, err := findFirstTodoStep(wb)
		if err != nil {
			fmt.Println("All steps are complete!")
			return "", "", ""
		}
		fmt.Printf("Continuing from step %d: %s\n", wb.Steps[idx].Number, wb.Steps[idx].Description)
		return "", "", ""
	}

	// Handle special case: /code fix <path>
	if strings.HasPrefix(strings.ToLower(args), "fix ") {
		fixArgs := strings.TrimSpace(args[4:])
		parts := strings.SplitN(fixArgs, " ", 2)
		scriptPath := parts[0]
		errorDesc := ""
		if len(parts) > 1 {
			errorDesc = parts[1]
		}

		reader := bufio.NewReader(os.Stdin)
		if errorDesc == "" {
			fmt.Print("Describe the error: ")
			errorDesc, _ = reader.ReadString('\n')
			errorDesc = strings.TrimSpace(errorDesc)
		}

		resp := handleFixWorkflow(scriptPath, errorDesc, reader)
		return resp, "", ""
	}

	fmt.Println("🔨 Processing with code agent...")

	// READ Phase: Cache Scanning
	var cachedContext string
	matches, _ := scanCacheFiles(args)
	if len(matches) > 0 {
		for _, match := range matches {
			content, err := os.ReadFile(match)
			if err != nil {
				continue
			}
			if isRelevantMatch(args, string(content)) {
				fmt.Printf("📁 Found relevant cache file: %s\n", filepath.Base(match))
				fmt.Print("Use this research as context? (y/n): ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))

				if response == "y" || response == "yes" {
					cachedContext = string(content)
					fmt.Println("✓ Using cached research as context")
				}
				break
			}
		}
	}

	// THINK Phase: Proposal Generation
	var prompt string
	if cachedContext != "" {
		prompt = fmt.Sprintf("Context from previous research:\n%s\n\nNew request: %s", cachedContext, args)
	} else {
		prompt = args
	}

	results, _, _, searchErr := ai.SmartSearch(prompt, true)
	if searchErr != nil {
		fmt.Printf("Error searching: %v\n", searchErr)
	}

	aiResp, _, aiErr := ai.AskAIPersona("code", prompt, results)
	if aiErr != nil {
		fmt.Printf("Error from AI: %v\n", aiErr)
		return "", "", ""
	}

	if aiResp == "" {
		fmt.Println("Failed to generate proposal")
		return "", "", ""
	}

	// Generate UUID and session file
	uid := generateRecommendationID()
	content := assembleCodeSessionContent(uid, args, aiResp)

	var savedFilePath string
	dir, dirErr := ensureCacheDir()
	if dirErr != nil {
		fmt.Printf("⚠ Could not create cache directory: %v\n", dirErr)
	} else {
		filename := buildCodeSessionFilename(time.Now())
		written, writeErr := writeRecommendationFile(dir, filename, content)
		if writeErr != nil {
			fmt.Printf("⚠ Could not write session file: %v\n", writeErr)
		} else {
			savedFilePath = written
		}
	}

	// Display the proposal
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

	return aiResp, uid, savedFilePath
}
