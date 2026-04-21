package cli

import (
	"testing"
)

func TestExtractCodeBlocks_PlainText(t *testing.T) {
	text := "This is plain text with no code blocks at all.\nJust regular prose."
	got := ExtractCodeBlocks(text)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestExtractCodeBlocks_OneBlockWithLanguageTag(t *testing.T) {
	text := "Here is a command:\n```bash\necho hello\n```\nDone."
	got := ExtractCodeBlocks(text)
	if len(got) != 1 {
		t.Fatalf("expected 1 block, got %d", len(got))
	}
	if got[0] != "echo hello" {
		t.Errorf("expected %q, got %q", "echo hello", got[0])
	}
}

func TestExtractCodeBlocks_OneBlockNoLanguageTag(t *testing.T) {
	text := "Try this:\n```\nls -la\n```\n"
	got := ExtractCodeBlocks(text)
	if len(got) != 1 {
		t.Fatalf("expected 1 block, got %d", len(got))
	}
	if got[0] != "ls -la" {
		t.Errorf("expected %q, got %q", "ls -la", got[0])
	}
}

func TestExtractCodeBlocks_MultipleBlocksMixedTags(t *testing.T) {
	text := "First:\n```bash\necho one\n```\nSecond:\n```python\nprint('two')\n```\nThird:\n```\nthree\n```\n"
	got := ExtractCodeBlocks(text)
	if len(got) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(got))
	}
	expected := []string{"echo one", "print('two')", "three"}
	for i, want := range expected {
		if got[i] != want {
			t.Errorf("block %d: expected %q, got %q", i, want, got[i])
		}
	}
}

func TestExtractCodeBlocks_MultiLineContent(t *testing.T) {
	text := "Script:\n```bash\nline1\nline2\nline3\n```\n"
	got := ExtractCodeBlocks(text)
	if len(got) != 1 {
		t.Fatalf("expected 1 block, got %d", len(got))
	}
	want := "line1\nline2\nline3"
	if got[0] != want {
		t.Errorf("expected %q, got %q", want, got[0])
	}
}

func TestExtractCodeBlocks_UnclosedFence(t *testing.T) {
	text := "Broken:\n```bash\necho hello\nno closing fence"
	got := ExtractCodeBlocks(text)
	if got != nil {
		t.Errorf("expected nil for unclosed fence, got %v", got)
	}
}

func TestExtractCodeBlocks_EmptyBlock(t *testing.T) {
	text := "Empty:\n```\n```\nDone."
	got := ExtractCodeBlocks(text)
	if len(got) != 1 {
		t.Fatalf("expected 1 block, got %d", len(got))
	}
	if got[0] != "" {
		t.Errorf("expected empty string, got %q", got[0])
	}
}
