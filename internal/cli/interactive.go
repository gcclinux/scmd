package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gcclinux/scmd/internal/ai"
	"github.com/gcclinux/scmd/internal/ai/gemini"
	"github.com/gcclinux/scmd/internal/ai/ollama"
	"github.com/gcclinux/scmd/internal/database"
	"github.com/gcclinux/scmd/internal/markdown"
	"github.com/gcclinux/scmd/internal/search"
	"github.com/gcclinux/scmd/internal/updater"
	"github.com/gcclinux/scmd/internal/util"
)

// StartInteractiveMode starts the interactive CLI prompt.
func StartInteractiveMode() {
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDB()

	ai.InitProviders()

	// Check for unfinished workbooks at startup
	if wb, err := checkStartupWorkbook(); err == nil && wb != nil {
		fmt.Printf("\n📋 Unfinished workbook available: %s\n", filepath.Base(wb.FilePath))
		fmt.Printf("   Status: %s | Steps: %d total\n", wb.Status, len(wb.Steps))
		fmt.Print("   Would you like to load and continue? (y/n/finish): ")

		startupReader := bufio.NewReader(os.Stdin)
		response, _ := startupReader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		switch response {
		case "y", "yes":
			fmt.Printf("Loading workbook: %s\n", wb.Title)
			for _, step := range wb.Steps {
				fmt.Printf("   [%s] %d. %s\n", step.Status, step.Number, step.Description)
			}
			fmt.Println()
			if idx, err := findFirstTodoStep(wb); err == nil {
				fmt.Printf("Continuing from step %d: %s\n\n", wb.Steps[idx].Number, wb.Steps[idx].Description)
			}
		case "finish", "f":
			wb.Status = WorkbookFinished
			content := serializeWorkbook(wb)
			if err := os.WriteFile(wb.FilePath, []byte(content), 0644); err != nil {
				fmt.Printf("⚠ Could not update workbook: %v\n", err)
			} else {
				fmt.Println("   Workbook marked as FINISHED.")
			}
			fmt.Println()
		default:
			fmt.Println()
		}
	}

	reader := bufio.NewReader(os.Stdin)
	printWelcome()

	var lastAIResponse string
	var lastQuery string
	var lastCodeBlocks []string
	var lastFromShow bool
	var lastFromResearch bool
	var lastFromCode bool

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

		// Research-specific feedback path
		if lastAIResponse != "" && lastFromResearch && isResearchFeedbackInput(input) {
			if input == "s" {
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
				lastCodeBlocks = nil
				lastFromShow = false
				lastFromResearch = false
			} else if input == "n" {
				fmt.Println("Regenerating research recommendation...")
				fmt.Println()
				researchArgs := lastResearchQuery
				resp, uid, fpath := handleResearchCommand(researchArgs)
				if resp != "" {
					lastAIResponse = resp
					lastResearchUID = uid
					lastResearchFilePath = fpath
					lastCodeBlocks = nil
					fmt.Println(buildResearchFeedbackPrompt())
				} else {
					fmt.Println("Failed to regenerate recommendation.")
					fmt.Println()
					lastAIResponse = ""
					lastQuery = ""
					lastCodeBlocks = nil
					lastFromResearch = false
				}
			}
			continue
		}

		// If lastFromResearch is true but input is NOT a research feedback input,
		// clear the research state and fall through to process as new command
		if lastFromResearch && lastAIResponse != "" && !isResearchFeedbackInput(input) {
			lastAIResponse = ""
			lastQuery = ""
			lastCodeBlocks = nil
			lastFromResearch = false
			// Fall through to process as new command (don't continue)
		}

		// Code-specific feedback path
		if lastAIResponse != "" && lastFromCode && isCodeFeedbackInput(input, len(lastCodeBlocks)) {
			if input == "s" {
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
				lastCodeBlocks = nil
				lastFromCode = false
			} else if input == "n" {
				fmt.Println("Regenerating code proposal...")
				fmt.Println()
				resp, uid, fpath := handleCodeCommand(lastCodeQuery)
				if resp != "" {
					lastAIResponse = resp
					lastCodeUID = uid
					lastCodeFilePath = fpath
					lastCodeBlocks = ExtractCodeBlocks(resp)
					fmt.Println(buildCodeFeedbackPrompt(len(lastCodeBlocks)))
				} else {
					fmt.Println("Failed to regenerate proposal.")
					fmt.Println()
					lastAIResponse = ""
					lastQuery = ""
					lastCodeBlocks = nil
					lastFromCode = false
				}
			} else if strings.HasPrefix(input, "x") {
				// EXECUTE phase: create files from code blocks
				arg := strings.TrimSpace(strings.TrimPrefix(input, "x"))
				codeReader := bufio.NewReader(os.Stdin)
				if arg == "" {
					switch len(lastCodeBlocks) {
					case 0:
						fmt.Println("No code blocks to create files from.")
						fmt.Println(buildCodeFeedbackPrompt(len(lastCodeBlocks)))
					default:
						// Create files from all blocks (or single block)
						handleCodeExecute(lastAIResponse, lastCodeBlocks, -1, codeReader)
						fmt.Println(buildCodeFeedbackPrompt(len(lastCodeBlocks)))
					}
				} else {
					n, usageMsg := parseExecuteArg(arg)
					if usageMsg != "" {
						fmt.Println(usageMsg)
						fmt.Println(buildCodeFeedbackPrompt(len(lastCodeBlocks)))
					} else if errMsg := validateExecuteIndex(n, len(lastCodeBlocks)); errMsg != "" {
						fmt.Println(errMsg)
						fmt.Println(buildCodeFeedbackPrompt(len(lastCodeBlocks)))
					} else {
						// Create file from specific block
						handleCodeExecute(lastAIResponse, lastCodeBlocks, n-1, codeReader)
						fmt.Println(buildCodeFeedbackPrompt(len(lastCodeBlocks)))
					}
				}
			} else if n, err := strconv.Atoi(input); err == nil && len(lastCodeBlocks) > 0 {
				// Bare numeric input — create file from specific block
				codeReader := bufio.NewReader(os.Stdin)
				if errMsg := validateExecuteIndex(n, len(lastCodeBlocks)); errMsg != "" {
					fmt.Println(errMsg)
				} else {
					handleCodeExecute(lastAIResponse, lastCodeBlocks, n-1, codeReader)
				}
				fmt.Println(buildCodeFeedbackPrompt(len(lastCodeBlocks)))
			}
			continue
		}

		// If lastFromCode is true but input is NOT a code feedback input,
		// clear the code state and fall through to process as new command
		if lastFromCode && lastAIResponse != "" && !isCodeFeedbackInput(input, len(lastCodeBlocks)) {
			lastAIResponse = ""
			lastQuery = ""
			lastCodeBlocks = nil
			lastFromCode = false
			// Fall through to process as new command
		}

		// Feedback on last AI response
		if lastAIResponse != "" && isFeedbackInput(input, len(lastCodeBlocks)) {
			if input == "s" && !lastFromShow {
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
				lastCodeBlocks = nil
				lastFromShow = false
			} else if input == "n" {
				fmt.Println("Regenerating response...")
				fmt.Println()
				aiResp := regenerateAIResponse(lastQuery)
				if aiResp != "" {
					lastAIResponse = aiResp
					lastCodeBlocks = ExtractCodeBlocks(aiResp)
					lastFromShow = false
					fmt.Println(buildFeedbackPrompt(len(lastCodeBlocks), false))
				} else {
					fmt.Println("Failed to regenerate response.")
					fmt.Println()
					lastAIResponse = ""
					lastQuery = ""
					lastCodeBlocks = nil
					lastFromShow = false
				}
			} else if strings.HasPrefix(input, "x") {
				arg := strings.TrimSpace(strings.TrimPrefix(input, "x"))
				if arg == "" {
					// bare "x"
					switch len(lastCodeBlocks) {
					case 0:
						// no code blocks — re-display prompt
						fmt.Println(buildFeedbackPrompt(len(lastCodeBlocks), lastFromShow))
					case 1:
						handleRunCommand(lastCodeBlocks[0])
						fmt.Println(buildFeedbackPrompt(len(lastCodeBlocks), lastFromShow))
					default:
						fmt.Printf("Multiple code blocks found. Type 1 to %d to execute, or x <number>.\n", len(lastCodeBlocks))
						fmt.Println(buildFeedbackPrompt(len(lastCodeBlocks), lastFromShow))
					}
				} else {
					n, usageMsg := parseExecuteArg(arg)
					if usageMsg != "" {
						fmt.Println(usageMsg)
						fmt.Println(buildFeedbackPrompt(len(lastCodeBlocks), lastFromShow))
					} else if errMsg := validateExecuteIndex(n, len(lastCodeBlocks)); errMsg != "" {
						fmt.Println(errMsg)
						fmt.Println(buildFeedbackPrompt(len(lastCodeBlocks), lastFromShow))
					} else {
						handleRunCommand(lastCodeBlocks[n-1])
						fmt.Println(buildFeedbackPrompt(len(lastCodeBlocks), lastFromShow))
					}
				}
			} else if n, err := strconv.Atoi(input); err == nil && len(lastCodeBlocks) > 0 {
				// Bare numeric input — direct block execution shortcut
				if errMsg := validateExecuteIndex(n, len(lastCodeBlocks)); errMsg != "" {
					fmt.Println(errMsg)
				} else {
					handleRunCommand(lastCodeBlocks[n-1])
				}
				fmt.Println(buildFeedbackPrompt(len(lastCodeBlocks), lastFromShow))
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
			lastFromResearch = strings.HasPrefix(input, "/research")
			lastFromShow = strings.HasPrefix(input, "/show")
			lastFromCode = strings.HasPrefix(input, "/code")
			if lastFromResearch {
				lastQuery = input
				lastCodeBlocks = nil
				fmt.Println(buildResearchFeedbackPrompt())
			} else if lastFromCode {
				lastQuery = input
				lastCodeBlocks = ExtractCodeBlocks(aiResp)
				fmt.Println(buildCodeFeedbackPrompt(len(lastCodeBlocks)))
			} else if lastFromShow && showOriginalQuery != "" {
				lastQuery = showOriginalQuery
				showOriginalQuery = ""
				lastCodeBlocks = ExtractCodeBlocks(aiResp)
				fmt.Println(buildFeedbackPrompt(len(lastCodeBlocks), lastFromShow))
			} else {
				lastQuery = input
				lastCodeBlocks = ExtractCodeBlocks(aiResp)
				fmt.Println(buildFeedbackPrompt(len(lastCodeBlocks), lastFromShow))
			}
		} else {
			lastAIResponse = ""
			lastQuery = ""
			lastCodeBlocks = nil
			lastFromShow = false
			lastFromResearch = false
			lastFromCode = false
		}
	}
}

