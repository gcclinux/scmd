package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/gcclinux/scmd/internal/ai"
	"github.com/gcclinux/scmd/internal/ai/gemini"
	"github.com/gcclinux/scmd/internal/ai/ollama"
	"github.com/gcclinux/scmd/internal/config"
	"github.com/gcclinux/scmd/internal/database"
	"github.com/gcclinux/scmd/internal/markdown"
	"github.com/gcclinux/scmd/internal/util"
)

// showOriginalQuery holds the original query extracted from a /show record's
// description, so the feedback loop can use it for regeneration.
var showOriginalQuery string

func handleSlashCommand(input string) string {
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
			return ""
		}
		return performInteractiveSearch(args)
	case "/add":
		if args == "" {
			fmt.Println("Usage: /add <command> | <description>")
			fmt.Println("Example: /add docker ps -a | List all containers")
			return ""
		}
		handleAddCommand(args)
	case "/delete":
		if args == "" {
			fmt.Println("Usage: /delete <id>")
			return ""
		}
		handleDeleteCommand(args)
	case "/list":
		handleListCommand()
	case "/ai":
		handleAIStatus()
	case "/config":
		handleConfigShow()
	case "/embeddings":
		handleEmbeddingsCheck()
	case "/generate":
		handleGenerateEmbeddings()
	case "/show":
		if args == "" {
			fmt.Println("Usage: /show <id>")
			return ""
		}
		content, origQuery := handleShowCommand(args)
		showOriginalQuery = origQuery
		return content

	case "/run":
		if args == "" {
			fmt.Println("Usage: /run <command>")
			return ""
		}
		handleRunCommand(args)
	case "/ubuntu", "/debian", "/fedora", "/windows", "/powershell", "/archlinux":
		if args == "" {
			fmt.Printf("Usage: %s <question>\n", command)
			return ""
		}
		return handlePersonaCommand(command[1:], args)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Type '/help' for available commands")
	}
	return ""
}

func handlePersonaCommand(persona, query string) string {
	fmt.Printf("🤖 Processing with %s persona...\n", persona)

	// We'll perform a search to get context for the persona
	results, _, _, err := ai.SmartSearch(query, true)
	if err != nil {
		fmt.Printf("Error searching: %v\n", err)
	}

	aiResp, _, err := ai.AskAIPersona(persona, query, results)
	if err != nil {
		fmt.Printf("Error from AI: %v\n", err)
		return ""
	}

	fmt.Println()
	fmt.Printf("🤖 AI %s Persona:\n", strings.Title(persona))
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Print(markdown.Render(aiResp))
	if !strings.HasSuffix(aiResp, "\n") {
		fmt.Println()
	}
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println()

	return aiResp
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
		fmt.Println("To enable AI features, set gemini_api in ~/.scmd/config.json or run Ollama locally.")
		fmt.Println()
		return
	}

	fmt.Println("AI-enhanced search is automatically used when available.")
	fmt.Println()
}
func handleConfigShow() {
	cfgPath := config.ConfigFilePath()
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		fmt.Println()
		fmt.Printf("⚠ Could not read config file: %s\n", cfgPath)
		fmt.Printf("  Error: %v\n", err)
		fmt.Println()
		fmt.Println("To create one, copy the example:")
		fmt.Println("  mkdir -p ~/.scmd")
		fmt.Println("  cp config.json.example ~/.scmd/config.json")
		fmt.Println()
		return
	}

	var cfg config.ConfigData
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Println()
		fmt.Printf("⚠ Could not parse config file: %v\n", err)
		fmt.Println()
		return
	}

	mask := func(s string) string {
		if s == "" {
			return "(not set)"
		}
		if len(s) <= 6 {
			return "****"
		}
		return s[:3] + "****" + s[len(s)-3:]
	}

	fmt.Println()
	fmt.Printf("📄 Config: %s\n", cfgPath)
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("  AI Settings:")
	fmt.Printf("    agent:                  %s\n", cfg.Agent)
	fmt.Println()
	fmt.Println("  Gemini:")
	fmt.Printf("    gemini_api:             %s\n", mask(cfg.GeminiAPI))
	fmt.Printf("    gemini_model:           %s\n", cfg.GeminiModel)
	fmt.Printf("    gemini_embedding_model: %s\n", cfg.GeminiEmbeddingModel)
	fmt.Println()
	fmt.Println("  Ollama:")
	fmt.Printf("    ollama:                 %s\n", cfg.Ollama)
	fmt.Printf("    model:                  %s\n", cfg.Model)
	fmt.Printf("    embedding_model:        %s\n", cfg.EmbeddingModel)
	fmt.Printf("    embedding_dim:          %s\n", cfg.EmbeddingDim)
	fmt.Println()
	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════════════════")
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



