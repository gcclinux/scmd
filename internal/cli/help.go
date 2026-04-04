package cli

import (
	"fmt"
	"strings"
)

func printInteractiveHelp() {
	fmt.Println()
	fmt.Println("Available Slash Commands:")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
	fmt.Println("  /help or /?           - Show this help message                │  /ai                   - Show AI/Ollama status")
	fmt.Println("  /search <pattern>     - Search for commands matching pattern  │  /config               - Show current config.json settings")
	fmt.Println("  /add <cmd> | <desc>   - Add a new command (use | separator)   │  /embeddings           - Check embedding statistics")
	fmt.Println("  /delete <id>          - Delete a command by ID                │  /import <path>        - Import a markdown document")
	fmt.Println("  /show <id>            - Show command and description by ID    │  /generate             - Generate embeddings for all commands")
	fmt.Println("  /list                 - List recent commands                  │  /clear or /cls        - Clear the screen")
	fmt.Println("  /run <command>        - Execute a system command              │  /exit, /quit, or /q   - Exit interactive mode")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
	fmt.Println("AI Personas (Focused Context):")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
	fmt.Println("  /ubuntu <query>       - Ubuntu expert persona                 │  /powershell <query>   - PowerShell guru persona")
	fmt.Println("  /debian <query>       - Debian expert persona                 │  /archlinux <query>    - Arch Linux master persona")
	fmt.Println("  /fedora <query>       - Fedora expert persona                 │  /windows <query>      - Windows admin persona")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("AI Response Feedback:")
	fmt.Println("──────────────────────────────────────────────────────────────")
	fmt.Println("  After an AI response, you can rate it:")
	fmt.Println("  [1] - Good answer (saves to database for future searches)")
	fmt.Println("  [2] - Bad answer (discards without saving)")
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
