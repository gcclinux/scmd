package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// OllamaConfig holds Ollama configuration
type OllamaConfig struct {
	Host  string
	Model string
	Port  string
}

// OllamaEmbeddingRequest represents the request to Ollama for embeddings
type OllamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// OllamaEmbeddingResponse represents the response from Ollama
type OllamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// OllamaChatRequest represents a chat request to Ollama
type OllamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

// OllamaMessage represents a message in the chat
type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaChatResponse represents the response from Ollama chat
type OllamaChatResponse struct {
	Model           string        `json:"model"`
	CreatedAt       string        `json:"created_at"`
	Message         OllamaMessage `json:"message"`
	Done            bool          `json:"done"`
	PromptEvalCount int           `json:"prompt_eval_count"`
	EvalCount       int           `json:"eval_count"`
}

var (
	ollamaAvailable bool
	ollamaConfig    OllamaConfig
	ollamaChecked   bool
)

// InitOllama initializes Ollama configuration and checks availability
func InitOllama() {
	ollamaConfig = OllamaConfig{
		Host:  os.Getenv("OLLAMA"),
		Model: os.Getenv("MODEL"),
		Port:  "11434", // Default Ollama port
	}

	if ollamaConfig.Host == "" {
		ollamaConfig.Host = "localhost"
	}

	if ollamaConfig.Model == "" {
		ollamaConfig.Model = "llama2"
	}

	// Check if Ollama is available
	ollamaAvailable = checkOllamaAvailability()
	ollamaChecked = true

	if ollamaAvailable {
		embeddingDim := os.Getenv("EMBEDDING_DIM")
		if embeddingDim == "" {
			embeddingDim = "384"
		}
		fmt.Printf("✓ Ollama available at %s:%s (model: %s, embedding_dim: %s)\n",
			ollamaConfig.Host, ollamaConfig.Port, ollamaConfig.Model, embeddingDim)
	}
}

// IsOllamaAvailable returns whether Ollama is available
func IsOllamaAvailable() bool {
	if !ollamaChecked {
		InitOllama()
	}
	return ollamaAvailable
}

