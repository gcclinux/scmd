package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// StartInteractiveMode starts the interactive CLI prompt
func StartInteractiveMode() {
	// Initialize database connection
	if err := InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer CloseDB()

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
	fmt.Println("║          SCMD Interactive CLI - PostgreSQL Edition            ║")
	fmt.Println("║                    Version", Release, "                           ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
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
	fmt.Println("  /list                 - List recent commands")
	fmt.Println("  /count                - Show total number of commands")
	fmt.Println("  /clear or /cls        - Clear the screen")
	fmt.Println("  /exit, /quit, or /q   - Exit interactive mode")
	fmt.Println()
	fmt.Println("Quick Shortcuts (without slash):")
	fmt.Println("──────────────────────────────────────────────────────────────")
	fmt.Println("  help or ?             - Show this help message")
	fmt.Println("  clear or cls          - Clear the screen")
	fmt.Println("  exit, quit, or q      - Exit interactive mode")
	fmt.Println()
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

	case "/list":
		handleListCommand()

	case "/count":
		handleCountCommand()

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
	received, err := SearchCommands(pattern, "json")
	if err != nil {
		fmt.Printf("Error searching: %v\n", err)
		return
	}

	var results []CommandRecord
	if err := json.Unmarshal(received, &results); err != nil {
		fmt.Printf("Error parsing results: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println()
		fmt.Printf("No results found for: %s\n", pattern)
		fmt.Println("Try different keywords or use /search with comma-separated patterns")
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Printf("Found %d result(s) for: %s\n", len(results), pattern)
	fmt.Println("══════════════════════════════════════════════════════════════")

	for _, result := range results {
		fmt.Println()
		fmt.Printf("ID: %d\n", result.Id)
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
