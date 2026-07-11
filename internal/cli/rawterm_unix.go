//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly

package cli

import (
	"os"

	"golang.org/x/sys/unix"
)

// termState holds the original terminal settings to be restored on exit.
type termState struct {
	termios unix.Termios
	fd      int
}

// enableRawMode switches stdin to raw mode:
//   - no line buffering (each keystroke is immediately available)
//   - no echo
//   - no signal generation (Ctrl+C sends 0x03 instead of SIGINT)
//   - no output post-processing (caller must use explicit \r\n)
//
// Returns a *termState that must be passed to restoreTermMode when done.
// Returns an error if stdin is not a terminal.
func enableRawMode() (*termState, error) {
	fd := int(os.Stdin.Fd())

	var orig unix.Termios
	if err := ioctlGetTermios(fd, &orig); err != nil {
		return nil, err
	}

	raw := orig

	// cfmakeraw equivalent — clear flags that prevent raw behaviour.
	raw.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP |
		unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	raw.Oflag &^= unix.OPOST // disable output post-processing; use \r\n explicitly
	raw.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	raw.Cflag &^= unix.CSIZE | unix.PARENB
	raw.Cflag |= unix.CS8
	raw.Cc[unix.VMIN] = 1  // return after each byte
	raw.Cc[unix.VTIME] = 0 // no read timeout

	if err := ioctlSetTermios(fd, &raw); err != nil {
		return nil, err
	}
	return &termState{termios: orig, fd: fd}, nil
}

// restoreTermMode restores the terminal to the state saved by enableRawMode.
func restoreTermMode(s *termState) {
	if s == nil {
		return
	}
	_ = ioctlSetTermios(s.fd, &s.termios)
}
