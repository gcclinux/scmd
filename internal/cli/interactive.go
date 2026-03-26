package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gcclinux/scmd/internal/ai"
	"github.com/gcclinux/scmd/internal/ai/gemini"
	"github.com/gcclinux/scmd/internal/ai/ollama"
	"github.com/gcclinux/scmd/internal/database"
	"github.com/gcclinux/scmd/internal/markdown"
	"github.com/gcclinux/scmd/internal/search"
	"github.com/gcclinux/scmd/internal/updater"
)

// StartInteractiveMode starts the interactive CLI prompt.
func StartInteractiveMode() {
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDB()

	ai.InitProviders()

	reader := bufio.NewReader(os.Stdin)
	printWelcome()

	var lastAIResponse string
	var lastQuery string

	for {
		fmt.Print("scmd> ")

		// Multi-line input: accumulate lines until the user submits
		// by pressing Enter on an empty line (blank line = submit).
		// This lets users paste or type multi-line queries freely.
		var lines []string
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				break
			}
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				// Empty line: submit whatever we have accumulated
				break
			}
			lines = append(lines, trimmed)
			fmt.Print("  ... ")
		}

		input := strings.Join(lines, " ")
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Feedback on last AI response
		if lastAIResponse != "" && (input == "1" || input == "2") {
			if input == "1" {
				if _, err := database.AddCommand(lastAIResponse,
					fmt.Sprintf("AI-generated response for: %s", lastQuery),
					ai.GetBestEmbedding); err != nil {
					fmt.Printf("Error saving response: %v\n", err)
				} else {
					fmt.Println("✓ Response saved to database!")
					fmt.Println()
				}
				lastAIResponse = ""
				lastQuery = ""
			} else if input == "2" {
				fmt.Println("Regenerating response...")
				fmt.Println()
				aiResp := regenerateAIResponse(lastQuery)
				if aiResp != "" {
					lastAIResponse = aiResp
					fmt.Println("Good Answer [1], Bad Answer [2]")
				} else {
					fmt.Println("Failed to regenerate response.")
					fmt.Println()
					lastAIResponse = ""
					lastQuery = ""
				}
			}
			continue
		}

		if input == "exit" || input == "quit" || input == "q" {
			fmt.Println("Goodbye!")
			break
		}

		if input == "help" || input == "?" {
			printInteractiveHelp()
			continue
		}

		if input == "clear" || input == "cls" {
			clearScreen()
			continue
		}

		aiResp := processInteractiveCommand(input)
		if aiResp != "" {
			lastAIResponse = aiResp
			lastQuery = input
			fmt.Println("Good Answer [1], Bad Answer [2]")
		} else {
			lastAIResponse = ""
			lastQuery = ""
		}
	}
}

