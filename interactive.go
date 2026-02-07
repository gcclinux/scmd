package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// StartInteractiveMode starts the interactive CLI prompt
func StartInteractiveMode() {
	// Initialize database connection
	if err := InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer CloseDB()

	// Initialize embedding providers (Gemini first, then Ollama)
	InitGemini()
	InitOllama()

	reader := bufio.NewReader(os.Stdin)

	printWelcome()

	for {
		fmt.Print("scmd> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		// Handle exit commands
		if input == "exit" || input == "quit" || input == "q" {
			fmt.Println("Goodbye!")
			break
		}

		// Handle help command
		if input == "help" || input == "?" {
			printInteractiveHelp()
			continue
		}

		// Handle clear command
		if input == "clear" || input == "cls" {
			clearScreen()
			continue
		}

		// Process the command
		processInteractiveCommand(input)
	}
}

func printWelcome() {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║           SCMD Interactive CLI - PostgreSQL Edition            ║")
	fmt.Println("║                      Version", Release, "                            ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Show embedding provider status
	if IsGeminiAvailable() {
		fmt.Println("🚀 Gemini API: Active (vector search enabled)")
	} else if IsOllamaAvailable() {
		fmt.Println("🤖 Ollama: Active (vector search enabled)")
	} else {
		fmt.Println("⚠️  No embedding provider (traditional search only)")
	}
	fmt.Println()

	fmt.Println("Type '/help' or 'help' for available commands")
	fmt.Println("Type '/exit' or 'exit' to quit")
	fmt.Println()
	fmt.Println("Search Tips:")
	fmt.Println("  Space-separated = AND: postgresql replication slave")
	fmt.Println("  Comma-separated = OR:  docker,kubernetes,postgresql")
	fmt.Println("  Natural language:      show me postgresql replication examples")
	fmt.Println()
}

func printInteractiveHelp() {
	fmt.Println()
	fmt.Println("Available Slash Commands:")
	fmt.Println("──────────────────────────────────────────────────────────────")
	fmt.Println("  /help or /?           - Show this help message")
	fmt.Println("  /search <pattern>     - Search for commands matching pattern")
	fmt.Println("  /add <cmd> | <desc>   - Add a new command (use | as separator)")
	fmt.Println("  /delete <id>          - Delete a command by ID")
	fmt.Println("  /list                 - List recent commands")
	fmt.Println("  /count                - Show total number of commands")
	fmt.Println("  /ai                   - Show AI/Ollama status")
	fmt.Println("  /embeddings           - Check embedding statistics")
	fmt.Println("  /generate             - Generate embeddings for all commands")
	fmt.Println("  /clear or /cls        - Clear the screen")
	fmt.Println("  /exit, /quit, or /q   - Exit interactive mode")
	fmt.Println()
	fmt.Println("Quick Shortcuts (without slash):")
	fmt.Println("──────────────────────────────────────────────────────────────")
	fmt.Println("  help or ?             - Show this help message")
	fmt.Println("  clear or cls          - Clear the screen")
	fmt.Println("  exit, quit, or q      - Exit interactive mode")
	fmt.Println()
	if IsGeminiAvailable() || IsOllamaAvailable() {
		aiProvider := "Gemini"
		if !IsGeminiAvailable() && IsOllamaAvailable() {
			aiProvider = "Ollama"
		}
		fmt.Printf("🤖 AI Features (%s Active):\n", aiProvider)
		fmt.Println("──────────────────────────────────────────────────────────────")
		fmt.Println("  - Vector similarity search for better relevance")
		fmt.Println("  - AI-generated explanations and context")
		fmt.Println("  - Natural language understanding")
		fmt.Println("  - Automatic fallback to traditional search")
		fmt.Println()
	}
	fmt.Println("Natural Language Queries:")
	fmt.Println("──────────────────────────────────────────────────────────────")
	fmt.Println("  Just type your question naturally:")
	fmt.Println("  - show me docker commands")
	fmt.Println("  - how to check postgresql replication")
	fmt.Println("  - provide kubernetes examples")
	fmt.Println("  - find git commands")
	fmt.Println()
	fmt.Println("Search Logic:")
	fmt.Println("──────────────────────────────────────────────────────────────")
	fmt.Println("  Space-separated = AND (all words must match):")
	fmt.Println("    postgresql replication slave")
	fmt.Println("    → Finds records with ALL three words")
	fmt.Println()
	fmt.Println("  Comma-separated = OR (any pattern can match):")
	fmt.Println("    docker,kubernetes,postgresql")
	fmt.Println("    → Finds records with ANY of these words")
	fmt.Println()
	fmt.Println("  Combined:")
	fmt.Println("    postgresql replication,docker backup")
	fmt.Println("    → Finds (postgresql AND replication) OR (docker AND backup)")
	fmt.Println()
	fmt.Println("Direct Search:")
	fmt.Println("──────────────────────────────────────────────────────────────")
	fmt.Println("  Type keywords directly without /search:")
	fmt.Println("  - postgresql replication slave    (AND logic)")
	fmt.Println("  - docker,kubernetes               (OR logic)")
	fmt.Println()
}

func processInteractiveCommand(input string) {
	// Handle slash commands
	if strings.HasPrefix(input, "/") {
		handleSlashCommand(input)
		return
	}

	// Extract keywords from natural language or use direct input
	keywords := extractKeywords(input)

	if keywords == "" {
		fmt.Println("Could not extract search terms. Try using /search <pattern>")
		return
	}

	// Perform search
	performInteractiveSearch(keywords)
}

func handleSlashCommand(input string) {
	parts := strings.SplitN(input, " ", 2)
	command := parts[0]
	args := ""
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}

	switch command {
	case "/help", "/?":
		printInteractiveHelp()

	case "/exit", "/quit", "/q":
		fmt.Println("Goodbye!")
		os.Exit(0)

	case "/clear", "/cls":
		clearScreen()

	case "/search":
		if args == "" {
			fmt.Println("Usage: /search <pattern>")
			return
		}
		performInteractiveSearch(args)

	case "/add":
		if args == "" {
			fmt.Println("Usage: /add <command> | <description>")
			fmt.Println("Example: /add docker ps -a | List all containers")
			return
		}
		handleAddCommand(args)

	case "/delete":
		if args == "" {
			fmt.Println("Usage: /delete <id>")
			fmt.Println("Example: /delete 123")
			return
		}
		handleDeleteCommand(args)

	case "/list":
		handleListCommand()

	case "/count":
		handleCountCommand()

	case "/ai":
		handleAIStatus(args)

	case "/embeddings":
		handleEmbeddingsCheck()

	case "/generate":
		handleGenerateEmbeddings()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Type '/help' for available commands")
	}
}

func extractKeywords(input string) string {
	input = strings.ToLower(input)

	// Remove common question words and phrases
	removeWords := []string{
		"show me", "give me", "provide me with", "provide", "find me",
		"how to", "how do i", "what is", "what are", "can you",
		"please", "i need", "i want", "looking for", "search for",
		"example", "examples", "command", "commands", "for",
	}

	for _, word := range removeWords {
		input = strings.ReplaceAll(input, word, " ")
	}

	// Clean up extra spaces
	input = strings.Join(strings.Fields(input), " ")
	input = strings.TrimSpace(input)

	return input
}

func performInteractiveSearch(pattern string) {
	// Use smart search with Ollama if available
	results, aiResponse, err := SmartSearch(pattern, true)
	if err != nil {
		fmt.Printf("Error searching: %v\n", err)
		return
	}

	fmt.Println()

	// If we have an AI response, show it first
	if aiResponse != "" {
		fmt.Println("🤖 AI Assistant:")
		fmt.Println("══════════════════════════════════════════════════════════════")
		fmt.Println(aiResponse)
		fmt.Println("══════════════════════════════════════════════════════════════")
		fmt.Println()
	}

	if len(results) == 0 {
		if aiResponse == "" {
			fmt.Printf("No results found for: %s\n", pattern)
			fmt.Println("Try different keywords or use /search with comma-separated patterns")
			fmt.Println()
		}
		return
	}

	fmt.Printf("Found %d result(s) for: %s\n", len(results), pattern)

	// Score results to show match quality
	scored := ScoreCommands(results, pattern)

	// Filter out weak matches (less than 25%)
	var filteredResults []CommandRecord
	var filteredScored []CommandScore
	minMatchThreshold := 25 // Minimum 25% match required

	for i, s := range scored {
		if s.Score >= minMatchThreshold {
			filteredResults = append(filteredResults, results[i])
			filteredScored = append(filteredScored, s)
		}
	}

	// If no results after filtering, show message
	if len(filteredResults) == 0 {
		fmt.Println()
		fmt.Printf("No relevant results found for: %s\n", pattern)
		if len(results) > 0 {
			fmt.Printf("Found %d results but all had less than %d%% match.\n", len(results), minMatchThreshold)
		}
		if aiResponse == "" {
			fmt.Println("Try different or more specific keywords.")
		}
		fmt.Println()
		return
	}

	// Update count to show filtered results
	filteredCount := len(results) - len(filteredResults)
	if filteredCount > 0 {
		fmt.Printf("(Showing %d results with ≥%d%% match, filtered %d weak matches)\n", len(filteredResults), minMatchThreshold, filteredCount)
	}

	if len(filteredScored) > 0 && filteredScored[0].Score > 0 {
		fmt.Printf("(Best match: %d%% - %d/%d words matched)\n", filteredScored[0].Score, filteredScored[0].MatchCount, filteredScored[0].TotalWords)
	}
	fmt.Println("══════════════════════════════════════════════════════════════")

	for i, result := range filteredResults {
		fmt.Println()

		// Show match score
		if i < len(filteredScored) {
			fmt.Printf("ID: %d (Match: %d%%)\n", result.Id, filteredScored[i].Score)
		} else {
			fmt.Printf("ID: %d\n", result.Id)
		}

		fmt.Printf("Description: %s\n", result.Data)
		fmt.Println("Command:")

		// Check if it's code
		if isCode(result.Key) {
			cmd := result.Key
			if !strings.HasSuffix(cmd, "{{end}}") {
				cmd = replaceLast(cmd, "}", "\n}")
			}
			cmd = strings.ReplaceAll(cmd, "\n\t\n\t", "\n\t\t")
			fmt.Println(cmd)
		} else {
			fmt.Println(result.Key)
		}
		fmt.Println("──────────────────────────────────────────────────────────────")
	}
	fmt.Println()
}

func handleAddCommand(args string) {
	parts := strings.SplitN(args, "|", 2)
	if len(parts) != 2 {
		fmt.Println("Error: Use | to separate command and description")
		fmt.Println("Example: /add docker ps -a | List all containers")
		return
	}

	command := strings.TrimSpace(parts[0])
	description := strings.TrimSpace(parts[1])

	if command == "" || description == "" {
		fmt.Println("Error: Both command and description are required")
		return
	}

	// Check if command exists
	exists, err := CheckCommandExists(command)
	if err != nil {
		fmt.Printf("Error checking command: %v\n", err)
		return
	}

	if exists {
		fmt.Println("Error: This command already exists in the database")
		return
	}

	// Add command
	success, err := AddCommand(command, description)
	if err != nil {
		fmt.Printf("Error adding command: %v\n", err)
		return
	}

	if success {
		fmt.Println()
		fmt.Println("✓ Command added successfully!")
		fmt.Printf("  Command: %s\n", command)
		fmt.Printf("  Description: %s\n", description)
		fmt.Println()
	} else {
		fmt.Println("Failed to add command")
	}
}

func handleDeleteCommand(args string) {
	idStr := strings.TrimSpace(args)
	if idStr == "" {
		fmt.Println("Error: ID required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Error: Invalid ID. Please provide a number.")
		return
	}

	// Verify command exists and show it before deleting
	// We need a way to get a single command by ID.
	// SearchCommands doesn't support ID lookup directly but we can query DB.

	// Let's first implementation just try to delete and report success/fail
	// Better UX: Show what we are deleting.

	// We'll trust the user knows the ID from /list or /search for now to keep it simple,
	// or we can quickly implement a GetCommandById in database.go if needed.
	// But let's just do the delete for now with a confirmation.

	fmt.Printf("Are you sure you want to delete command with ID %d? (y/n): ", id)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("Cancelled.")
		return
	}

	success, err := DeleteCommand(id)
	if err != nil {
		fmt.Printf("Error deleting command: %v\n", err)
		return
	}

	if success {
		fmt.Printf("✓ Command %d deleted successfully.\n", id)
	} else {
		fmt.Printf("Command %d not found or could not be deleted.\n", id)
	}
}

