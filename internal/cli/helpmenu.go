package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gcclinux/scmd/internal/ai"
	"github.com/gcclinux/scmd/internal/database"
	"github.com/gcclinux/scmd/internal/updater"
)

const (
	InfoColor    = "\033[1;34m%s\033[0m"
	NoticeColor  = "\033[1;36m%s\033[0m"
	WarningColor = "\033[1;33m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"
	DebugColor   = "\033[0;36m%s\033[0m"
)

// PrintHelp displays the comprehensive CLI help menu.
func PrintHelp(name string) {
	green := "\033[32m"
	cyan := "\033[36m"
	reset := "\033[0m"

	fmt.Println()
	fmt.Println("  ╔════════════════════════════════════════════════════════════╗")
	fmt.Println("  ║                                                            ║")
	fmt.Printf("  ║            %s███████╗ ██████╗███╗   ███╗██████╗%s              ║\n", green, reset)
	fmt.Printf("  ║            %s██╔════╝██╔════╝████╗ ████║██╔══██╗%s             ║\n", green, reset)
	fmt.Printf("  ║            %s███████╗██║     ██╔████╔██║██║  ██║%s             ║\n", green, reset)
	fmt.Printf("  ║            %s╚════██║██║     ██║╚██╔╝██║██║  ██║%s             ║\n", green, reset)
	fmt.Printf("  ║            %s███████║╚██████╗██║ ╚═╝ ██║██████╔╝%s             ║\n", green, reset)
	fmt.Printf("  ║             %s╚══════╝ ╚═════╝╚═╝     ╚═╝╚═════╝%s             ║\n", green, reset)
	fmt.Println("  ║                                                            ║")
	fmt.Printf("  ║          %s⚡ AI-Powered Command Search · v%-15s%s  ║\n", cyan, updater.Release+" ⚡", reset)
	fmt.Println("  ╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Display this help menu\n\r")
	fmt.Println("Usage: \t", name, "--help")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Start Interactive CLI Mode (NEW!)\n\r")
	fmt.Println("Usage: \t", name, "--cli")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Start MCP Server (Model Context Protocol)\n\r")
	fmt.Println("Usage: \t", name, "--mcp")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Opens the Web UI with default Port: \"3333\" \n\r")
	fmt.Println("Usage: \t", name, "--web")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Opens the Web UI with default Port: \"3333\" & \"DISABLE\" add commands\n\r")
	fmt.Println("Usage: \t", name, "--web", "-block")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Opens the Web UI with alternative Port:\n\r")
	fmt.Println("Usage: \t", name, "--web", "-port", "[port]")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Opens the Web UI with alternative Port: & \"DISABLE\" add commands\n\r")
	fmt.Println("Usage: \t", name, "--web", "-port", "[port]", "-block")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Starts SCMD without launching Web UI\n\r")
	fmt.Println("Usage: \t", name, "--web", "-port", "[port]", "-service")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Starts SCMD without launching Web UI & \"DISABLE\" add commands\n\r")
	fmt.Println("Usage: \t", name, "--web", "-port", "[port]", "-service", "-block")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Opens SSL Web UI with default Port: \"3333\" \n\r")
	fmt.Println("Usage: \t", name, "--ssl", "[certificate.pem]", "[privkey.pem]")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Starts SCMD without launching Web UI default Port: \"3333\" \n\r")
	fmt.Println("Usage: \t", name, "--ssl", "-service", "[certificate.pem]", "[privkey.pem]")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Opens SSL Web UI with default Port: \"3333\" & \"DISABLE\" add commands\n\r")
	fmt.Println("Usage: \t", name, "--ssl", "[certificate.pem]", "[privkey.pem]", "-block")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Opens SSL web UI with alternative Port:\n\r")
	fmt.Println("Usage: \t", name, "--ssl", "-port", "[port]", "[certificate.pem]", "[privkey.pem]")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Opens SSL web UI with alternative Port: & \"DISABLE\" add commands\n\r")
	fmt.Println("Usage: \t", name, "--ssl", "-port", "[port]", "[certificate.pem]", "[privkey.pem]", "-block")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Starts SCMD SSL without launching Web UI\n\r")
	fmt.Println("Usage: \t", name, "--ssl", "-port", "[port]", "-service", "[certificate.pem]", "[privkey.pem]")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Starts SCMD SSL without launching Web UI & \"DISABLE\" add commands\n\r")
	fmt.Println("Usage: \t", name, "--ssl", "-port", "[port]", "-service", "[certificate.pem]", "[privkey.pem]", "-block")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Show local and available scmd version\n\r")
	fmt.Println("Usage: \t", name, "--version")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Download and upgrade the latest version of the scmd application binary\n\r")
	fmt.Println("Usage: \t", name, "--upgrade")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Search command based on comma separated pattern(s)\n\r")
	fmt.Println("Usage: \t", name, "--search [pattern(s)]")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Save new command with description in the local database\n\r")
	fmt.Println("Usage: \t", name, "--save [command] [description]")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Generate embeddings for all commands (enables vector search)\n\r")
	fmt.Println("Usage: \t", name, "--generate-embeddings")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Show embedding statistics for the database\n\r")
	fmt.Println("Usage: \t", name, "--embedding-stats")
	fmt.Println()

	fmt.Printf(NoticeColor, "*** Interactive Ollama AI server setup (prompts for host, model, embedding config)\n\r")
	fmt.Println("Usage: \t", name, "--server-ollama")
	fmt.Println()
	fmt.Printf(NoticeColor, "*** Interactive Gemini AI server setup (prompts for API key, model, embedding config)\n\r")
	fmt.Println("Usage: \t", name, "--server-gemini")
	fmt.Println()
}

// PrintWrongSyntax shows usage error.
func PrintWrongSyntax() {
	fmt.Println()
	fmt.Println("Usage: \t", GetName(), "--help")
	fmt.Println()
}

// GetName returns the executable base name.
func GetName() string {
	return filepath.Base(os.Args[0])
}

// SaveCmd saves a command with description to the database.
func SaveCmd(cmd, details string) {
	details = strings.TrimSpace(details)
	if strings.HasPrefix(details, "```") {
		if idx := strings.Index(details, "\n"); idx != -1 {
			details = details[idx+1:]
		}
	}
	details = strings.TrimSpace(details)
	details = strings.TrimSuffix(details, "```")
	details = strings.TrimSpace(details)

	// Import needed packages
	ai.InitProviders()

	if err := database.InitDB(); err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer database.CloseDB()

	status, err := database.AddCommand(cmd, details, ai.GetBestEmbedding)
	if err != nil {
		fmt.Println("Error saving command:", err)
		fmt.Println("returned: ( false )")
	} else {
		fmt.Println("returned: (", status, ")")
	}
	fmt.Println()
}
