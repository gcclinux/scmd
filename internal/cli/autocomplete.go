package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	isatty "github.com/mattn/go-isatty"
)

// slashEntry is a single autocomplete candidate for the '/' menu.
type slashEntry struct {
	name string
	desc string
}

// allSlashEntries is the master list used for autocomplete.
// Keep in sync with help.go (printInteractiveHelp) and
// commands.go (handleSlashCommand).
var allSlashEntries = []slashEntry{
	{"/setup", "Edit & create current config.json"},
	{"/add", "Add a new command (use | separator)"},
	{"/search", "Search for commands matching pattern"},
	{"/delete", "Delete a command by ID"},
	{"/show", "Show command and description by ID"},
	{"/list", "List recent commands"},
	{"/help", "Show this help message"},
	{"/ai", "Show AI/Ollama status"},
	{"/config", "Show current config.json settings"},
	{"/embeddings", "Check embedding statistics"},
	{"/generate", "Generate embeddings for all commands"},
	{"/clear", "Clear the screen"},
	{"/exit", "Exit interactive mode"},
	{"/quit", "Exit interactive mode"},
	{"/ubuntu", "Ubuntu expert persona"},
	{"/debian", "Debian expert persona"},
	{"/fedora", "Fedora expert persona"},
	{"/archlinux", "Arch Linux master persona"},
	{"/windows", "Windows admin persona"},
	{"/powershell", "PowerShell guru persona"},
	{"/research", "Analyse, recommend & document"},
	{"/code", "Read, Think, execute Agentic RAG"},
	{"/run", "Execute a system command"},
}

const maxMenuItems = 8

// filterSlashEntries returns up to maxMenuItems entries whose .name begins
// with prefix (case-insensitive).
func filterSlashEntries(prefix string) []slashEntry {
	low := strings.ToLower(prefix)
	var out []slashEntry
	for _, e := range allSlashEntries {
		if strings.HasPrefix(e.name, low) {
			out = append(out, e)
			if len(out) >= maxMenuItems {
				break
			}
		}
	}
	return out
}

