package main

import (
	"fmt"
	"log"
	"strings"
)

// saveCMD takes the cmd and details (string) and saves it to PostgreSQL database
func saveCmd(cmd, details string) {
	// Clean markdown code blocks if present
	// This allows users to paste content wrapped in ``` ... ``` directly
	details = strings.TrimSpace(details)
	if strings.HasPrefix(details, "```") {
		if idx := strings.Index(details, "\n"); idx != -1 {
			details = details[idx+1:]
		}
	}
	details = strings.TrimSpace(details)
	if strings.HasSuffix(details, "```") {
		details = strings.TrimSuffix(details, "```")
	}
	details = strings.TrimSpace(details)

	// Initialize embedding providers (Gemini first, then Ollama)
	InitGemini()
	InitOllama()

	// Initialize database connection
	if err := InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer CloseDB()

	// Add command to database
	status, err := AddCommand(cmd, details)
	if err != nil {
		fmt.Println("Error saving command:", err)
		fmt.Println("returned: ( false )")
	} else {
		fmt.Println("returned: (", status, ")")
	}
	fmt.Println()
}
