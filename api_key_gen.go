package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// aiAccessGranted is the package-level gate for AI and embedding features.
// It is set once during startup by ValidateAPIAccess().
var aiAccessGranted = false

// CreateAPIKey generates a 32-character random API key, stores it in the
// access table paired with the given email, and prints it to the screen.
// Usage: scmd --create-api "user@example.com"
func CreateAPIKey(email string) {
	email = strings.TrimSpace(email)
	if email == "" {
		log.Fatal("Usage: scmd --create-api \"user@example.com\"")
	}

	// Ensure the database is initialised
	if err := InitDB(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer CloseDB()

	// Generate 16 random bytes → 32 hex characters
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		log.Fatalf("Failed to generate random key: %v", err)
	}
	apiKey := strings.ToUpper(hex.EncodeToString(raw))

	accessTbl := os.Getenv("ACCESS_TB")
	if accessTbl == "" {
		accessTbl = "access"
	}

	// Upsert: if the email already exists update the key, otherwise insert
	query := fmt.Sprintf(`
		INSERT INTO %s (email, api_key)
		VALUES ($1, $2)
		ON CONFLICT (email) DO UPDATE SET api_key = EXCLUDED.api_key`, accessTbl)

	if _, err := db.Exec(query, email, apiKey); err != nil {
		log.Fatalf("Failed to store API key: %v", err)
	}

	fmt.Println()
	fmt.Println("======================================================")
	fmt.Printf("  Email   : %s\n", email)
	fmt.Printf("  API Key : %s\n", apiKey)
	fmt.Println()

	// Automatically write / update API_ACCESS in the .env file
	if err := writeAPIAccessToEnv(apiKey); err != nil {
		fmt.Printf("  ⚠ Could not update .env automatically: %v\n", err)
		fmt.Println("  Add this line manually to your .env file:")
		fmt.Printf("  API_ACCESS=%s\n", apiKey)
	} else {
		fmt.Println("  ✅ API_ACCESS has been written to your .env file.")
	}
	fmt.Println("======================================================")
	fmt.Println()
}

// writeAPIAccessToEnv updates or appends API_ACCESS=<key> in the .env file
// found in the current working directory.
func writeAPIAccessToEnv(apiKey string) error {
	// Locate the .env file
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	envPath := []string{cwd}

	// Also check executable directory as fallback
	if execPath, err2 := os.Executable(); err2 == nil {
		envPath = append(envPath, filepath.Dir(execPath))
	}

	var filePath string
	for _, dir := range envPath {
		p := filepath.Join(dir, ".env")
		if _, statErr := os.Stat(p); statErr == nil {
			filePath = p
			break
		}
	}
	if filePath == "" {
		return fmt.Errorf(".env file not found")
	}

	// Read existing content
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(raw), "\n")
	newLine := "API_ACCESS=" + apiKey
	found := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "API_ACCESS=") {
			lines[i] = newLine
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, newLine)
	}

	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
}

// ValidateAPIAccess checks whether the API_ACCESS key in the .env file
// matches an entry in the access table. Sets aiAccessGranted accordingly.
// Returns true if access is granted or if no access table / key is configured
// (so existing installs without the access system still work).
func ValidateAPIAccess() bool {
	apiKey := strings.TrimSpace(os.Getenv("API_ACCESS"))

	// If API_ACCESS is not set in .env, allow unrestricted access (backward-compatible)
	if apiKey == "" {
		aiAccessGranted = true
		return true
	}

	if db == nil {
		log.Println("⚠ API access validation skipped: database not connected")
		aiAccessGranted = false
		return false
	}

	accessTbl := os.Getenv("ACCESS_TB")
	if accessTbl == "" {
		accessTbl = "access"
	}

	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE api_key = $1", accessTbl)
	err := db.QueryRow(query, apiKey).Scan(&count)
	if err != nil {
		log.Printf("⚠ API access check failed: %v", err)
		aiAccessGranted = false
		return false
	}

	if count > 0 {
		aiAccessGranted = true
		log.Println("✓ API access validated")
		return true
	}

	log.Println("✗ API access denied: key not found in database")
	aiAccessGranted = false
	return false
}

// requireAIAccess returns true if AI features are unlocked.
// Call this at the top of any AI/embedding function.
func requireAIAccess() bool {
	if !aiAccessGranted {
		fmt.Println("⛔ AI features are locked. Add a valid API_ACCESS key to your .env")
		fmt.Println("   Generate one with:  scmd --create-api \"your@email.com\"")
	}
	return aiAccessGranted
}
