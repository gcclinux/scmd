package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gcclinux/scmd/internal/ai"
	"github.com/gcclinux/scmd/internal/ai/gemini"
	"github.com/gcclinux/scmd/internal/ai/ollama"
	"github.com/gcclinux/scmd/internal/database"
	"github.com/gcclinux/scmd/internal/markdown"
	"github.com/gcclinux/scmd/internal/util"
)

func handleSlashCommand(input string) {
	parts := strings.SplitN(input, " ", 2)
	command := parts[0]
	args := ""
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}

	switch command {
	case "/help", "/?":
		if args == "next" {
			printInteractiveHelpNext()
		} else {
			printInteractiveHelp()
		}
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
			return
		}
		handleDeleteCommand(args)
	case "/list":
		handleListCommand()
	case "/count":
		handleCountCommand()
	case "/ai":
		handleAIStatus()
	case "/embeddings":
		handleEmbeddingsCheck()
	case "/generate":
		handleGenerateEmbeddings()
	case "/show":
		if args == "" {
			fmt.Println("Usage: /show <id>")
			return
		}
		handleShowCommand(args)
	case "/import":
		if args == "" {
			fmt.Println("Usage: /import <path>")
			return
		}
		handleImportCommand(args)
	case "/run":
		if args == "" {
			fmt.Println("Usage: /run <command>")
			return
		}
		handleRunCommand(args)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Type '/help' for available commands")
	}
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

	exists, err := database.CheckCommandExists(command)
	if err != nil {
		fmt.Printf("Error checking command: %v\n", err)
		return
	}
	if exists {
		fmt.Println("Error: This command already exists in the database")
		return
	}

	success, err := database.AddCommand(command, description, ai.GetBestEmbedding)
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
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Error: Invalid ID. Please provide a number.")
		return
	}

	fmt.Printf("Are you sure you want to delete command with ID %d? (y/n): ", id)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("Cancelled.")
		return
	}

	success, err := database.DeleteCommand(id)
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
	received, err := database.SearchCommands("", "json")
	if err != nil {
		fmt.Printf("Error listing commands: %v\n", err)
		return
	}

	var results []database.CommandRecord
	if err := json.Unmarshal(received, &results); err != nil {
		fmt.Printf("Error parsing results: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("No commands found in database")
		return
	}

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
	received, err := database.SearchCommands("", "json")
	if err != nil {
		fmt.Printf("Error counting commands: %v\n", err)
		return
	}

	var results []database.CommandRecord
	if err := json.Unmarshal(received, &results); err != nil {
		fmt.Printf("Error parsing results: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Printf("Total commands in database: %d\n", len(results))
	fmt.Println()
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
	printWelcome()
}

func handleAIStatus() {
	if gemini.IsAvailable() {
		fmt.Println()
		fmt.Println("🚀 Gemini API is available and active")
		fmt.Printf("  Model: %s\n", os.Getenv("GEMINIMODEL"))
		fmt.Printf("  Embedding: %s\n", os.Getenv("GEMINI_EMBEDDING_MODEL"))
		fmt.Println()
	}

	if ollama.IsAvailable() {
		fmt.Println()
		fmt.Println("🤖 Ollama is available and active")
		fmt.Printf("  Host: %s\n", os.Getenv("OLLAMA"))
		fmt.Printf("  Model: %s\n", os.Getenv("MODEL"))
		fmt.Printf("  Embedding Model: %s\n", os.Getenv("EMBEDDING_MODEL"))
		fmt.Println()
	}

	if !gemini.IsAvailable() && !ollama.IsAvailable() {
		fmt.Println()
		fmt.Println("⚠ No AI providers available")
		fmt.Println("To enable AI features, set GEMINIAPI in .env or run Ollama locally.")
		fmt.Println()
		return
	}

	fmt.Println("AI-enhanced search is automatically used when available.")
	fmt.Println()
}

func handleEmbeddingsCheck() {
	if err := ai.CheckEmbeddingStats(); err != nil {
		fmt.Printf("Error checking embeddings: %v\n", err)
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

	if err := ai.GenerateEmbeddingsForAll(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func handleImportCommand(args string) {
	title, err := markdown.ImportMarkdown(args, ai.GetBestEmbedding)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("✓ Document imported successfully!")
	fmt.Printf("  Title: %s\n", title)
	fmt.Println()
}

func handleShowCommand(args string) {
	idStr := strings.TrimSpace(args)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Error: Invalid ID. Please provide a number.")
		return
	}

	record, err := database.GetCommandByID(id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Printf("ID: %d\n", record.Id)
	fmt.Println("══════════════════════════════════════════════════════════════")

	fmt.Println("Description:")
	if markdown.IsMarkdownContent(record.Data) {
		fmt.Print(markdown.Render(record.Data))
	} else {
		fmt.Printf("  %s\n", record.Data)
	}

	fmt.Println("──────────────────────────────────────────────────────────────")

	if markdown.IsMarkdownContent(record.Key) {
		fmt.Println("Content:")
		fmt.Print(markdown.Render(record.Key))
	} else {
		fmt.Println("Command:")
		if util.IsCode(record.Key) {
			cmd := record.Key
			if !strings.HasSuffix(cmd, "{{end}}") {
				cmd = util.ReplaceLast(cmd, "}", "\n}")
			}
			cmd = strings.ReplaceAll(cmd, "\n\t\n\t", "\n\t\t")
			fmt.Println(cmd)
		} else {
			fmt.Println(record.Key)
		}
	}

	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println()
}

func handleRunCommand(args string) {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  SYSTEM COMMAND EXECUTION")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("  Command: %s\n", args)
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println()

	cmdParts := strings.Fields(args)
	if len(cmdParts) == 0 {
		fmt.Println("Error: No command provided")
		return
	}

	blockedCommands := []string{
		"vi", "vim", "nvim", "nano", "emacs", "pico",
		"less", "more", "top", "htop", "man",
		"ssh", "telnet", "ftp", "sftp",
		"python", "python3", "node", "irb", "ruby",
		"mysql", "psql", "mongo", "redis-cli",
	}

	baseCommand := strings.ToLower(cmdParts[0])
	if strings.Contains(baseCommand, "/") {
		parts := strings.Split(baseCommand, "/")
		baseCommand = parts[len(parts)-1]
	}

	for _, blocked := range blockedCommands {
		if baseCommand == blocked {
			fmt.Printf("Error: Interactive command '%s' is not allowed\n", cmdParts[0])
			fmt.Println("The /run command only supports non-interactive commands.")
			fmt.Println()
			fmt.Println("═══════════════════════════════════════════════════════════════")
			fmt.Println()
			return
		}
	}

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if stdout.Len() > 0 {
		fmt.Print(stdout.String())
	}
	if stderr.Len() > 0 {
		fmt.Fprintf(os.Stderr, "%s", stderr.String())
	}
	if err != nil {
		fmt.Println()
		fmt.Printf("Command exited with error: %v\n", err)
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
}
