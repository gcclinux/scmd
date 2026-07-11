//go:build linux

package cli

import "golang.org/x/sys/unix"

// ioctlGetTermios reads the current terminal attributes for fd.
func ioctlGetTermios(fd int, t *unix.Termios) error {
	term, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return err
	}
	*t = *term
	return nil
}

// ioctlSetTermios writes terminal attributes for fd (TCSANOW = immediate).
func ioctlSetTermios(fd int, t *unix.Termios) error {
	return unix.IoctlSetTermios(fd, unix.TCSETS, t) // TCSETS = set now
}