// checkOllamaAvailability checks if Ollama server is reachable
func checkOllamaAvailability() bool {
	url := fmt.Sprintf("http://%s:%s/api/tags", ollamaConfig.Host, ollamaConfig.Port)

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetEmbedding gets an embedding vector from Ollama
func GetEmbedding(text string) ([]float64, error) {
	url := fmt.Sprintf("http://%s:%s/api/embeddings", ollamaConfig.Host, ollamaConfig.Port)

	reqBody := OllamaEmbeddingRequest{
		Model:  ollamaConfig.Model,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error calling Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var response OllamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// Get target dimension from environment
	targetDim := 384 // Default
	if dimStr := os.Getenv("EMBEDDING_DIM"); dimStr != "" {
		if dim, err := strconv.Atoi(dimStr); err == nil {
			targetDim = dim
		}
	}

	// Adjust embedding dimension if needed
	embedding := response.Embedding
	if len(embedding) > targetDim {
		// Truncate to target dimension
		embedding = embedding[:targetDim]
	} else if len(embedding) < targetDim {
		// Pad with zeros to target dimension
		padding := make([]float64, targetDim-len(embedding))
		embedding = append(embedding, padding...)
	}

	return embedding, nil
}

// SearchVectorInDB performs a vector similarity search in the database using a given embedding
func SearchVectorInDB(embedding []float64, limit int) ([]CommandRecord, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is null")
	}

	// Convert embedding to PostgreSQL vector format
	embeddingStr := "["
	for i, val := range embedding {
		if i > 0 {
			embeddingStr += ","
		}
		embeddingStr += fmt.Sprintf("%f", val)
	}
	embeddingStr += "]"

	tableName := os.Getenv("TB_NAME")
	if tableName == "" {
		tableName = "scmd"
	}

	// Use vector similarity search (cosine distance)
	query := fmt.Sprintf(`
		SELECT id, key, data, 
		       1 - (embedding <=> $1::vector) as similarity
		FROM %s
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`, tableName)

	rows, err := db.Query(query, embeddingStr, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %v", err)
	}
	defer rows.Close()

	var results []CommandRecord
	for rows.Next() {
		var record CommandRecord
		var similarity float64
		if err := rows.Scan(&record.Id, &record.Key, &record.Data, &similarity); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		results = append(results, record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// SearchWithOllama performs a vector similarity search prioritizing Ollama
func SearchWithOllama(query string, limit int) ([]CommandRecord, error) {
	var embedding []float64
	var err error

	// Priority: Ollama then Gemini
	if IsOllamaAvailable() {
		embedding, err = GetEmbedding(query)
		if err != nil {
			fmt.Printf("⚠ Ollama embedding failed: %v, trying Gemini...\n", err)
		}
	}

	if embedding == nil && IsGeminiAvailable() {
		embedding, err = GetGeminiEmbedding(query)
	}

	if embedding == nil {
		return nil, fmt.Errorf("no embedding provider available")
	}

	return SearchVectorInDB(embedding, limit)
}

// AskOllama sends a question to Ollama and gets a response
// Returns (responseText, totalTokens, error)
func AskOllama(question string, context []CommandRecord) (string, int, error) {
	if !requireAIAccess() {
		return "", 0, fmt.Errorf("AI access denied: invalid or missing API_ACCESS key")
	}
	url := fmt.Sprintf("http://%s:%s/api/chat", ollamaConfig.Host, ollamaConfig.Port)

	// Build context from command records
	contextStr := ""
	if len(context) > 0 {
		contextStr = "Here are some relevant commands from the database:\n\n"
		for i, cmd := range context {
			contextStr += fmt.Sprintf("%d. Description: %s\n   Command: %s\n\n", i+1, cmd.Data, cmd.Key)
		}
	}

	// Create the prompt
	systemPrompt := `You are a helpful assistant that helps users find and understand command-line commands. 
You have access to a database of commands. When answering questions:
1. Always start with a brief, natural introduction
2. Reference the specific commands from the context provided
3. ALWAYS format commands in code blocks with the appropriate language tag (bash, powershell, sql, docker, etc.)
4. Use triple backticks with language tags for code blocks
5. Explain what the command does after showing it
6. Be concise but informative
7. If multiple commands are relevant, show each in its own code block
8. Detect the command type and use the correct language tag (bash, powershell, postgresql, mysql, docker, kubernetes, python, etc.)`

	userPrompt := fmt.Sprintf("%s\nUser question: %s", contextStr, question)

	reqBody := OllamaChatRequest{
		Model: ollamaConfig.Model,
		Messages: []OllamaMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("error marshaling request: %v", err)
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, fmt.Errorf("error calling Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var response OllamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", 0, fmt.Errorf("error decoding response: %v", err)
	}

	totalTokens := response.PromptEvalCount + response.EvalCount
	return response.Message.Content, totalTokens, nil
}

// AskAI sends a question to the best available AI provider
// Returns (responseText, totalTokens, error)
func AskAI(question string, context []CommandRecord) (string, int, error) {
	var errs []error
	preferredAgent := strings.ToLower(os.Getenv("AGENT"))

	// Check preferred agent first
	if preferredAgent == "ollama" && IsOllamaAvailable() {
		return AskOllama(question, context)
	} else if preferredAgent == "gemini" && IsGeminiAvailable() {
		return AskGemini(question, context)
	}

	// If a preferred agent was explicitly set but failed or wasn't available, do NOT fallback
	if preferredAgent != "" {
		return "", 0, fmt.Errorf("preferred AI provider '%s' is not available or failed", preferredAgent)
	}

	// Fallback logic when no preferred agent is set
	if IsOllamaAvailable() {
		response, tokens, err := AskOllama(question, context)
		if err == nil {
			return response, tokens, nil
		}
		errs = append(errs, fmt.Errorf("Ollama failed: %v", err))
	}

	if IsGeminiAvailable() {
		response, tokens, err := AskGemini(question, context)
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
// 1. PostgreSQL plain text (keyword) search (using cleaned query)
// 2. Ollama vector search + chat
// 3. Gemini vector search + chat
func SmartSearch(query string, useEmbeddings bool) ([]CommandRecord, string, int, error) {
	// Clean the query to remove "how do i", "show me", etc.
	// This makes database keyword search much more accurate.
	cleanedQuery := extractKeywords(query)
	if cleanedQuery == "" {
		cleanedQuery = query
	}

	var results []CommandRecord
	var aiResponse string
	var aiTokens int

	// --- 1st PRIORITY: Cleaned Keyword Search ---
	jsonData, err := SearchCommands(cleanedQuery, "json")
	if err != nil {
		return nil, "", 0, err
	}
	var keywordResults []CommandRecord
	json.Unmarshal(jsonData, &keywordResults)
	scoredKeywords := ScoreCommands(keywordResults, cleanedQuery)

	// If we have very high quality matches, return them immediately
	if HasGoodMatches(scoredKeywords, 60) {
		fmt.Println("✓ Found high-quality matches in database")
		// Only include results that themselves meet the 60% threshold
		// This prevents loosely-related results from slipping through
		qualifiedScored := FilterByMinScore(scoredKeywords, 60)
		bestScored := GetBestMatches(qualifiedScored, 10)
		for _, s := range bestScored {
			results = append(results, s.Record)
		}

		// DO NOT use AI here to save power when high-quality matches are found
		aiResponse = ""
		return results, aiResponse, 0, nil
	}

	if !useEmbeddings {
		// Not using AI, return whatever decent keyword matches we found
		bestScored := GetBestMatches(scoredKeywords, 10)
		for _, s := range bestScored {
			if s.Score >= 25 {
				results = append(results, s.Record)
			}
		}
		return results, "", 0, nil
	}

	preferredAgent := strings.ToLower(os.Getenv("AGENT"))

	// Helper function for Ollama search logic
	tryOllama := func() bool {
		if !IsOllamaAvailable() {
			return false
		}
		fmt.Println("⚠ Trying Ollama (Vector Search & Chat)...")
		emb, err := GetEmbedding(query)
		if err == nil {
			vResults, err := SearchVectorInDB(emb, 10)
			if err == nil && len(vResults) > 0 {
				scoredVector := ScoreCommands(vResults, cleanedQuery)
				var filteredVector []CommandRecord
				for _, s := range scoredVector {
					if s.Score > 0 {
						filteredVector = append(filteredVector, s.Record)
					}
				}
				if len(filteredVector) > 0 {
					results = filteredVector
					res, tok, err := AskOllama(query, results)
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
			var contextResults []CommandRecord
			for _, s := range GetBestMatches(scoredKeywords, 5) {
				if s.Score > 0 {
					contextResults = append(contextResults, s.Record)
				}
			}
			res, tok, err := AskOllama(query, contextResults)
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

	// Helper function for Gemini search logic
	tryGemini := func() bool {
		if !IsGeminiAvailable() {
			return false
		}
		fmt.Println("⚠ Trying Gemini (Vector Search & Chat)...")
		emb, err := GetGeminiEmbedding(query)
		if err == nil {
			vResults, err := SearchVectorInDB(emb, 10)
			if err == nil && len(vResults) > 0 {
				scoredVector := ScoreCommands(vResults, cleanedQuery)
				var filteredVector []CommandRecord
				for _, s := range scoredVector {
					if s.Score > 0 {
						filteredVector = append(filteredVector, s.Record)
					}
				}
				if len(filteredVector) > 0 {
					results = filteredVector
					res, tok, err := AskGemini(query, results)
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
			var contextResults []CommandRecord
			for _, s := range GetBestMatches(scoredKeywords, 5) {
				if s.Score > 0 {
					contextResults = append(contextResults, s.Record)
				}
			}
			res, tok, err := AskGemini(query, contextResults)
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

	// --- 2nd PRIORITY: Preferred Agent ---
	success := false
	if preferredAgent == "ollama" {
		success = tryOllama()
	} else if preferredAgent == "gemini" {
		success = tryGemini()
	}

	if success {
		return results, aiResponse, aiTokens, nil
	}

	// --- 3rd PRIORITY: Fallback Agent ---
	// Only fallback if no explicit preferredAgent was set
	if preferredAgent == "" {
		if tryOllama() {
			return results, aiResponse, aiTokens, nil
		}
		if tryGemini() {
			return results, aiResponse, aiTokens, nil
		}
	}

	// --- 4th PRIORITY: Last Resort AI chat with no context ---
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

	// FINAL CLEANUP: Make sure we only return results that actually have some keyword relevance
	finalResults := []CommandRecord{}
	scoredFinal := ScoreCommands(results, cleanedQuery)
	for _, s := range scoredFinal {
		if s.Score > 0 {
			finalResults = append(finalResults, s.Record)
		}
	}

	return finalResults, aiResponse, aiTokens, nil
}
