package main

import (
	"strings"
)

// CommandScore holds a command record with its match score
type CommandScore struct {
	Record     CommandRecord
	Score      int
	MatchCount int
	TotalWords int
}

// ScoreCommands scores commands based on how many query words they match
func ScoreCommands(commands []CommandRecord, query string) []CommandScore {
	// Extract words from query (lowercase, remove common words)
	queryWords := extractQueryWords(query)
	totalWords := len(queryWords)

	var scored []CommandScore

	for _, cmd := range commands {
		// Combine key and data for searching
		searchText := strings.ToLower(cmd.Key + " " + cmd.Data)

		// Count how many query words are found
		matchCount := 0
		for _, word := range queryWords {
			// Check for whole word match to avoid false positives
			// Add spaces around searchText for boundary checking
			paddedText := " " + searchText + " "
			paddedWord := " " + word + " "
			if strings.Contains(paddedText, paddedWord) {
				matchCount++
			}
		}

		// Calculate score (percentage of words matched)
		score := 0
		if totalWords > 0 {
			score = (matchCount * 100) / totalWords
		}

		scored = append(scored, CommandScore{
			Record:     cmd,
			Score:      score,
			MatchCount: matchCount,
			TotalWords: totalWords,
		})
	}

	// Sort by score (highest first)
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].Score > scored[i].Score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	return scored
}

// extractQueryWords extracts meaningful words from query
func extractQueryWords(query string) []string {
	query = strings.ToLower(query)

	// Remove common question words
	removeWords := []string{
		"show", "me", "give", "provide", "with", "find",
		"how", "to", "do", "i", "a", "the", "is", "are",
		"can", "you", "please", "need", "want", "looking",
		"for", "search", "example", "examples", "command",
		"commands", "what", "where", "when", "why", "which",
	}

	for _, word := range removeWords {
		query = strings.ReplaceAll(query, " "+word+" ", " ")
		query = strings.ReplaceAll(query, word+" ", " ")
		query = strings.ReplaceAll(query, " "+word, " ")
	}

	// Split into words
	words := strings.Fields(query)

	// Filter out very short words (< 3 chars)
	var filtered []string
	for _, word := range words {
		if len(word) >= 3 {
			filtered = append(filtered, word)
		}
	}

	return filtered
}

// FilterByMinScore filters commands that meet minimum score threshold
func FilterByMinScore(scored []CommandScore, minScore int) []CommandScore {
	var filtered []CommandScore
	for _, s := range scored {
		if s.Score >= minScore {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// GetBestMatches returns commands with the highest scores
func GetBestMatches(scored []CommandScore, limit int) []CommandScore {
	if len(scored) <= limit {
		return scored
	}
	return scored[:limit]
}

// HasGoodMatches checks if we have commands with at least minScore% match
func HasGoodMatches(scored []CommandScore, minScore int) bool {
	for _, s := range scored {
		if s.Score >= minScore {
			return true
		}
	}
	return false
}
