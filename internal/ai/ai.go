package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gcclinux/scmd/internal/ai/gemini"
	"github.com/gcclinux/scmd/internal/ai/ollama"
	"github.com/gcclinux/scmd/internal/database"
	"github.com/gcclinux/scmd/internal/search"
)

// InitProviders initializes both AI providers and prints the active provider status.
func InitProviders() {
	gemini.Init()
	ollama.Init()

	preferred := strings.ToLower(os.Getenv("AGENT"))
	embeddingDim := os.Getenv("EMBEDDING_DIM")
	if embeddingDim == "" {
		embeddingDim = "384"
	}

	switch preferred {
	case "ollama":
		if ollama.IsAvailable() {
			fmt.Printf("✓ Ollama (host: %s, model: %s, embeddings: %s, dim: %s)\n",
				ollama.Host(), ollama.ModelName(), ollama.EmbeddingModelName(), embeddingDim)
		} else {
			fmt.Println("⚠ Preferred agent 'ollama' is not available")
		}
	case "gemini":
		if gemini.IsAvailable() {
			fmt.Printf("✓ Gemini (chat: %s, embeddings: %s, dim: %s)\n",
				gemini.ModelName(), gemini.EmbeddingModelName(), embeddingDim)
		} else {
			fmt.Println("⚠ Preferred agent 'gemini' is not available")
		}
	default:
		// No preference set — show whichever is available (Ollama first)
		if ollama.IsAvailable() {
			fmt.Printf("✓ Ollama (host: %s, model: %s, embeddings: %s, dim: %s)\n",
				ollama.Host(), ollama.ModelName(), ollama.EmbeddingModelName(), embeddingDim)
		} else if gemini.IsAvailable() {
			fmt.Printf("✓ Gemini (chat: %s, embeddings: %s, dim: %s)\n",
				gemini.ModelName(), gemini.EmbeddingModelName(), embeddingDim)
		} else {
			fmt.Println("⚠ No AI provider available")
		}
	}
}

// GetBestEmbedding tries Ollama first, then Gemini to generate an embedding.
func GetBestEmbedding(text string) ([]float64, error) {
	if ollama.IsAvailable() {
		emb, err := ollama.GetEmbedding(text)
		if err == nil {
			return emb, nil
		}
	}
	if gemini.IsAvailable() {
		return gemini.GetEmbedding(text)
	}
	return nil, fmt.Errorf("no embedding provider available")
}

// AskAI sends a question to the best available AI provider.
// Returns (responseText, totalTokens, error).
func AskAI(question string, context []database.CommandRecord) (string, int, error) {
	var errs []error
	preferredAgent := strings.ToLower(os.Getenv("AGENT"))

	if preferredAgent == "ollama" && ollama.IsAvailable() {
		return ollama.Ask(question, context)
	} else if preferredAgent == "gemini" && gemini.IsAvailable() {
		return gemini.Ask(question, context)
	}

	if preferredAgent != "" {
		return "", 0, fmt.Errorf("preferred AI provider '%s' is not available or failed", preferredAgent)
	}

	if ollama.IsAvailable() {
		response, tokens, err := ollama.Ask(question, context)
		if err == nil {
			return response, tokens, nil
		}
		errs = append(errs, fmt.Errorf("Ollama failed: %v", err))
	}

	if gemini.IsAvailable() {
		response, tokens, err := gemini.Ask(question, context)
		if err == nil {
			return response, tokens, nil
		}
		errs = append(errs, fmt.Errorf("Gemini failed: %v", err))
	}

	if len(errs) > 0 {
		return "", 0, fmt.Errorf("all AI providers failed: %v", errs)
	}
	return "", 0, fmt.Errorf("no AI provider available")
}

