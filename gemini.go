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

// GeminiConfig holds Gemini API configuration
type GeminiConfig struct {
	APIKey         string
	Model          string // Chat model (e.g., gemini-2.5-flash)
	EmbeddingModel string // Embedding model (e.g., text-embedding-004)
}

// GeminiEmbeddingRequest represents the request to Gemini for embeddings
type GeminiEmbeddingRequest struct {
	Content struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"content"`
}

// GeminiEmbeddingResponse represents the response from Gemini
type GeminiEmbeddingResponse struct {
	Embedding struct {
		Values []float64 `json:"values"`
	} `json:"embedding"`
}

var (
	geminiAvailable bool
	geminiConfig    GeminiConfig
	geminiChecked   bool
)

// InitGemini initializes Gemini configuration and checks availability
func InitGemini() {
	geminiConfig = GeminiConfig{
		APIKey:         os.Getenv("GEMINIAPI"),
		Model:          os.Getenv("GEMINIMODEL"),
		EmbeddingModel: os.Getenv("GEMINI_EMBEDDING_MODEL"),
	}

	if geminiConfig.APIKey == "" {
		geminiAvailable = false
		geminiChecked = true
		return
	}

	// Set default chat model if not specified
	if geminiConfig.Model == "" {
		geminiConfig.Model = "gemini-2.0-flash-exp"
	}

	// Set default embedding model if not specified
	if geminiConfig.EmbeddingModel == "" {
		geminiConfig.EmbeddingModel = "embedding-001"
	}

	// Check if Gemini API is available
	geminiAvailable = checkGeminiAvailability()
	geminiChecked = true

	if geminiAvailable {
		embeddingDim := os.Getenv("EMBEDDING_DIM")
		if embeddingDim == "" {
			embeddingDim = "384"
		}
		fmt.Printf("✓ Gemini API available (chat: %s, embeddings: %s, dim: %s)\n",
			geminiConfig.Model, geminiConfig.EmbeddingModel, embeddingDim)
	}
}

// IsGeminiAvailable returns whether Gemini API is available
func IsGeminiAvailable() bool {
	if !geminiChecked {
		InitGemini()
	}
	return geminiAvailable
}

// checkGeminiAvailability checks if Gemini API is reachable
func checkGeminiAvailability() bool {
	if geminiConfig.APIKey == "" {
		return false
	}

	// Simple check - if we have an API key, assume it's available
	// We'll validate on first actual use
	return true
}

// GetGeminiEmbedding gets an embedding vector from Gemini API
func GetGeminiEmbedding(text string) ([]float64, error) {
	// Use the embedding model, not the chat model
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s",
		geminiConfig.EmbeddingModel, geminiConfig.APIKey)

	reqBody := GeminiEmbeddingRequest{}
	reqBody.Content.Parts = []struct {
		Text string `json:"text"`
	}{
		{Text: text},
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
		return nil, fmt.Errorf("error calling Gemini API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response GeminiEmbeddingResponse
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
	embedding := response.Embedding.Values
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
// GeminiChatRequest represents a chat request to Gemini
type GeminiChatRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent represents content in the chat
type GeminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of the content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiChatResponse represents the response from Gemini chat
type GeminiChatResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// AskGemini sends a question to Gemini and gets a response
func AskGemini(question string, context []CommandRecord) (string, error) {
	if !IsGeminiAvailable() {
		return "", fmt.Errorf("Gemini API is not available")
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		geminiConfig.Model, geminiConfig.APIKey)

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
1. Reference the specific commands from the context provided (if any)
2. Explain what the command does
3. Be concise but informative
4. If multiple commands are relevant, mention them
5. Format commands in code blocks using backticks
6. If no commands are relevant or no context provided, provide the best answer you can based on your knowledge.`

	userPrompt := fmt.Sprintf("%s\nUser question: %s", contextStr, question)

	reqBody := GeminiChatRequest{
		Contents: []GeminiContent{
			{
				Role: "user",
				Parts: []GeminiPart{
					{Text: systemPrompt + "\n\n" + userPrompt},
				},
			},
		},
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
		return "", fmt.Errorf("error calling Gemini API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response GeminiChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini API")
	}

	return response.Candidates[0].Content.Parts[0].Text, nil
}
