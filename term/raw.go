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

func ReadStdin(buf []byte) (int, error) {
	return unix.Read(int(os.Stdin.Fd()), buf)
}
