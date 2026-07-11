//go:build windows

package cli

import (
	"os"

	"golang.org/x/sys/windows"
)

// Windows console-mode flags for virtual-terminal I/O (available since Win10 v1511).
// Defined locally so we compile on older sys package versions too.
const (
	enableVTInput      uint32 = 0x0200 // ENABLE_VIRTUAL_TERMINAL_INPUT  (stdin)
	enableVTProcessing uint32 = 0x0004 // ENABLE_VIRTUAL_TERMINAL_PROCESSING (stdout)
)

// termState holds the original console modes for stdin and stdout.
type termState struct {
	inMode  uint32
	outMode uint32
	inH     windows.Handle
	outH    windows.Handle
}

// enableRawMode switches the Windows console to raw mode:
//   - stdin: disables line-input, echo and signal processing; enables VT input
//             so arrow keys arrive as \x1b[A / \x1b[B / … escape sequences.
//   - stdout: enables ANSI/VT sequence processing so our escape codes render.
//
// Returns an error if stdin/stdout are not console handles (e.g. piped).
func enableRawMode() (*termState, error) {
	inH := windows.Handle(os.Stdin.Fd())
	outH := windows.Handle(os.Stdout.Fd())

	var inMode, outMode uint32
	if err := windows.GetConsoleMode(inH, &inMode); err != nil {
		return nil, err
	}
	if err := windows.GetConsoleMode(outH, &outMode); err != nil {
		return nil, err
	}

	// Build raw stdin mode.
	newIn := inMode &^ (windows.ENABLE_LINE_INPUT |
		windows.ENABLE_ECHO_INPUT |
		windows.ENABLE_PROCESSED_INPUT)
	newIn |= enableVTInput

	// Enable ANSI processing on stdout.
	newOut := outMode | enableVTProcessing

	if err := windows.SetConsoleMode(inH, newIn); err != nil {
		return nil, err
	}
	if err := windows.SetConsoleMode(outH, newOut); err != nil {
		_ = windows.SetConsoleMode(inH, inMode) // rollback stdin change
		return nil, err
	}

	return &termState{
		inMode: inMode, outMode: outMode,
		inH: inH, outH: outH,
	}, nil
}

// restoreTermMode restores the Windows console to the modes saved by enableRawMode.
func restoreTermMode(s *termState) {
	if s == nil {
		return
	}
	_ = windows.SetConsoleMode(s.inH, s.inMode)
	_ = windows.SetConsoleMode(s.outH, s.outMode)
}
