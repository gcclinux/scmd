package ollama

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

// Config holds Ollama configuration.
type Config struct {
	Host           string
	Model          string
	EmbeddingModel string
	Port           string
}

type embeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Model           string  `json:"model"`
	CreatedAt       string  `json:"created_at"`
	Message         message `json:"message"`
	Done            bool    `json:"done"`
	PromptEvalCount int     `json:"prompt_eval_count"`
	EvalCount       int     `json:"eval_count"`
}

var (
	available bool
	cfg       Config
	checked   bool
)

// Init initializes Ollama configuration and checks availability.
func Init() {
	cfg = Config{
		Host:           os.Getenv("OLLAMA"),
		Model:          os.Getenv("MODEL"),
		EmbeddingModel: os.Getenv("EMBEDDING_MODEL"),
		Port:           "11434",
	}

	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Model == "" {
		cfg.Model = "llama2"
	}
	if cfg.EmbeddingModel == "" {
		cfg.EmbeddingModel = cfg.Model
	}

	available = checkAvailability()
	checked = true
}

// IsAvailable returns whether Ollama is available.
func IsAvailable() bool {
	if !checked {
		Init()
	}
	return available
}

func checkAvailability() bool {
	url := fmt.Sprintf("http://%s:%s/api/tags", cfg.Host, cfg.Port)
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// GetEmbedding gets an embedding vector from Ollama.
func GetEmbedding(text string) ([]float64, error) {
	url := fmt.Sprintf("http://%s:%s/api/embeddings", cfg.Host, cfg.Port)

	reqBody := embeddingRequest{
		Model:  cfg.EmbeddingModel,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error calling Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
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

	embedding := response.Embedding
	if len(embedding) > targetDim {
		embedding = embedding[:targetDim]
	} else if len(embedding) < targetDim {
		padding := make([]float64, targetDim-len(embedding))
		embedding = append(embedding, padding...)
	}

	return embedding, nil
}

// Ask sends a question to Ollama and gets a response.
// Returns (responseText, totalTokens, error).
func Ask(question string, context []database.CommandRecord) (string, int, error) {
	url := fmt.Sprintf("http://%s:%s/api/chat", cfg.Host, cfg.Port)

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
2. Reference the specific commands from the context provided
3. ALWAYS format commands in code blocks with the appropriate language tag (bash, powershell, sql, docker, etc.)
4. Use triple backticks with language tags for code blocks
5. Explain what the command does after showing it
6. Be concise but informative
7. If multiple commands are relevant, show each in its own code block
8. Detect the command type and use the correct language tag (bash, powershell, postgresql, mysql, docker, kubernetes, python, etc.)`

	userPrompt := fmt.Sprintf("%s\nUser question: %s", contextStr, question)

	reqBody := chatRequest{
		Model: cfg.Model,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("error marshaling request: %v", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, fmt.Errorf("error calling Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var response chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", 0, fmt.Errorf("error decoding response: %v", err)
	}

	totalTokens := response.PromptEvalCount + response.EvalCount
	return response.Message.Content, totalTokens, nil
}

// ModelName returns the configured model name.
func ModelName() string {
	return cfg.Model
}

// EmbeddingModelName returns the configured embedding model name.
func EmbeddingModelName() string {
	return cfg.EmbeddingModel
}

// Host returns the configured host.
func Host() string {
	return cfg.Host
}

// Port returns the configured port.
func Port() string {
	return cfg.Port
}
