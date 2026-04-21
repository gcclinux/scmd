package cli

import (
	"strings"
	"testing"
)

func TestBuildFeedbackPrompt_ZeroBlocks(t *testing.T) {
	got := buildFeedbackPrompt(0, false)
	if !strings.Contains(got, "[s]") {
		t.Error("expected prompt to contain [s]")
	}
	if !strings.Contains(got, "[n]") {
		t.Error("expected prompt to contain [n]")
	}
	if strings.Contains(got, "[x]") {
		t.Error("expected prompt to NOT contain [x] when codeBlockCount is 0")
	}
	if strings.Contains(got, "Execute") {
		t.Error("expected prompt to NOT contain Execute when codeBlockCount is 0")
	}
}

func TestBuildFeedbackPrompt_OneBlock(t *testing.T) {
	got := buildFeedbackPrompt(1, false)
	if !strings.Contains(got, "[s]") {
		t.Error("expected prompt to contain [s]")
	}
	if !strings.Contains(got, "[n]") {
		t.Error("expected prompt to contain [n]")
	}
	if !strings.Contains(got, "[x] Execute") {
		t.Error("expected prompt to contain [x] Execute")
	}
}

func TestBuildFeedbackPrompt_FiveBlocks(t *testing.T) {
	got := buildFeedbackPrompt(5, false)
	if !strings.Contains(got, "[s]") {
		t.Error("expected prompt to contain [s]")
	}
	if !strings.Contains(got, "[n]") {
		t.Error("expected prompt to contain [n]")
	}
	if !strings.Contains(got, "Execute:") {
		t.Errorf("expected prompt to contain 'Execute:', got %q", got)
	}
	for _, n := range []string{"[1]", "[2]", "[3]", "[4]", "[5]"} {
		if !strings.Contains(got, n) {
			t.Errorf("expected prompt to contain %q, got %q", n, got)
		}
	}
}

func TestBuildFeedbackPrompt_SavedResponse_OmitsSaveKeepsRegenerate(t *testing.T) {
	got := buildFeedbackPrompt(3, true)
	if strings.Contains(got, "[s]") {
		t.Error("saved response prompt should NOT contain [s]")
	}
	if !strings.Contains(got, "[n]") {
		t.Error("saved response prompt should still contain [n]")
	}
	if !strings.Contains(got, "Execute:") {
		t.Errorf("expected prompt to contain 'Execute:', got %q", got)
	}
	for _, n := range []string{"[1]", "[2]", "[3]"} {
		if !strings.Contains(got, n) {
			t.Errorf("expected prompt to contain %q, got %q", n, got)
		}
	}
}

func TestBuildFeedbackPrompt_SavedResponse_OneBlock(t *testing.T) {
	got := buildFeedbackPrompt(1, true)
	if strings.Contains(got, "[s]") {
		t.Error("saved response prompt should NOT contain [s]")
	}
	if !strings.Contains(got, "[n]") {
		t.Error("saved response prompt should still contain [n]")
	}
	if !strings.Contains(got, "[x] Execute") {
		t.Errorf("expected prompt to contain '[x] Execute', got %q", got)
	}
}

func TestBuildFeedbackPrompt_SavedResponse_ZeroBlocks(t *testing.T) {
	got := buildFeedbackPrompt(0, true)
	if strings.Contains(got, "[s]") {
		t.Error("saved response prompt should NOT contain [s]")
	}
	if !strings.Contains(got, "[n]") {
		t.Error("saved response prompt should still contain [n]")
	}
}

func TestIsFeedbackInput_RejectsNumericWithNoBlocks(t *testing.T) {
	if isFeedbackInput("0", 0) {
		t.Error("expected isFeedbackInput(\"0\", 0) to return false")
	}
	if isFeedbackInput("1", 0) {
		t.Error("expected isFeedbackInput(\"1\", 0) to return false")
	}
}

func TestIsFeedbackInput_AcceptsNumericWithBlocks(t *testing.T) {
	if !isFeedbackInput("1", 3) {
		t.Error("expected isFeedbackInput(\"1\", 3) to return true")
	}
	if !isFeedbackInput("2", 3) {
		t.Error("expected isFeedbackInput(\"2\", 3) to return true")
	}
}

func TestIsFeedbackInput_AcceptsValidInputs(t *testing.T) {
	validInputs := []string{"s", "n", "x", "x 1", "x 3"}
	for _, input := range validInputs {
		if !isFeedbackInput(input, 0) {
			t.Errorf("expected isFeedbackInput(%q, 0) to return true", input)
		}
	}
}