func printWelcome() {
	fmt.Println()
	green := "\033[32m"
	reset := "\033[0m"

	fmt.Println("  ╔════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("  ║                                                                                                                    ║")
	fmt.Printf("  ║                                        %s███████╗ ██████╗███╗   ███╗██████╗%s                                          ║\n", green, reset)
	fmt.Printf("  ║                                        %s██╔════╝██╔════╝████╗ ████║██╔══██╗%s                                         ║\n", green, reset)
	fmt.Printf("  ║                                        %s███████╗██║     ██╔████╔██║██║  ██║%s                                         ║\n", green, reset)
	fmt.Printf("  ║                                        %s╚════██║██║     ██║╚██╔╝██║██║  ██║%s                                         ║\n", green, reset)
	fmt.Printf("  ║                                        %s███████║╚██████╗██║ ╚═╝ ██║██████╔╝%s                                         ║\n", green, reset)
	fmt.Printf("  ║                                        %s╚══════╝ ╚═════╝╚═╝     ╚═╝╚═════╝%s                                          ║\n", green, reset)
	fmt.Println("  ║                                                                                                                    ║")
	fmt.Printf("  ║                                         ⚡ Interactive CLI  ·  v%-20s                              ║\n", updater.Release+" ⚡")
	fmt.Println("  ╚════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╝")
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
		return handleSlashCommand(input)
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
		fmt.Print(markdown.Render(aiResponse))
		if !strings.HasSuffix(aiResponse, "\n") {
			fmt.Println()
		}
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
		util.StartSpinner()
		defer util.StopSpinner()
		resp, _, err := ollama.Ask(query, contextResults)
		if err == nil && resp != "" {
			aiResponse = resp
			fmt.Println("🤖 AI Assistant:")
			fmt.Println("══════════════════════════════════════════════════════════════")
			fmt.Print(markdown.Render(aiResponse))
			if !strings.HasSuffix(aiResponse, "\n") {
				fmt.Println()
			}
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
		util.StartSpinner()
		defer util.StopSpinner()
		resp, _, err := gemini.Ask(query, contextResults)
		if err == nil && resp != "" {
			aiResponse = resp
			fmt.Println("🤖 AI Assistant:")
			fmt.Println("══════════════════════════════════════════════════════════════")
			fmt.Print(markdown.Render(aiResponse))
			if !strings.HasSuffix(aiResponse, "\n") {
				fmt.Println()
			}
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

// buildFeedbackPrompt returns the feedback prompt string based on the
// number of extracted code blocks. When savedResponse is true (e.g. from
// /show), the save option is omitted since the response is already stored.
func buildFeedbackPrompt(codeBlockCount int, savedResponse bool) string {
	var prompt string
	if savedResponse {
		prompt = "[n] - New answer (regenerates)"
	} else {
		prompt = "[s] - Good answer (saves to db)  |  [n] - New answer (discards)"
	}
	if codeBlockCount == 1 {
		prompt += "  |  [x] Execute"
	} else if codeBlockCount > 1 {
		prompt += "  |  Execute:"
		for i := 1; i <= codeBlockCount; i++ {
			prompt += fmt.Sprintf(" [%d]", i)
		}
	}
	return prompt
}

// isFeedbackInput returns true if the input is a valid feedback command:
// "s", "n", any input starting with "x", or a bare number when code blocks
// are available (codeBlockCount > 0).
func isFeedbackInput(input string, codeBlockCount int) bool {
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

// validateExecuteIndex returns an error message if index is out of range
// (< 1 or > total), or an empty string if valid.
func validateExecuteIndex(index int, total int) string {
	if index < 1 || index > total {
		return fmt.Sprintf("Error: block number must be between 1 and %d", total)
	}
	return ""
}

// parseExecuteArg parses the argument after "x". Returns (index, "") if
// valid number, or (0, "Usage: x or x <number>") if non-numeric.
func parseExecuteArg(arg string) (int, string) {
	n, err := strconv.Atoi(arg)
	if err != nil {
		return 0, "Usage: x or x <number>"
	}
	return n, ""
}
