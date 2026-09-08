package canvas

import (
	"errors"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/term"
)

type Canvas struct {
	root *Container
	Buf  *core.Buffer

	// RequestedWidth/RequestedHeight mirror CanvasConfig.Width/Height
	// exactly, including the <=0 sentinel meaning "no fixed size --
	// auto-size to whatever the terminal currently is." ApplySize reads
	// these on every resize to decide whether to track the terminal
	// exactly (<=0) or cap at a fixed size (>0).
	//
	// These must NEVER be set to a size resolved against the terminal
	// at any particular point in time, startup included -- doing that
	// used to silently convert an "auto" request into a fixed cap equal
	// to whatever the terminal happened to be when the app started.
	// Concretely: start the app in a normal (non-fullscreen) window,
	// then maximize it -- the canvas would stay stuck at the original
	// window's dimensions forever, because RequestedWidth/
	// RequestedHeight no longer read as <=0 for ApplySize to recognize
	// as "auto." Starting already maximized hid the bug entirely, since
	// there was nowhere bigger left to grow into.
	RequestedWidth  int
	RequestedHeight int
}

type CanvasConfig struct {
	Width, Height int
	Fg, Bg        core.Color
}

// NewCanvas resolves an initial width/height to build the very first
// root Container and Buf against, but that resolved value is NOT what
// gets stored in RequestedWidth/RequestedHeight below -- see those
// fields' doc comment for why conflating the two was the bug.
func NewCanvas(cfg CanvasConfig) (*Canvas, error) {
	// initW/initH are only for constructing the FIRST frame, before any
	// real resize has happened. cfg.Width/cfg.Height themselves --
	// including a <=0 "auto" request -- are preserved untouched below.
	initW, initH := cfg.Width, cfg.Height

	if initW <= 0 || initH <= 0 {
		size, err := term.TermSize()
		if err != nil {
			return nil, errors.New("could not retrieve terminal size")
		}
		if initW <= 0 {
			initW = size.Cols - 1
		}
		if initH <= 0 {
			initH = size.Rows - 1
		}
	}

	bg := cfg.Bg
	if bg == core.Transparent {
		bg = core.White
	}
	fg := cfg.Fg
	if fg == core.Transparent {
		fg = core.Black
	}

	root, err := NewContainer(geom.NewBounds(0, 0, initW, initH), ContainerConfig{
		Style: framework.Style{Bg: bg, Fg: fg},
	})
	if err != nil {
		return nil, err
	}

	return &Canvas{
		root: root,
		Buf:  core.NewBuffer(initW, initH, fg, bg),
		// cfg.Width/cfg.Height, NOT initW/initH: a <=0 "no fixed size"
		// request has to stay <=0 here, or ApplySize (below) can never
		// tell "auto-size to the terminal" apart from "fixed size that
		// happens to equal whatever the terminal was at startup" ever
		// again. See RequestedWidth's doc comment.
		RequestedWidth:  cfg.Width,
		RequestedHeight: cfg.Height,
	}, nil
}

// ApplySize recomputes the canvas's actual size from the current
// terminal size, termW x termH, respecting RequestedWidth/
// RequestedHeight as a cap: <=0 means "no fixed size, track the
// terminal exactly," anything else means "fixed size, but still never
// bigger than the terminal actually is right now."
//
// RequestedWidth/RequestedHeight must keep meaning EXACTLY what the
// caller asked for in CanvasConfig -- not a resolved, one-time snapshot
// of the terminal at startup. That distinction is the whole reason this
// method still works correctly on every subsequent resize: a caller
// that asked for "no fixed size" needs w/h to track termW/termH forever,
// growing and shrinking with it, not just once at construction.

func (c *Canvas) ApplySize(termW, termH int) {
	w := c.RequestedWidth
	if w <= 0 {
		w = termW
	} else {
		w = min(w, termW)
	}

	h := c.RequestedHeight
	if h <= 0 {
		h = termH
	} else {
		h = min(h, termH)
	}

	s := c.root.Style()
	c.root.Resize(w, h)
	c.Buf = core.NewBuffer(w, h, s.Fg, s.Bg)
	c.Compose()
}

func (c *Canvas) Restore() {
	s := c.root.Style()
	c.Buf.Clear(s.Fg, s.Bg)
}

func (c *Canvas) AddShape(s framework.Drawable) {
	c.root.AddChild(s)
}

// RemoveShape is AddShape's counterpart. Before this, once something
// was added to the canvas root there was no way to remove it short of
// reaching into the unexported canvas.root directly -- an asymmetric
// API that ruled out any dynamic UI that adds/removes top-level shapes
// at runtime.
func (c *Canvas) RemoveShape(s framework.Drawable) {
	c.root.RemoveChild(s)
}

func (c *Canvas) Shapes() []framework.Drawable {
	return c.root.Children()
}

func (c *Canvas) BringToFront(s framework.Drawable) {
	c.root.BringToFront(s)
}

func (c *Canvas) SendToBack(s framework.Drawable) {
	c.root.SendToBack(s)
}

func (c *Canvas) Compose() {
	c.Restore()
	c.root.Draw(c.Buf, geom.Vector{})
}

func (c *Canvas) CollectFocusable() []framework.Focusable {
	var out []framework.Focusable
	collectFocusable(c.root.Children(), &out)
	return out
}

func collectFocusable(children []framework.Drawable, out *[]framework.Focusable) {
	for _, child := range children {
		if f, ok := child.(framework.Focusable); ok {
			*out = append(*out, f)
			continue
		}
		if cl, ok := child.(framework.ChildrenLister); ok {
			collectFocusable(cl.Children(), out)
		}
	}
}

func (c *Canvas) SetContext(ctx framework.AppContext) {
	c.root.SetContext(ctx)
}

func (c *Canvas) SetParentStyle(s *framework.Style) {
	c.root.SetParentStyle(s)
}
