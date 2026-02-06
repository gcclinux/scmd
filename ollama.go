package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
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
	Model     string        `json:"model"`
	CreatedAt string        `json:"created_at"`
	Message   OllamaMessage `json:"message"`
	Done      bool          `json:"done"`
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

// SearchWithOllama performs a vector similarity search using Ollama embeddings
func SearchWithOllama(query string, limit int) ([]CommandRecord, error) {
	// Get embedding for the query
	embedding, err := GetEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("error getting embedding: %v", err)
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
	query = fmt.Sprintf(`
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

// AskOllama sends a question to Ollama and gets a response
func AskOllama(question string, context []CommandRecord) (string, error) {
	url := fmt.Sprintf("http://%s:%s/api/chat", ollamaConfig.Host, ollamaConfig.Port)

	// Build context from command records
	contextStr := "Here are some relevant commands from the database:\n\n"
	for i, cmd := range context {
		contextStr += fmt.Sprintf("%d. Description: %s\n   Command: %s\n\n", i+1, cmd.Data, cmd.Key)
	}

	// Create the prompt
	systemPrompt := `You are a helpful assistant that helps users find and understand command-line commands. 
You have access to a database of commands. When answering questions:
1. Reference the specific commands from the context provided
2. Explain what the command does
3. Be concise but informative
4. If multiple commands are relevant, mention them
5. Format commands in code blocks using backticks`

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
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error calling Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var response OllamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	return response.Message.Content, nil
}

// SmartSearch performs an intelligent search using Ollama if available, falls back to traditional search
func SmartSearch(query string, useOllama bool) ([]CommandRecord, string, error) {
	var results []CommandRecord
	var aiResponse string

	// First, always do traditional search to get candidates
	jsonData, err := SearchCommands(query, "json")
	if err != nil {
		return nil, "", err
	}

	var traditionalResults []CommandRecord
	json.Unmarshal(jsonData, &traditionalResults)

	// Score the traditional results
	scored := ScoreCommands(traditionalResults, query)

	// Check if we have good matches (50%+ score)
	hasGoodMatches := HasGoodMatches(scored, 50)

	if hasGoodMatches {
		// We have good traditional matches, use them
		fmt.Println("✓ Found good matches in database")

		// Get best matches (top 10)
		bestScored := GetBestMatches(scored, 10)
		for _, s := range bestScored {
			results = append(results, s.Record)
		}

		// Only invoke AI if Ollama is available and we have results
		if useOllama && IsOllamaAvailable() && len(results) > 0 {
			// Use top 5 for AI context
			contextResults := results
			if len(contextResults) > 5 {
				contextResults = contextResults[:5]
			}
			var aiErr error
			aiResponse, aiErr = AskOllama(query, contextResults)
			if aiErr != nil {
				fmt.Printf("⚠ Ollama chat failed: %v\n", aiErr)
				aiResponse = ""
			}
		}
	} else if useOllama && IsOllamaAvailable() {
		// No good traditional matches, try vector search
		fmt.Println("⚠ No strong keyword matches, trying vector search...")

		vectorResults, vectorErr := SearchWithOllama(query, 10)
		if vectorErr != nil {
			// Vector search failed, use whatever traditional results we have
			fmt.Printf("⚠ Vector search failed: %v\n", vectorErr)
			bestScored := GetBestMatches(scored, 10)
			for _, s := range bestScored {
				results = append(results, s.Record)
			}
		} else {
			// Combine vector and traditional results
			results = vectorResults

			// Add traditional results that aren't in vector results
			vectorIDs := make(map[int]bool)
			for _, r := range vectorResults {
				vectorIDs[r.Id] = true
			}

			for _, s := range scored {
				if !vectorIDs[s.Record.Id] && len(results) < 10 {
					results = append(results, s.Record)
				}
			}

			// Get AI response
			if len(results) > 0 {
				contextResults := results
				if len(contextResults) > 5 {
					contextResults = contextResults[:5]
				}
				var aiErr error
				aiResponse, aiErr = AskOllama(query, contextResults)
				if aiErr != nil {
					fmt.Printf("⚠ Ollama chat failed: %v\n", aiErr)
					aiResponse = ""
				}
			}
		}
	} else {
		// No Ollama, just use traditional results
		bestScored := GetBestMatches(scored, 10)
		for _, s := range bestScored {
			results = append(results, s.Record)
		}
	}

	return results, aiResponse, nil
}