// SmartSearch performs an intelligent search following the priority:
// 1. PostgreSQL keyword search
// 2. Vector search + AI chat
// 3. Pure AI chat
func SmartSearch(query string, useEmbeddings bool) ([]database.CommandRecord, string, int, error) {
	cleanedQuery := search.ExtractKeywords(query)
	if cleanedQuery == "" {
		cleanedQuery = query
	}

	var results []database.CommandRecord
	var aiResponse string
	var aiTokens int

	// 1st PRIORITY: Keyword Search
	jsonData, err := database.SearchCommands(cleanedQuery, "json")
	if err != nil {
		return nil, "", 0, err
	}
	var keywordResults []database.CommandRecord
	json.Unmarshal(jsonData, &keywordResults)
	scoredKeywords := search.ScoreCommands(keywordResults, cleanedQuery)

	if search.HasGoodMatches(scoredKeywords, 60) {
		fmt.Println("✓ Found high-quality matches in database")
		qualifiedScored := search.FilterByMinScore(scoredKeywords, 60)
		bestScored := search.GetBestMatches(qualifiedScored, 10)
		for _, s := range bestScored {
			results = append(results, s.Record)
		}
		return results, "", 0, nil
	}

	if !useEmbeddings {
		bestScored := search.GetBestMatches(scoredKeywords, 10)
		for _, s := range bestScored {
			if s.Score >= 25 {
				results = append(results, s.Record)
			}
		}
		return results, "", 0, nil
	}

	preferredAgent := strings.ToLower(os.Getenv("AGENT"))

	tryOllama := func() bool {
		if !ollama.IsAvailable() {
			return false
		}
		fmt.Println("⚠ Trying Ollama (Vector Search & Chat)...")
		emb, err := ollama.GetEmbedding(query)
		if err == nil {
			vResults, err := database.SearchByVector(emb, 10)
			if err == nil && len(vResults) > 0 {
				scoredVector := search.ScoreCommands(vResults, cleanedQuery)
				var filteredVector []database.CommandRecord
				for _, s := range scoredVector {
					if s.Score > 0 {
						filteredVector = append(filteredVector, s.Record)
					}
				}
				if len(filteredVector) > 0 {
					results = filteredVector
					res, tok, err := ollama.Ask(query, results)
					if err == nil && res != "" {
						aiResponse = res
						aiTokens = tok
						return true
					} else if err != nil {
						fmt.Printf("⚠ Ollama API error: %v\n", err)
					}
				}
			}
		}

		if aiResponse == "" && len(scoredKeywords) > 0 {
			var contextResults []database.CommandRecord
			for _, s := range search.GetBestMatches(scoredKeywords, 5) {
				if s.Score > 0 {
					contextResults = append(contextResults, s.Record)
				}
			}
			res, tok, err := ollama.Ask(query, contextResults)
			if err == nil && res != "" {
				results = contextResults
				aiResponse = res
				aiTokens = tok
				return true
			} else if err != nil {
				fmt.Printf("⚠ Ollama API error: %v\n", err)
			}
		}
		return false
	}

	tryGemini := func() bool {
		if !gemini.IsAvailable() {
			return false
		}
		fmt.Println("⚠ Trying Gemini (Vector Search & Chat)...")
		emb, err := gemini.GetEmbedding(query)
		if err == nil {
			vResults, err := database.SearchByVector(emb, 10)
			if err == nil && len(vResults) > 0 {
				scoredVector := search.ScoreCommands(vResults, cleanedQuery)
				var filteredVector []database.CommandRecord
				for _, s := range scoredVector {
					if s.Score > 0 {
						filteredVector = append(filteredVector, s.Record)
					}
				}
				if len(filteredVector) > 0 {
					results = filteredVector
					res, tok, err := gemini.Ask(query, results)
					if err == nil && res != "" {
						aiResponse = res
						aiTokens = tok
						return true
					} else if err != nil {
						fmt.Printf("⚠ Gemini API error: %v\n", err)
					}
				}
			}
		}

		if aiResponse == "" && len(scoredKeywords) > 0 {
			var contextResults []database.CommandRecord
			for _, s := range search.GetBestMatches(scoredKeywords, 5) {
				if s.Score > 0 {
					contextResults = append(contextResults, s.Record)
				}
			}
			res, tok, err := gemini.Ask(query, contextResults)
			if err == nil && res != "" {
				results = contextResults
				aiResponse = res
				aiTokens = tok
				return true
			} else if err != nil {
				fmt.Printf("⚠ Gemini API error: %v\n", err)
			}
		}
		return false
	}

	success := false
	if preferredAgent == "ollama" {
		success = tryOllama()
	} else if preferredAgent == "gemini" {
		success = tryGemini()
	}

	if success {
		return results, aiResponse, aiTokens, nil
	}

	if preferredAgent == "" {
		if tryOllama() {
			return results, aiResponse, aiTokens, nil
		}
		if tryGemini() {
			return results, aiResponse, aiTokens, nil
		}
	}

	// Last resort: AI chat with no context
	if aiResponse == "" {
		resp, tok, err := AskAI(query, nil)
		if err != nil {
			fmt.Printf("⚠ AskAI Error: %v\n", err)
			aiResponse = fmt.Sprintf("⚠️ **AI Provider Error**\n\n```text\n%v\n```\n\nPlease check your configuration, model name, and API keys.", err)
		} else {
			aiResponse = resp
			aiTokens = tok
		}
	}

	finalResults := []database.CommandRecord{}
	scoredFinal := search.ScoreCommands(results, cleanedQuery)
	for _, s := range scoredFinal {
		if s.Score > 0 {
			finalResults = append(finalResults, s.Record)
		}
	}

	return finalResults, aiResponse, aiTokens, nil
}

// GetProviderLabel returns a label describing the active AI provider.
func GetProviderLabel() string {
	preferredAgent := strings.ToLower(os.Getenv("AGENT"))

	if preferredAgent == "ollama" && ollama.IsAvailable() {
		return fmt.Sprintf("Ollama (%s)", ollama.ModelName())
	} else if preferredAgent == "gemini" && gemini.IsAvailable() {
		return fmt.Sprintf("Gemini (%s)", gemini.ModelName())
	}

	if ollama.IsAvailable() {
		return fmt.Sprintf("Ollama (%s)", ollama.ModelName())
	}
	if gemini.IsAvailable() {
		return fmt.Sprintf("Gemini (%s)", gemini.ModelName())
	}
	return "None"
}
