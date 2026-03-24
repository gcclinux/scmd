package ai

import (
	"fmt"
	"log"
	"time"

	"github.com/gcclinux/scmd/internal/ai/gemini"
	"github.com/gcclinux/scmd/internal/ai/ollama"
	"github.com/gcclinux/scmd/internal/database"
)

// GenerateEmbeddingsForAll generates embeddings for all commands that don't have them.
func GenerateEmbeddingsForAll() error {
	fmt.Println()
	fmt.Println("Generating embeddings for existing commands...")
	fmt.Println("This may take a few minutes depending on the number of commands.")
	fmt.Println()

	if !gemini.IsAvailable() && !ollama.IsAvailable() {
		return fmt.Errorf("no embedding provider available (need Gemini API or Ollama)")
	}

	commands, err := database.GetCommandsWithoutEmbeddings()
	if err != nil {
		return err
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
	if !ollama.IsAvailable() {
		provider = "Gemini"
	}

	for i, cmd := range commands {
		if i%10 == 0 && i > 0 {
			fmt.Printf("Progress: %d/%d (%.1f%%)\n", i, len(commands), float64(i)/float64(len(commands))*100)
		}

		text := cmd.Key + " " + cmd.Data
		var embedding []float64
		var embErr error

		if ollama.IsAvailable() {
			embedding, embErr = ollama.GetEmbedding(text)
			if embErr != nil {
				if gemini.IsAvailable() {
					embedding, embErr = gemini.GetEmbedding(text)
					if embErr == nil {
						provider = "Gemini"
					}
				}
			} else {
				provider = "Ollama"
			}
		} else if gemini.IsAvailable() {
			embedding, embErr = gemini.GetEmbedding(text)
			provider = "Gemini"
		}

		if embErr != nil {
			log.Printf("Failed to generate embedding for ID %d: %v\n", cmd.Id, embErr)
			failCount++
			continue
		}

		if err := database.UpdateEmbedding(cmd.Id, embedding); err != nil {
			log.Printf("Failed to update embedding for ID %d: %v\n", cmd.Id, err)
			failCount++
			continue
		}

		successCount++

		if gemini.IsAvailable() && i < len(commands)-1 {
			time.Sleep(100 * time.Millisecond)
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

// CheckEmbeddingStats shows statistics about embeddings in the database.
func CheckEmbeddingStats() error {
	total, withEmbeddings, err := database.GetEmbeddingStats()
	if err != nil {
		return err
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
