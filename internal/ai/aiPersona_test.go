package ai

import (
	"strings"
	"testing"
)

// TestResearchPersonaRegistered verifies that the "research" key exists in
// GetPersonas() with the correct Name, Description, and a non-empty
// SystemPrompt.
// Validates: Requirements 1.2, 1.3
func TestResearchPersonaRegistered(t *testing.T) {
	personas := GetPersonas()

	persona, ok := personas["research"]
	if !ok {
		t.Fatal(`GetPersonas() does not contain a "research" key`)
	}

	if persona.Name != "Research Analyst" {
		t.Errorf("Name = %q, want %q", persona.Name, "Research Analyst")
	}

	wantDesc := "Analyse, research, and provide markdown recommendations"
	if persona.Description != wantDesc {
		t.Errorf("Description = %q, want %q", persona.Description, wantDesc)
	}

	if persona.SystemPrompt == "" {
		t.Error("SystemPrompt is empty, expected a non-empty prompt")
	}
}

// TestCodePersonaRegistered verifies that the "code" key exists in
// GetPersonas() with the correct Name, Description, and a non-empty
// SystemPrompt containing the required section keywords.
// Validates: Requirements 1.1, 1.2, 1.5, 1.6
func TestCodePersonaRegistered(t *testing.T) {
	personas := GetPersonas()

	persona, ok := personas["code"]
	if !ok {
		t.Fatal(`GetPersonas() does not contain a "code" key`)
	}

	if persona.Name != "Agentic RAG Developer" {
		t.Errorf("Name = %q, want %q", persona.Name, "Agentic RAG Developer")
	}

	wantDesc := "Read, Think, execute Agentic RAG"
	if persona.Description != wantDesc {
		t.Errorf("Description = %q, want %q", persona.Description, wantDesc)
	}

	if persona.SystemPrompt == "" {
		t.Error("SystemPrompt is empty, expected a non-empty prompt")
	}

	requiredKeywords := []string{"Analysis", "Proposed Code", "Summary"}
	for _, kw := range requiredKeywords {
		if !strings.Contains(persona.SystemPrompt, kw) {
			t.Errorf("SystemPrompt does not contain required keyword %q", kw)
		}
	}
}

// TestResearchSystemPromptContent verifies that the research persona's
// system prompt contains the required section names: Analysis,
// Recommendation, Summary, Proposed Fix, and Alternatives.
// Validates: Requirements 2.1, 2.2, 2.3, 2.4, 10.1
func TestResearchSystemPromptContent(t *testing.T) {
	personas := GetPersonas()

	persona, ok := personas["research"]
	if !ok {
		t.Fatal(`GetPersonas() does not contain a "research" key`)
	}

	requiredSections := []string{
		"Analysis",
		"Recommendation",
		"Summary",
		"Proposed Fix",
		"Alternatives",
	}

	for _, section := range requiredSections {
		if !strings.Contains(persona.SystemPrompt, section) {
			t.Errorf("SystemPrompt does not contain required section %q", section)
		}
	}
}
