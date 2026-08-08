package term

import (
	"os"
	"os/signal"
	"syscall"
	"time"

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

// WatchResize returns a channel that receives a value each time the
// terminal has been resized (debounced 50ms so a drag-resize doesn't
// flood the channel). It does NOT call anything itself -- the caller
// (Renderer.Run) is responsible for reading this channel on the same
// goroutine that owns the Canvas/Buffer, so a resize is never handled
// concurrently with a normal render tick.
func WatchResize() <-chan struct{} {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)

	out := make(chan struct{}, 1)

	go func() {
		var timer *time.Timer
		for range sig {
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(50*time.Millisecond, func() {
				select {
				case out <- struct{}{}:
				default:
				}
			})
		}
	}()

	return out
}