package render

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/term"
)

type Renderer struct {
	out *bufio.Writer
	RenderMode
	isDirty bool

	ctx    context.Context
	cancel context.CancelFunc
}

func NewRenderer(mode LoopMode, fps int) *Renderer {
	ctx, cancel := context.WithCancel(context.Background())

	var renderMode RenderMode

	switch mode {
	case FixedFPS:
		renderMode = FixedFPSMode(fps)
	case OnDemand:
		renderMode = OnDemandMode()
	}

	r := &Renderer{
		out:        bufio.NewWriter(os.Stdout),
		RenderMode: renderMode,
		ctx:        ctx,
		cancel:     cancel,
	}

	r.Init()

	return r
}

func (r *Renderer) Init() {
    fmt.Fprint(r.out, "\x1b[?25l")
    fmt.Fprint(r.out, "\x1b[?1049h")
    fmt.Fprint(r.out, "\x1b[2J")
    fmt.Fprint(r.out, "\x1b[H")
    r.out.Flush()
}

func (r *Renderer) Run(c *canvas.Canvas) {
	// Wire self-refresh: any node in the tree calling Invalidate() (e.g.
	// via Text.SetValue from a background goroutine) now reaches
	// RequestRedraw. In FixedFPS mode RequestRedraw is already a no-op --
	// every tick redraws unconditionally, so a changed value just shows
	// up on the next tick with no extra plumbing needed.
	c.SetInvalidator(r.RequestRedraw)

	var ticker *time.Ticker

	if r.RenderMode.Mode == FixedFPS {
		ticker = time.NewTicker(time.Second / time.Duration(r.Fps))
		defer ticker.Stop()
	} else {
		r.Redraw <- struct{}{}
	}

	applySize := func() {
		size, err := term.TermSize()
		if err != nil {
			// panic("something went wrong resizing")
		}
		c.ApplySize(size.Cols, size.Rows)
	}

	applySize()

	term.WatchResize(func() {
		applySize()
		r.render(c)
	})

	for {
		select {
		case <-r.ctx.Done():
			r.restore()
			return

		case <-r.Redraw:
			if r.Mode == OnDemand {
				r.render(c)
			}

		case <-func() <-chan time.Time {
			if ticker != nil {
				return ticker.C
			}
			return nil
		}():
			if r.Mode == FixedFPS {
				r.render(c)
			}
		}
	}
}

func (r *Renderer) RequestRedraw() {
	if r.RenderMode.Mode != OnDemand {
		return
	}

	select {
	case r.Redraw <- struct{}{}:
	default:
	}
}

func (r *Renderer) render(c *canvas.Canvas) {
	fmt.Fprint(r.out, "\x1b[H")

	c.Compose()

	r.Flush(c.Buf)
	r.out.Flush()
}

func (r *Renderer) restore() {
    fmt.Fprint(r.out,
        "\x1b[?1000l"+
            "\x1b[?1002l"+
            "\x1b[?1003l"+
            "\x1b[?1005l"+
            "\x1b[?1006l"+
            "\x1b[?1015l"+
            "\x1b[?25h"+
            "\x1b[?1049l",
    )
    r.out.Flush()
}

func (r *Renderer) Stop() {
	r.cancel()
	r.restore()
}

func (r *Renderer) Flush(buf *core.Buffer) {
	cells, width, height := buf.GetCells()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := cells[y][x]
			fmt.Fprintf(r.out, "\x1b[%d;%dH%s", y+1, x+1, term.CellToANSI(*cell))
		}
	}
}