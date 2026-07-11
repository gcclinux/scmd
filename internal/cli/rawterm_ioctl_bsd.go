//go:build darwin || freebsd || netbsd || openbsd || dragonfly

package cli

import "golang.org/x/sys/unix"

// ioctlGetTermios reads the current terminal attributes for fd.
func ioctlGetTermios(fd int, t *unix.Termios) error {
	term, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		return err
	}
	*t = *term
	return nil
}

// ioctlSetTermios writes terminal attributes for fd (TIOCSETA = set now).
func ioctlSetTermios(fd int, t *unix.Termios) error {
	return unix.IoctlSetTermios(fd, unix.TIOCSETA, t)
}
