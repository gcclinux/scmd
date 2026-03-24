package cli

import (
	"fmt"
	"strings"

	"github.com/gcclinux/scmd/internal/ai/gemini"
	"github.com/gcclinux/scmd/internal/ai/ollama"
)

func printInteractiveHelp() {
	fmt.Println()
	fmt.Println("Available Slash Commands:")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
	fmt.Println("  /help or /?           - Show this help message                │  /ai                   - Show AI/Ollama status")
	fmt.Println("  /help next            - Show next page of help                │  /config	         - Step by step setup ai parameters")
	fmt.Println("  /search <pattern>     - Search for commands matching pattern  │  /embeddings           - Check embedding statistics")
	fmt.Println("  /add <cmd> | <desc>   - Add a new command (use | separator)   │  /import <path>        - Import a markdown document")
	fmt.Println("  /delete <id>          - Delete a command by ID                │  /generate             - Generate embeddings for all commands")
	fmt.Println("  /show <id>            - Show command and description by ID    │  /clear or /cls        - Clear the screen")
	fmt.Println("  /list                 - List recent commands                  │  /count                - Show total number of commands")
	fmt.Println("  /run <command>        - Execute a system command              │  /exit, /quit, or /q   - Exit interactive mode")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("AI Response Feedback:")
	fmt.Println("──────────────────────────────────────────────────────────────")
	fmt.Println("  After an AI response, you can rate it:")
	fmt.Println("  [1] - Good answer (saves to database for future searches)")
	fmt.Println("  [2] - Bad answer (discards without saving)")
	fmt.Println()
	fmt.Println("Quick Shortcuts (without slash):")
	fmt.Println("──────────────────────────────────────────────────────────────")
	fmt.Println("  help or ?             - Show this help message")
	fmt.Println("  clear or cls          - Clear the screen")
	fmt.Println("  exit, quit, or q      - Exit interactive mode")
	fmt.Println()
}

func printInteractiveHelpNext() {
	fmt.Println()
	if gemini.IsAvailable() || ollama.IsAvailable() {
		aiProvider := "Gemini"
		if !gemini.IsAvailable() && ollama.IsAvailable() {
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
	fmt.Println()
	fmt.Println("  Comma-separated = OR (any pattern can match):")
	fmt.Println("    docker,kubernetes,postgresql")
	fmt.Println()
}

// detectCommandLanguage detects the appropriate language tag for code blocks.
func detectCommandLanguage(command, description string) string {
	combined := strings.ToLower(command + " " + description)

	if strings.Contains(combined, "docker") {
		return "docker"
	}
	if strings.Contains(combined, "kubectl") || strings.Contains(combined, "kubernetes") {
		return "bash"
	}
	if strings.Contains(combined, "psql") || strings.Contains(combined, "postgresql") ||
		strings.Contains(command, "SELECT") || strings.Contains(command, "INSERT") ||
		strings.Contains(command, "UPDATE") || strings.Contains(command, "DELETE") ||
		strings.Contains(command, "CREATE TABLE") {
		return "postgresql"
	}
	if strings.Contains(combined, "mysql") {
		return "sql"
	}
	if strings.Contains(combined, "mongo") {
		return "javascript"
	}
	if strings.Contains(command, "import ") || strings.Contains(command, "def ") ||
		strings.Contains(command, "print(") {
		return "python"
	}
	if strings.Contains(command, "const ") || strings.Contains(command, "let ") ||
		strings.Contains(command, "function ") || strings.Contains(command, "=>") {
		return "javascript"
	}
	if strings.Contains(combined, "powershell") || strings.Contains(command, "Get-") ||
		strings.Contains(command, "Set-") || strings.Contains(command, "$_") {
		return "powershell"
	}
	if strings.Contains(command, "git ") {
		return "bash"
	}

	return "bash"
}