func handleShowCommand(args string) (string, string) {
	idStr := strings.TrimSpace(args)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Error: Invalid ID. Please provide a number.")
		return "", ""
	}

	record, err := database.GetCommandByID(id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return "", ""
	}

	// Extract the original query from the description if it was AI-generated.
	originalQuery := ""
	const prefix = "AI-generated response for: "
	if strings.HasPrefix(record.Data, prefix) {
		originalQuery = strings.TrimPrefix(record.Data, prefix)
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

	var content string
	if markdown.IsMarkdownContent(record.Key) {
		fmt.Println("Content:")
		fmt.Print(markdown.Render(record.Key))
		content = record.Key
	} else {
		fmt.Println("Command:")
		if util.IsCode(record.Key) {
			cmd := record.Key
			if !strings.HasSuffix(cmd, "{{end}}") {
				cmd = util.ReplaceLast(cmd, "}", "\n}")
			}
			cmd = strings.ReplaceAll(cmd, "\n\t\n\t", "\n\t\t")
			fmt.Println(cmd)
			content = cmd
		} else {
			fmt.Println(record.Key)
			content = record.Key
		}
	}

	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Return the content and the original query for regeneration.
	return content, originalQuery
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

	// Run through the user's login shell so that shell built-ins,
	// aliases, functions (e.g. nvm), and the full PATH/environment
	// from ~/.bashrc / ~/.profile are available.
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", args)
	} else {
		shell := findBash()
		if shell != "" {
			// Login + interactive flags cause bash to source
			// ~/.bash_profile / ~/.bashrc, loading functions like nvm.
			cmd = exec.Command(shell, "--login", "-c", args)
		} else {
			cmd = exec.Command("sh", "-c", args)
		}
	}

	// When running inside a snap, the confined environment may not
	// include standard system directories in PATH. Merge common paths
	// so that tools like curl, wget, git, etc. are reachable.
	cmd.Env = enrichPathEnv(os.Environ())

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

// enrichPathEnv returns a copy of env with common system directories merged
// into the PATH variable. This is a no-op when those directories are already
// present (i.e. outside snap confinement). Inside a snap the default PATH is
// very restricted, so we append the standard locations that contain tools
// like curl, wget, git, etc.
func enrichPathEnv(env []string) []string {
	// Directories that should always be reachable.
	extras := []string{
		"/usr/local/sbin",
		"/usr/local/bin",
		"/usr/sbin",
		"/usr/bin",
		"/sbin",
		"/bin",
	}

	// When running inside a snap, $SNAP/usr/bin etc. hold staged packages.
	snap := ""
	home := ""
	for _, e := range env {
		if strings.HasPrefix(e, "SNAP=") {
			snap = strings.TrimPrefix(e, "SNAP=")
		} else if strings.HasPrefix(e, "HOME=") {
			home = strings.TrimPrefix(e, "HOME=")
		}
	}
	if snap != "" {
		extras = append([]string{
			snap + "/usr/bin",
			snap + "/usr/sbin",
			snap + "/bin",
			snap + "/sbin",
		}, extras...)
	}

	// Also include the user's own ~/bin and ~/.local/bin.
	if home != "" {
		extras = append(extras, home+"/bin", home+"/.local/bin")
	}

	out := make([]string, 0, len(env))
	found := false
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			found = true
			current := strings.TrimPrefix(e, "PATH=")
			dirs := strings.Split(current, ":")
			existing := make(map[string]bool, len(dirs))
			for _, d := range dirs {
				existing[d] = true
			}
			for _, d := range extras {
				if !existing[d] {
					dirs = append(dirs, d)
				}
			}
			out = append(out, "PATH="+strings.Join(dirs, ":"))
		} else {
			out = append(out, e)
		}
	}
	if !found {
		out = append(out, "PATH="+strings.Join(extras, ":"))
	}
	return out
}

// findBash locates a usable bash binary. Inside a snap, bash may live
// under $SNAP/usr/bin/bash rather than the usual /bin/bash.
func findBash() string {
	// Check the SNAP location first (staged cli-tools part).
	if snap := os.Getenv("SNAP"); snap != "" {
		candidate := snap + "/usr/bin/bash"
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	// Standard host locations.
	for _, p := range []string{"/bin/bash", "/usr/bin/bash"} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// Last resort: let the OS find it.
	if p, err := exec.LookPath("bash"); err == nil {
		return p
	}
	return ""
}
