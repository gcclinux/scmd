package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// GenerateEmbeddingsForAll generates embeddings for all commands that don't have them
func GenerateEmbeddingsForAll() error {
	fmt.Println()
	fmt.Println("Generating embeddings for existing commands...")
	fmt.Println("This may take a few minutes depending on the number of commands.")
	fmt.Println()

	// Initialize embedding providers
	InitGemini()
	InitOllama()

	if !IsGeminiAvailable() && !IsOllamaAvailable() {
		return fmt.Errorf("no embedding provider available (need Gemini API or Ollama)")
	}

	tableName := os.Getenv("TB_NAME")
	if tableName == "" {
		tableName = "scmd"
	}

	// Get all commands without embeddings
	query := fmt.Sprintf("SELECT id, key, data FROM %s WHERE embedding IS NULL", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("error querying database: %v", err)
	}
	defer rows.Close()

	var commands []CommandRecord
	for rows.Next() {
		var record CommandRecord
		if err := rows.Scan(&record.Id, &record.Key, &record.Data); err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}
		commands = append(commands, record)
	}

	if len(commands) == 0 {
		fmt.Println("✓ All commands already have embeddings!")
		return nil
	}

	fmt.Printf("Found %d commands without embeddings\n", len(commands))
	fmt.Println()

	successCount := 0
	failCount := 0
	provider := "Ollama"
	if !IsOllamaAvailable() {
		provider = "Gemini"
	}

	for i, cmd := range commands {
		// Show progress
		if i%10 == 0 && i > 0 {
			fmt.Printf("Progress: %d/%d (%.1f%%)\n", i, len(commands), float64(i)/float64(len(commands))*100)
		}

		// Generate embedding
		text := cmd.Key + " " + cmd.Data
		var embedding []float64
		var embErr error

		// Try Ollama first
		if IsOllamaAvailable() {
			embedding, embErr = GetEmbedding(text)
			if embErr != nil {
				// Try Gemini as fallback
				if IsGeminiAvailable() {
					embedding, embErr = GetGeminiEmbedding(text)
					if embErr == nil {
						provider = "Gemini"
					}
				}
			} else {
				provider = "Ollama"
			}
		} else if IsGeminiAvailable() {
			embedding, embErr = GetGeminiEmbedding(text)
			provider = "Gemini"
		}

		if embErr != nil {
			log.Printf("Failed to generate embedding for ID %d: %v\n", cmd.Id, embErr)
			failCount++
			continue
		}

		// Convert embedding to PostgreSQL vector format
		embeddingStr := "["
		for j, val := range embedding {
			if j > 0 {
				embeddingStr += ","
			}
			embeddingStr += fmt.Sprintf("%f", val)
		}
		embeddingStr += "]"

		// Update the command with embedding
		updateQuery := fmt.Sprintf("UPDATE %s SET embedding = $1::vector WHERE id = $2", tableName)
		_, err := db.Exec(updateQuery, embeddingStr, cmd.Id)
		if err != nil {
			log.Printf("Failed to update embedding for ID %d: %v\n", cmd.Id, err)
			failCount++
			continue
		}

		successCount++

		// Rate limiting for API calls (Gemini has rate limits)
		if IsGeminiAvailable() && i < len(commands)-1 {
			time.Sleep(100 * time.Millisecond) // 10 requests per second max
		}
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("✓ Embedding generation complete using %s!\n", provider)
	fmt.Printf("  Success: %d commands\n", successCount)
	if failCount > 0 {
		fmt.Printf("  Failed:  %d commands\n", failCount)
	}
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	return nil
}

// CheckEmbeddingStats shows statistics about embeddings in the database
func CheckEmbeddingStats() error {
	tableName := os.Getenv("TB_NAME")
	if tableName == "" {
		tableName = "scmd"
	}

	// Count total commands
	var total int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := db.QueryRow(query).Scan(&total)
	if err != nil {
		return fmt.Errorf("error counting commands: %v", err)
	}

	// Count commands with embeddings
	var withEmbeddings int
	query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE embedding IS NOT NULL", tableName)
	err = db.QueryRow(query).Scan(&withEmbeddings)
	if err != nil {
		return fmt.Errorf("error counting embeddings: %v", err)
	}

	withoutEmbeddings := total - withEmbeddings
	percentage := 0.0
	if total > 0 {
		percentage = float64(withEmbeddings) / float64(total) * 100
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("Embedding Statistics")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("Total commands:           %d\n", total)
	fmt.Printf("With embeddings:          %d (%.1f%%)\n", withEmbeddings, percentage)
	fmt.Printf("Without embeddings:       %d\n", withoutEmbeddings)
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	if withoutEmbeddings > 0 {
		fmt.Println("💡 Tip: Use --generate-embeddings to create embeddings for all commands")
		fmt.Println()
	}

	return nil
}
