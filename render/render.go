package render

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/term"
)

type Renderer struct {
	out *bufio.Writer
	RenderMode
	isDirty      bool
	logger       framework.Logger
	mouseEnabled bool

	lastFrame    []core.Cell
	lastW, lastH int

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

func NewRenderer(mode RenderMode, mouseEnabled bool, logger framework.Logger) (*Renderer, error) {
	if !mode.valid() {
		return nil, errors.New("render: invalid RenderMode; build one with render.FixedFPSMode(fps) or render.OnDemandMode(), don't leave AppConfig.RenderMode unset")
	}

	ctx, cancel := context.WithCancel(context.Background())

	r := &Renderer{
		out:          bufio.NewWriter(os.Stdout),
		RenderMode:   mode,
		mouseEnabled: mouseEnabled,
		ctx:          ctx,
		cancel:       cancel,
		logger:       logger,
		done:         make(chan struct{}),
	}

	return r, nil
}

func (r *Renderer) Init() {
	fmt.Fprint(r.out, "\x1b[?25l")
	fmt.Fprint(r.out, "\x1b[?1049h")
	fmt.Fprint(r.out, "\x1b[2J")
	fmt.Fprint(r.out, "\x1b[H")
	if r.mouseEnabled {
		fmt.Fprint(r.out, "\x1b[?1000h\x1b[?1002h\x1b[?1006h")
	}
	r.out.Flush()
}

func (r *Renderer) Start(c *canvas.Canvas) {
	r.Init()
	go r.Run(c)
}

func (r *Renderer) Run(c *canvas.Canvas) {
	defer close(r.done)

	c.SetParentStyle(&framework.Style{Bg: core.Transparent, Fg: core.Transparent})

	var ticker *time.Ticker
	if r.mode == FixedFPS {
		ticker = time.NewTicker(time.Second / time.Duration(r.fps))
		defer ticker.Stop()
	} else {
		r.redraw <- struct{}{}
	}

	applySize := func() {
		size, err := term.TermSize()
		if err != nil {
			return
		}
		c.ApplySize(size.Cols, size.Rows)
	}

	applySize()

	resizeCh := term.WatchResize()

	r.logger.Info("Renderer Started")

	for {
		select {
		case <-r.ctx.Done():
			r.restore()
			return

		case <-resizeCh:
			applySize()
			r.render(c)

		case <-r.redraw:
			if r.mode == OnDemand {
				r.render(c)
			}

		case <-func() <-chan time.Time {
			if ticker != nil {
				return ticker.C
			}
			return nil
		}():
			if r.mode == FixedFPS {
				r.render(c)
			}
		}
	}
}

func (r *Renderer) RequestRedraw() {
	if r.mode != OnDemand {
		return
	}

	r.logger.Debug("On Demand render cycle triggered")

	select {
	case r.redraw <- struct{}{}:
	default:
	}
}

func (r *Renderer) render(c *canvas.Canvas) {
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
	r.logger.Info("Renderer Stopped")
	r.cancel()
	<-r.done
}

func (r *Renderer) Flush(buf *core.Buffer) {
	cells := buf.Cells()
	w, h := buf.W, buf.H

	fullRepaint := r.lastFrame == nil || r.lastW != w || r.lastH != h
	if fullRepaint {
		r.lastFrame = make([]core.Cell, len(cells))
		r.lastW, r.lastH = w, h
	}

	for y := 0; y < h; y++ {
		rowStart := y * w
		x := 0
		for x < w {
			idx := rowStart + x
			cell := cells[idx]
			if !fullRepaint && cell == r.lastFrame[idx] {
				x++
				continue
			}

			fg, bg := term.ResolveCellColors(cell, buf.Bg, buf.Fg)
			fmt.Fprintf(r.out, "\x1b[%d;%dH%s", y+1, x+1, term.StyleANSI(fg, bg))

			for x < w {
				idx2 := rowStart + x
				c2 := cells[idx2]
				if !fullRepaint && c2 == r.lastFrame[idx2] {
					break
				}
				f2, b2 := term.ResolveCellColors(c2, buf.Bg, buf.Fg)
				if f2 != fg || b2 != bg {
					break
				}
				ch := c2.Ch
				if ch == 0 {
					ch = ' '
				}
				r.out.WriteRune(ch)
				x++
			}
		}
	}

	copy(r.lastFrame, cells)
}