func handleListCommand() {
	// Get recent commands (limit 10)
	received, err := SearchCommands("", "json")
	if err != nil {
		fmt.Printf("Error listing commands: %v\n", err)
		return
	}

	var results []CommandRecord
	if err := json.Unmarshal(received, &results); err != nil {
		fmt.Printf("Error parsing results: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("No commands found in database")
		return
	}

	// Show last 10
	start := 0
	if len(results) > 10 {
		start = len(results) - 10
	}

	fmt.Println()
	fmt.Printf("Recent Commands (showing %d of %d):\n", len(results[start:]), len(results))
	fmt.Println("══════════════════════════════════════════════════════════════")

	for _, result := range results[start:] {
		fmt.Printf("\nID: %d - %s\n", result.Id, result.Data)
		cmdPreview := result.Key
		if len(cmdPreview) > 80 {
			cmdPreview = cmdPreview[:77] + "..."
		}
		fmt.Printf("    %s\n", cmdPreview)
	}
	fmt.Println()
}

func handleCountCommand() {
	received, err := SearchCommands("", "json")
	if err != nil {
		fmt.Printf("Error counting commands: %v\n", err)
		return
	}

	var results []CommandRecord
	if err := json.Unmarshal(received, &results); err != nil {
		fmt.Printf("Error parsing results: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Printf("Total commands in database: %d\n", len(results))
	fmt.Println()
}

func clearScreen() {
	// Windows
	fmt.Print("\033[H\033[2J")
	printWelcome()
}

func handleAIStatus(args string) {
	if IsGeminiAvailable() {
		fmt.Println()
		fmt.Println("🚀 Gemini API is available and active")
		fmt.Printf("  Model: %s\n", os.Getenv("GEMINIMODEL"))
		fmt.Printf("  Embedding: %s\n", os.Getenv("GEMINI_EMBEDDING_MODEL"))
		fmt.Println()
	}

	if IsOllamaAvailable() {
		fmt.Println()
		fmt.Println("🤖 Ollama is available and active")
		fmt.Printf("  Host: %s\n", os.Getenv("OLLAMA"))
		fmt.Printf("  Model: %s\n", os.Getenv("MODEL"))
		fmt.Println()
	}

	if !IsGeminiAvailable() && !IsOllamaAvailable() {
		fmt.Println()
		fmt.Println("⚠ No AI providers available")
		fmt.Println("To enable AI features, set GEMINIAPI in .env or run Ollama locally.")
		fmt.Println()
		return
	}

	fmt.Println("AI-enhanced search is automatically used when available.")
	fmt.Println("Features:")
	fmt.Println("  - Vector similarity search for better relevance")
	fmt.Println("  - AI-generated explanations and context")
	fmt.Println("  - Automatic fallback to traditional search if needed")
	fmt.Println()
}

func handleEmbeddingsCheck() {
	tableName := os.Getenv("TB_NAME")
	if tableName == "" {
		tableName = "scmd"
	}

	// Count records with and without embeddings
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total,
			COUNT(embedding) as with_embeddings,
			COUNT(*) - COUNT(embedding) as without_embeddings
		FROM %s
	`, tableName)

	var total, withEmbeddings, withoutEmbeddings int
	err := db.QueryRow(query).Scan(&total, &withEmbeddings, &withoutEmbeddings)
	if err != nil {
		fmt.Printf("Error checking embeddings: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("Embedding Statistics:")
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Printf("Total records:              %d\n", total)
	fmt.Printf("Records with embeddings:    %d (%.1f%%)\n", withEmbeddings, float64(withEmbeddings)/float64(total)*100)
	fmt.Printf("Records without embeddings: %d (%.1f%%)\n", withoutEmbeddings, float64(withoutEmbeddings)/float64(total)*100)
	fmt.Println()

	if withoutEmbeddings > 0 {
		fmt.Println("⚠ Warning: Some records don't have embeddings!")
		fmt.Println("Vector search only works on records with embeddings.")
		fmt.Println("Traditional keyword search will be used as fallback.")
		fmt.Println()
		fmt.Println("To generate embeddings for all records, you need to:")
		fmt.Println("1. Use a script to generate embeddings for each record")
		fmt.Println("2. Update the embedding column in PostgreSQL")
		fmt.Println()
	} else {
		fmt.Println("✓ All records have embeddings!")
		fmt.Println("Vector search will work optimally.")
		fmt.Println()
	}
}

func handleGenerateEmbeddings() {
	fmt.Println()
	fmt.Println("This will generate embeddings for all commands without them.")
	fmt.Print("Continue? (y/n): ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("Cancelled.")
		return
	}

	if err := GenerateEmbeddingsForAll(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
