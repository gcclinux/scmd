package search

import (
	"strings"

	"github.com/gcclinux/scmd/internal/database"
)

// CommandScore holds a command record with its match score.
type CommandScore struct {
	Record     database.CommandRecord
	Score      int
	MatchCount int
	TotalWords int
}

// ScoreCommands scores commands based on how many query words they match.
func ScoreCommands(commands []database.CommandRecord, query string) []CommandScore {
	queryWords := ExtractQueryWords(query)
	totalWords := len(queryWords)

	var scored []CommandScore

	for _, cmd := range commands {
		searchText := strings.ToLower(cmd.Key + " " + cmd.Data)

		matchCount := 0
		for _, word := range queryWords {
			paddedText := " " + searchText + " "
			paddedWord := " " + word + " "
			if strings.Contains(paddedText, paddedWord) {
				matchCount++
			}
		}

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

// ExtractQueryWords extracts meaningful words from a query, removing common words.
func ExtractQueryWords(query string) []string {
	query = strings.ToLower(query)

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

	words := strings.Fields(query)

	var filtered []string
	for _, word := range words {
		if len(word) >= 3 {
			filtered = append(filtered, word)
		}
	}

	return filtered
}

// ExtractKeywords is an alias for building a cleaned query string from ExtractQueryWords.
func ExtractKeywords(input string) string {
	words := ExtractQueryWords(input)
	return strings.Join(words, " ")
}

// FilterByMinScore filters commands that meet minimum score threshold.
func FilterByMinScore(scored []CommandScore, minScore int) []CommandScore {
	var filtered []CommandScore
	for _, s := range scored {
		if s.Score >= minScore {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// GetBestMatches returns commands with the highest scores.
func GetBestMatches(scored []CommandScore, limit int) []CommandScore {
	if len(scored) <= limit {
		return scored
	}
	return scored[:limit]
}

// HasGoodMatches checks if we have commands with at least minScore% match.
func HasGoodMatches(scored []CommandScore, minScore int) bool {
	for _, s := range scored {
		if s.Score >= minScore {
			return true
		}
	}
	return false
}