func printWelcome() {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║           SCMD Interactive CLI - PostgreSQL Edition            ║")
	fmt.Println("║                      Version", updater.Release, "                            ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	preferred := strings.ToLower(os.Getenv("AGENT"))
	switch preferred {
	case "ollama":
		if ollama.IsAvailable() {
			fmt.Println("🤖 Ollama: Active (vector search enabled)")
		} else {
			fmt.Println("⚠️  Preferred agent 'ollama' is not available")
		}
	case "gemini":
		if gemini.IsAvailable() {
			fmt.Println("🚀 Gemini API: Active (vector search enabled)")
		} else {
			fmt.Println("⚠️  Preferred agent 'gemini' is not available")
		}
	default:
		if ollama.IsAvailable() {
			fmt.Println("🤖 Ollama: Active (vector search enabled)")
		} else if gemini.IsAvailable() {
			fmt.Println("🚀 Gemini API: Active (vector search enabled)")
		} else {
			fmt.Println("⚠️  No embedding provider (traditional search only)")
		}
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
	fmt.Println("Input: Type or paste your query, then press Enter on an empty line to submit.")
	fmt.Println()
}

func processInteractiveCommand(input string) string {
	if strings.HasPrefix(input, "/") {
		handleSlashCommand(input)
		return ""
	}

	keywords := extractKeywords(input)
	if keywords == "" {
		fmt.Println("Could not extract search terms. Try using /search <pattern>")
		return ""
	}

	return performInteractiveSearch(keywords)
}

func extractKeywords(input string) string {
	input = strings.ToLower(input)

	removeWords := []string{
		"show me", "give me", "provide me with", "provide", "find me",
		"how to", "how do i", "what is", "what are", "can you",
		"please", "i need", "i want", "looking for", "search for",
		"example", "examples", "command", "commands", "for",
	}

	for _, word := range removeWords {
		input = strings.ReplaceAll(input, word, " ")
	}

	input = strings.Join(strings.Fields(input), " ")
	return strings.TrimSpace(input)
}

func performInteractiveSearch(pattern string) string {
	results, aiResponse, _, err := ai.SmartSearch(pattern, true)
	if err != nil {
		fmt.Printf("Error searching: %v\n", err)
		return ""
	}

	fmt.Println()

	if aiResponse != "" {
		fmt.Println("🤖 AI Assistant:")
		fmt.Println("══════════════════════════════════════════════════════════════")
		fmt.Println(aiResponse)
		fmt.Println("══════════════════════════════════════════════════════════════")
		fmt.Println()
		return aiResponse
	}

	if len(results) == 0 {
		return ""
	}

	fmt.Printf("Found %d result(s) for: %s\n", len(results), pattern)

	scored := search.ScoreCommands(results, pattern)

	var filteredResults []database.CommandRecord
	var filteredScored []search.CommandScore
	minMatchThreshold := 60

	for i, s := range scored {
		if s.Score >= minMatchThreshold {
			filteredResults = append(filteredResults, results[i])
			filteredScored = append(filteredScored, s)
		}
	}

	if len(filteredResults) == 0 {
		return ""
	}

	filteredCount := len(results) - len(filteredResults)
	if filteredCount > 0 {
		fmt.Printf("(Showing %d results with ≥%d%% match, filtered %d weak matches)\n", len(filteredResults), minMatchThreshold, filteredCount)
	}

	if len(filteredScored) > 0 && filteredScored[0].Score > 0 {
		fmt.Printf("(Best match: %d%% - %d/%d words matched)\n", filteredScored[0].Score, filteredScored[0].MatchCount, filteredScored[0].TotalWords)
	}
	fmt.Println("══════════════════════════════════════════════════════════════")

	for _, result := range filteredResults {
		fmt.Println()

		if markdown.IsMarkdownContent(result.Data) {
			fmt.Print(markdown.Render(result.Data))
		} else {
			fmt.Println(result.Data)
		}

		fmt.Println()

		if markdown.IsMarkdownContent(result.Key) {
			fmt.Print(markdown.Render(result.Key))
		} else {
			lang := detectCommandLanguage(result.Key, result.Data)
			fmt.Printf("```%s\n%s\n```\n", lang, strings.TrimSpace(result.Key))
		}

		fmt.Println("──────────────────────────────────────────────────────────────")
	}
	fmt.Println()

	return ""
}

func regenerateAIResponse(query string) string {
	cleanedQuery := extractKeywords(query)
	if cleanedQuery == "" {
		cleanedQuery = query
	}

	jsonData, err := database.SearchCommands(cleanedQuery, "json")
	if err != nil {
		fmt.Printf("Error searching: %v\n", err)
		return ""
	}

	var results []database.CommandRecord
	json.Unmarshal(jsonData, &results)

	scored := search.ScoreCommands(results, cleanedQuery)
	var contextResults []database.CommandRecord
	for _, s := range search.GetBestMatches(scored, 5) {
		if s.Score > 0 {
			contextResults = append(contextResults, s.Record)
		}
	}

	var aiResponse string
	preferredAgent := strings.ToLower(os.Getenv("AGENT"))

	tryOllama := func() bool {
		if !ollama.IsAvailable() {
			return false
		}
		fmt.Println("⚠ Regenerating with Ollama...")
		resp, _, err := ollama.Ask(query, contextResults)
		if err == nil && resp != "" {
			aiResponse = resp
			fmt.Println("🤖 AI Assistant:")
			fmt.Println("══════════════════════════════════════════════════════════════")
			fmt.Println(aiResponse)
			fmt.Println("══════════════════════════════════════════════════════════════")
			fmt.Println()
			return true
		}
		return false
	}

	tryGemini := func() bool {
		if !gemini.IsAvailable() {
			return false
		}
		fmt.Println("⚠ Regenerating with Gemini...")
		resp, _, err := gemini.Ask(query, contextResults)
		if err == nil && resp != "" {
			aiResponse = resp
			fmt.Println("🤖 AI Assistant:")
			fmt.Println("══════════════════════════════════════════════════════════════")
			fmt.Println(aiResponse)
			fmt.Println("══════════════════════════════════════════════════════════════")
			fmt.Println()
			return true
		}
		return false
	}

	if preferredAgent == "ollama" {
		if tryOllama() {
			return aiResponse
		}
	} else if preferredAgent == "gemini" {
		if tryGemini() {
			return aiResponse
		}
	}

	if preferredAgent == "" {
		if tryOllama() {
			return aiResponse
		}
		if tryGemini() {
			return aiResponse
		}
	}

	return ""
}
