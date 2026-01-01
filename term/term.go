package term

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
)

type Size struct {
    Cols int
    Rows int
}

func TermSize() (size Size, err error) {
    ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
    if err != nil {
        return Size{}, err
    }
    return Size{Cols: int(ws.Col), Rows: int(ws.Row)}, nil
}

func WatchResize(onResize func()) {
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGWINCH)

    go func() {
        for range ch {
            onResize()
        }
    }()
}
