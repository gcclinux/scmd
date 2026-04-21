package cli

import (
	"fmt"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// Feature: execute-command-option, Property 2: Feedback prompt correctness for varying block counts
// **Validates: Requirements 3.2, 3.3**
func TestProperty_FeedbackPrompt(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random non-negative integer for codeBlockCount.
		codeBlockCount := rapid.IntRange(0, 100).Draw(rt, "codeBlockCount")

		prompt := buildFeedbackPrompt(codeBlockCount, false)

		// Property: single-line (no newline characters).
		if strings.Contains(prompt, "\n") {
			rt.Fatalf("prompt contains newline for codeBlockCount=%d: %q", codeBlockCount, prompt)
		}

		// Property: always contains [s] and [n].
		if !strings.Contains(prompt, "[s]") {
			rt.Fatalf("prompt missing [s] for codeBlockCount=%d: %q", codeBlockCount, prompt)
		}
		if !strings.Contains(prompt, "[n]") {
			rt.Fatalf("prompt missing [n] for codeBlockCount=%d: %q", codeBlockCount, prompt)
		}

		// Property: contains execute option iff codeBlockCount >= 1.
		hasExecute := strings.Contains(prompt, "Execute")
		if codeBlockCount >= 1 && !hasExecute {
			rt.Fatalf("prompt should contain Execute for codeBlockCount=%d: %q", codeBlockCount, prompt)
		}
		if codeBlockCount == 0 && hasExecute {
			rt.Fatalf("prompt should NOT contain Execute for codeBlockCount=0: %q", prompt)
		}

		// Property: for 1 block, shows [x] Execute (no numbered options).
		if codeBlockCount == 1 {
			if !strings.Contains(prompt, "[x] Execute") {
				rt.Fatalf("prompt should contain '[x] Execute' for codeBlockCount=1: %q", prompt)
			}
		}

		// Property: for >1 blocks, shows each numbered option [1] through [N].
		if codeBlockCount > 1 {
			for i := 1; i <= codeBlockCount; i++ {
				tag := fmt.Sprintf("[%d]", i)
				if !strings.Contains(prompt, tag) {
					rt.Fatalf("prompt should contain %s for codeBlockCount=%d: %q", tag, codeBlockCount, prompt)
				}
			}
		}
	})
}

// Feature: execute-command-option, Property 3: Invalid index rejection with valid range
// **Validates: Requirements 5.2**
func TestProperty_InvalidIndexRejection(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate a random list length L >= 1.
		totalBlocks := rapid.IntRange(1, 500).Draw(rt, "totalBlocks")

		// Generate a random invalid index: either < 1 or > totalBlocks.
		// We pick from two ranges and combine them.
		invalidIndex := rapid.OneOf(
			rapid.IntRange(-100, 0),
			rapid.IntRange(totalBlocks+1, totalBlocks+200),
		).Draw(rt, "invalidIndex")

		errMsg := validateExecuteIndex(invalidIndex, totalBlocks)

		// Property: error message must not be empty for invalid indices.
		if errMsg == "" {
			rt.Fatalf("expected error for index=%d, total=%d, got empty string", invalidIndex, totalBlocks)
		}

		// Property: error message must contain "1" (the lower bound).
		if !strings.Contains(errMsg, "1") {
			rt.Fatalf("error message should contain '1' for total=%d: %q", totalBlocks, errMsg)
		}

		// Property: error message must contain the upper bound (totalBlocks).
		upperStr := fmt.Sprintf("%d", totalBlocks)
		if !strings.Contains(errMsg, upperStr) {
			rt.Fatalf("error message should contain %q for total=%d: %q", upperStr, totalBlocks, errMsg)
		}
	})
}

// Feature: execute-command-option, Property 4: Non-numeric execute argument rejection
// **Validates: Requirements 5.3**
func TestProperty_NonNumericExecuteArgRejection(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Generate random strings that are NOT parseable as integers.
		// We filter out strings that strconv.Atoi would accept.
		arg := rapid.StringMatching(`[a-zA-Z!@#$%^&*()\-_=+\[\]{}<>?,./;:'"\\|~ ]{1,50}`).Draw(rt, "arg")

		_, usageMsg := parseExecuteArg(arg)

		// Property: non-numeric argument must produce a usage message.
		if usageMsg == "" {
			rt.Fatalf("expected usage message for non-numeric arg %q, got empty string", arg)
		}

		// Property: usage message should mention the correct syntax.
		if !strings.Contains(usageMsg, "Usage") {
			rt.Fatalf("usage message should contain 'Usage' for arg %q: %q", arg, usageMsg)
		}
	})
}