// readLineWithCompletion reads one complete input line from the terminal,
// showing an interactive slash-command dropdown/uplift menu when the FIRST
// character typed is '/'.  The menu is filtered as the user types and can be
// navigated with the ↑ / ↓ arrow keys.  Tab accepts the highlighted entry
// without submitting; Enter accepts and submits.
//
// The function prints the "scmd> " prompt itself — callers must NOT print it.
//
// Falls back to a plain bufio read when stdin is not a TTY or raw mode
// cannot be enabled (e.g. piped, redirected, or older terminal).
func readLineWithCompletion(fallback *bufio.Reader) (string, error) {
	const promptStr = "scmd> "

	// ── Non-interactive stdin: plain read ────────────────────────────────────
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		fmt.Print(promptStr)
		line, err := fallback.ReadString('\n')
		return strings.TrimRight(line, "\r\n"), err
	}

	state, err := enableRawMode()
	if err != nil {
		// Raw mode failed — degrade gracefully.
		fmt.Print(promptStr)
		line, err2 := fallback.ReadString('\n')
		return strings.TrimRight(line, "\r\n"), err2
	}
	defer restoreTermMode(state)

	var (
		buf           []rune       // accumulated input characters
		matches       []slashEntry // current autocomplete candidates
		sel           = -1         // highlighted menu index (-1 = nothing selected)
		prevMenuLines = 0           // menu lines drawn above the prompt in last redraw
	)

	// redraw clears the previously drawn block (menu items above + prompt line)
	// and redraws it with the new state.  After redraw the cursor sits at the
	// end of the prompt line.
	//
	// Layout (menu appears ABOVE the prompt, uplift style):
	//
	//   /setup    - Edit & create config       ← menu item 0
	//   /search   - Search for commands  ◀▶    ← menu item 1 (highlighted)
	//   /show     - Show command by ID         ← menu item 2
	//   scmd> /s_                              ← prompt  (cursor here)
	//
	// Because OPOST is disabled on Unix we use explicit \r\n for newlines
	// so the rendering is identical on Unix and Windows.
	redraw := func(items []slashEntry, selected int, input []rune) {
		// Move cursor up past the previous menu block.
		if prevMenuLines > 0 {
			fmt.Printf("\x1b[%dA", prevMenuLines)
		}
		// Go to column 1 and erase from here to end of screen.
		fmt.Print("\r\x1b[J")

		// Render menu items (above the prompt line).
		for i, item := range items {
			if i == selected {
				fmt.Print("\x1b[7m") // reverse-video highlight
			}
			fmt.Printf("  %-24s %s\x1b[0m\x1b[K\r\n", item.name, item.desc)
		}
		prevMenuLines = len(items)

		// Render the prompt line; cursor ends here.
		fmt.Print(promptStr + string(input))
	}

	// Show the initial prompt (empty buffer, no menu).
	redraw(nil, -1, buf)

	inBuf := make([]byte, 4) // big enough for a 3-byte CSI escape sequence

	for {
		// Read the first byte of the next keystroke.
		n, readErr := os.Stdin.Read(inBuf[:1])
		if n == 0 || readErr != nil {
			// EOF or read error: submit whatever is in the buffer.
			if prevMenuLines > 0 {
				fmt.Printf("\x1b[%dA", prevMenuLines)
			}
			fmt.Printf("\r\x1b[J%s%s\r\n", promptStr, string(buf))
			if readErr == io.EOF {
				return strings.TrimSpace(string(buf)), nil
			}
			return string(buf), readErr
		}
		b := inBuf[0]

		switch {

		// ── Ctrl+C: exit ─────────────────────────────────────────────────────
		case b == 3:
			fmt.Print("\r\n")
			restoreTermMode(state) // defer won't run after os.Exit
			os.Exit(0)

		// ── Ctrl+D: EOF on empty buffer ──────────────────────────────────────
		case b == 4:
			if len(buf) == 0 {
				fmt.Print("\r\n")
				return "", io.EOF
			}

		// ── Tab: accept highlighted selection, DON'T submit ──────────────────
		case b == 9:
			if sel >= 0 && sel < len(matches) {
				buf = []rune(matches[sel].name + " ")
				matches = nil
				sel = -1
				redraw(nil, -1, buf)
			}

		// ── Enter: accept highlighted selection (if any) and submit ──────────
		case b == 13 || b == 10:
			if sel >= 0 && sel < len(matches) {
				buf = []rune(matches[sel].name + " ")
			}
			// Clear block and echo the final line.
			if prevMenuLines > 0 {
				fmt.Printf("\x1b[%dA", prevMenuLines)
			}
			fmt.Printf("\r\x1b[J%s%s\r\n", promptStr, string(buf))
			return strings.TrimSpace(string(buf)), nil

		// ── ESC: arrow key sequence or bare Escape ───────────────────────────
		case b == 27:
			n2, _ := os.Stdin.Read(inBuf[:1])
			if n2 == 0 || inBuf[0] != '[' {
				// Bare Escape: dismiss menu, keep buffer as-is.
				matches = nil
				sel = -1
				redraw(nil, -1, buf)
				continue
			}
			// CSI sequence — read the final byte.
			n3, _ := os.Stdin.Read(inBuf[:1])
			if n3 == 0 {
				continue
			}
			switch inBuf[0] {
			case 'A': // ↑ Up arrow: move selection up (wraps)
				if len(matches) > 0 {
					if sel <= 0 {
						sel = len(matches) - 1
					} else {
						sel--
					}
					redraw(matches, sel, buf)
				}
			case 'B': // ↓ Down arrow: move selection down (wraps)
				if len(matches) > 0 {
					if sel < 0 || sel >= len(matches)-1 {
						sel = 0
					} else {
						sel++
					}
					redraw(matches, sel, buf)
				}
			// C/D (right/left) and other sequences: silently ignored.
			}

		// ── Backspace / Delete ────────────────────────────────────────────────
		case b == 127 || b == 8:
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
			}
			if len(buf) > 0 && buf[0] == '/' {
				matches = filterSlashEntries(string(buf))
				if sel >= len(matches) {
					sel = len(matches) - 1
				}
				if len(matches) == 0 {
					sel = -1
				}
			} else {
				matches = nil
				sel = -1
			}
			redraw(matches, sel, buf)

		// ── Multi-byte UTF-8 lead byte ────────────────────────────────────────
		case b >= 0xC0:
			extra := 1
			if b >= 0xF0 {
				extra = 3
			} else if b >= 0xE0 {
				extra = 2
			}
			inBuf[0] = b
			nn, _ := os.Stdin.Read(inBuf[1 : 1+extra])
			if nn == extra {
				r, size := utf8.DecodeRune(inBuf[:1+extra])
				if r != utf8.RuneError || size > 1 {
					buf = append(buf, r)
					// UTF-8 can't start slash autocomplete (commands are ASCII).
					redraw(matches, sel, buf)
				}
			}

		// ── Printable ASCII ───────────────────────────────────────────────────
		case b >= 32:
			r := rune(b)
			switch {
			case len(buf) == 0 && r == '/':
				// '/' is the FIRST character → activate autocomplete.
				buf = append(buf, r)
				matches = filterSlashEntries("/")
				sel = -1
				redraw(matches, sel, buf)

			case len(buf) > 0 && buf[0] == '/':
				// Still in slash mode → narrow the filter.
				buf = append(buf, r)
				matches = filterSlashEntries(string(buf))
				sel = -1
				redraw(matches, sel, buf)

			default:
				// '/' was not the first character → plain text, no autocomplete.
				buf = append(buf, r)
				matches = nil
				sel = -1
				redraw(nil, -1, buf)
			}
		}
	}
}
