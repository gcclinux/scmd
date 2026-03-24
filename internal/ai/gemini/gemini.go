package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gcclinux/scmd/internal/database"
)

// Config holds Gemini API configuration.
type Config struct {
	APIKey         string
	Model          string
	EmbeddingModel string
}

type embeddingRequest struct {
	Content struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"content"`
}

type embeddingResponse struct {
	Embedding struct {
		Values []float64 `json:"values"`
	} `json:"embedding"`
}

type chatRequest struct {
	Contents []chatContent `json:"contents"`
}

type chatContent struct {
	Role  string     `json:"role,omitempty"`
	Parts []chatPart `json:"parts"`
}

type chatPart struct {
	Text string `json:"text"`
}

type chatResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

var (
	available bool
	cfg       Config
	checked   bool
)

// Init initializes Gemini configuration and checks availability.
func Init() {
	cfg = Config{
		APIKey:         os.Getenv("GEMINIAPI"),
		Model:          os.Getenv("GEMINIMODEL"),
		EmbeddingModel: os.Getenv("GEMINI_EMBEDDING_MODEL"),
	}

	if cfg.APIKey == "" {
		available = false
		checked = true
		return
	}

	if cfg.Model == "" {
		cfg.Model = "gemini-2.0-flash-exp"
	}
	if cfg.EmbeddingModel == "" {
		cfg.EmbeddingModel = "embedding-001"
	}

	available = cfg.APIKey != ""
	checked = true
}

// IsAvailable returns whether Gemini API is available.
func IsAvailable() bool {
	if !checked {
		Init()
	}
	return available
}

// GetEmbedding gets an embedding vector from Gemini API.
func GetEmbedding(text string) ([]float64, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s",
		cfg.EmbeddingModel, cfg.APIKey)

	reqBody := embeddingRequest{}
	reqBody.Content.Parts = []struct {
		Text string `json:"text"`
	}{
		{Text: text},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error calling Gemini API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	targetDim := 384
	if dimStr := os.Getenv("EMBEDDING_DIM"); dimStr != "" {
		if dim, err := strconv.Atoi(dimStr); err == nil {
			targetDim = dim
		}
	}

	embedding := response.Embedding.Values
	if len(embedding) > targetDim {
		embedding = embedding[:targetDim]
	} else if len(embedding) < targetDim {
		padding := make([]float64, targetDim-len(embedding))
		embedding = append(embedding, padding...)
	}

	return embedding, nil
}

// Ask sends a question to Gemini and gets a response.
// Returns (responseText, totalTokens, error).
func Ask(question string, context []database.CommandRecord) (string, int, error) {
	if !IsAvailable() {
		return "", 0, fmt.Errorf("Gemini API is not available")
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		cfg.Model, cfg.APIKey)

	contextStr := ""
	if len(context) > 0 {
		contextStr = "Here are some relevant commands from the database:\n\n"
		for i, cmd := range context {
			contextStr += fmt.Sprintf("%d. Description: %s\n   Command: %s\n\n", i+1, cmd.Data, cmd.Key)
		}
	}

	systemPrompt := `You are a helpful assistant that helps users find and understand command-line commands. 
You have access to a database of commands. When answering questions:
1. Always start with a brief, natural introduction
2. Reference the specific commands from the context provided (if any)
3. ALWAYS format commands in code blocks with the appropriate language tag (bash, powershell, sql, docker, etc.)
4. Use triple backticks with language tags for code blocks
5. Explain what the command does after showing it
6. Be concise but informative
7. If multiple commands are relevant, show each in its own code block
8. Detect the command type and use the correct language tag (bash, powershell, postgresql, mysql, docker, kubernetes, python, etc.)
9. If no commands are relevant or no context provided, provide the best answer you can based on your knowledge.`

	userPrompt := fmt.Sprintf("%s\nUser question: %s", contextStr, question)

	reqBody := chatRequest{
		Contents: []chatContent{
			{
				Role: "user",
				Parts: []chatPart{
					{Text: systemPrompt + "\n\n" + userPrompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("error marshaling request: %v", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, fmt.Errorf("error calling Gemini API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", 0, fmt.Errorf("error decoding response: %v", err)
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "", 0, fmt.Errorf("empty response from Gemini API")
	}

	return response.Candidates[0].Content.Parts[0].Text, response.UsageMetadata.TotalTokenCount, nil
}

// ModelName returns the configured chat model name.
func ModelName() string {
	return cfg.Model
}

// EmbeddingModelName returns the configured embedding model name.
func EmbeddingModelName() string {
	return cfg.EmbeddingModel
}
