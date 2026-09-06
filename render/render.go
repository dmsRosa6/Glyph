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

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

// NewRenderer takes the already-constructed RenderMode directly rather
// than a loose (mode, fps) pair it used to re-derive internally by
// calling FixedFPSMode/OnDemandMode a second time. That second call was
// redundant whenever the caller went through the constructors properly
// (the value was already valid) and actively wasteful when it didn't
// (e.g. it silently discarded the caller's OnDemandMode() Redraw
// channel and allocated a fresh one). Since FixedFPSMode now returns an
// error instead of panicking, NewRenderer can no longer call it
// unchecked anyway -- taking the pre-built value sidesteps needing to
// propagate that error through here at all.
//
// mode is validated here rather than trusted: RenderMode's fields are
// unexported (see rendermode.go), but the all-zero RenderMode{} is
// still legal from any package and reads as an unconfigured FixedFPS
// mode with Fps == 0 -- exactly what AppConfig{} produces when
// RenderMode is left unset. Left unchecked, that reaches Run's
// `time.NewTicker(time.Second / time.Duration(r.fps))` as a
// divide-by-zero panic. A mode built by hand outside this package
// (impossible today with unexported fields, but this guards the
// zero-value case regardless) could similarly leave redraw nil and
// hang Run forever on startup. Both are reported here instead.
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
		// 1000: click/release reporting. 1002: also report motion while
		// a button is held (drag). 1006: SGR extended coordinate mode --
		// modern, unambiguous, no 223-column limit like the legacy
		// 1005/1015 modes restore() still defensively disables below.
		//
		// Deliberately NOT enabling 1003 (report every mouse move even
		// with no button held): that would flood input.Manager's
		// 16-slot event buffer (Events() silently drops on a full
		// buffer, see Manager.send) under ordinary mouse movement, for
		// a feature (hover tracking) nothing in this framework consumes.
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

	fmt.Fprint(r.out, "\x1b[H")

	c.Compose()

	r.Flush(c.Buf)
	r.out.Flush()
}

// restore unconditionally disables all six mouse-reporting modes,
// regardless of mouseEnabled -- 1000/1002/1006 are the ones Init() may
// have turned on above; 1003/1005/1015 are never enabled by this
// package at all, but disabling an already-disabled mode is a harmless
// no-op, and this is cheap insurance against mouse-tracking state left
// behind by some OTHER program that ran in this terminal before glyph
// did (a crashed prior TUI app, for instance) -- not just this run's
// own state.
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
	cells, width, height := buf.GetCells()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := cells[y][x]
			fmt.Fprintf(r.out, "\x1b[%d;%dH%s", y+1, x+1, term.CellToANSI(*cell, buf.Fg, buf.Bg))
		}
	}
}
