package term

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

const (
	ioctlGetTermios = unix.TCGETS
	ioctlSetTermios = unix.TCSETS
)

// EnableRawMode puts stdin into raw mode:
//   - no line buffering (ICANON off) — bytes are available to read as
//     soon as they're typed, not just after Enter
//   - no local echo (ECHO off) — typed characters don't get printed back
//   - no signal-generating keys (ISIG off) — Ctrl+C/Ctrl+Z arrive as
//     plain bytes (0x03 / 0x1a) instead of killing the process via SIGINT
//   - VMIN=0, VTIME=1 — a read call returns after ~100ms even if nothing
//     was typed (0 bytes, no error), instead of blocking indefinitely.
//     This is what lets input decoding tell a lone ESC apart from the
//     start of an escape sequence like ESC [ A without hanging forever.
//
// Call the returned restore func (typically via defer) to put the
// terminal back into its original mode.
func EnableRawMode() (restore func() error, err error) {
	fd := int(os.Stdin.Fd())

	orig, err := unix.IoctlGetTermios(fd, ioctlGetTermios)
	if err != nil {
		return nil, err
	}

	raw := *orig

	raw.Iflag &^= unix.BRKINT | unix.ICRNL | unix.INPCK | unix.ISTRIP | unix.IXON
	raw.Oflag &^= unix.OPOST
	raw.Cflag |= unix.CS8
	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG | unix.IEXTEN

	raw.Cc[unix.VMIN] = 0
	raw.Cc[unix.VTIME] = 1

	if err := unix.IoctlSetTermios(fd, ioctlSetTermios, &raw); err != nil {
		return nil, err
	}

	var once sync.Once
	restore = func() error {
		var restoreErr error
		once.Do(func() {
			restoreErr = unix.IoctlSetTermios(fd, ioctlSetTermios, orig)
		})
		return restoreErr
	}

	return restore, nil
}

// SafeRawMode wraps EnableRawMode with a guarantee that the terminal
// gets restored even if the process is killed by SIGTERM/SIGHUP, or if
// the caller panics and recovers upstream (call restore() in a defer
// right after this returns — that covers the panic case; the signal
// goroutine covers external termination).
//
// Note SIGINT (Ctrl+C) isn't in this list: with ISIG off, Ctrl+C no
// longer generates a signal at all, it just arrives as byte 0x03 for
// your input reader to handle like any other key.
func SafeRawMode() (restore func(), err error) {
	rawRestore, err := EnableRawMode()
	if err != nil {
		return nil, err
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGHUP)

	done := make(chan struct{})
	go func() {
		select {
		case <-sigCh:
			rawRestore()
			os.Exit(1)
		case <-done:
		}
	}()

	var once sync.Once
	restore = func() {
		once.Do(func() {
			close(done)
			signal.Stop(sigCh)
			rawRestore()
		})
	}

	return restore, nil
}

// ReadStdin performs a single low-level read on fd 0, going straight
// through the read(2) syscall via unix.Read rather than os.Stdin.Read.
//
// Why: Go's os.File wraps pollable fds (ttys included) with the
// runtime's internal epoll-based poller. That poller can interfere
// with the VMIN/VTIME timeout configured in EnableRawMode, making the
// ~100ms timeout unreliable when read through os.File. Calling
// unix.Read directly keeps the fd in plain blocking mode, so VTIME
// behaves exactly as configured: returns (0, nil) on timeout, or
// (n, nil) with n>0 as soon as bytes arrive.
func ReadStdin(buf []byte) (int, error) {
	return unix.Read(int(os.Stdin.Fd()), buf)
}