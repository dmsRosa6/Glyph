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

	RequestedWidth  int
	RequestedHeight int
}

type CanvasConfig struct {
	Width, Height int // 0 = fill available terminal size
	Fg, Bg        core.Color
}

func NewCanvas(cfg CanvasConfig) (*Canvas, error) {
	w, h := cfg.Width, cfg.Height

	if w <= 0 || h <= 0 {
		size, err := term.TermSize()
		if err != nil {
			return nil, errors.New("could not retrieve terminal size")
		}
		if w <= 0 {
			w = size.Cols - 1
		}
		if h <= 0 {
			h = size.Rows - 1
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

	root, err := NewContainer(geom.NewBounds(0, 0, w, h), ContainerConfig{
		Style: framework.Style{Bg: bg, Fg: fg},
	})
	if err != nil {
		return nil, err
	}

	return &Canvas{
		root:            root,
		Buf:             core.NewBuffer(w, h, fg, bg),
		RequestedWidth:  w,
		RequestedHeight: h,
	}, nil
}

// ApplySize recomputes the Canvas's actual size against the terminal's
// current dimensions. Called on every resize event, and once up front
// before the first frame. Auto dimensions (RequestedWidth/Height <= 0)
// always follow the terminal exactly; a fixed dimension is capped at the
// terminal's current size but never grows past its requested value.
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

// AddShape delegates straight to the root Container's AddChild: same
// bounds check, same style propagation, same invalidator propagation,
// same layer-sorted insertion every nested Container already gives its
// children. Nothing about being "the top" needs its own version of this.
func (c *Canvas) AddShape(s framework.Drawable) {
	c.root.AddChild(s)
}

// SetInvalidator wires up the redraw callback for the whole tree, called
// by Renderer.Run. Container.SetInvalidator already handles reaching
// children added before this was called (the normal order: build the
// tree, then hand it to a Renderer) -- Canvas doesn't need its own copy
// of that logic.
func (c *Canvas) SetInvalidator(fn func()) {
	c.root.SetInvalidator(fn)
}

// Shapes returns the top-level shapes currently on the canvas. This is
// deliberately a read-only accessor, not an exported slice -- direct
// mutation would skip AddShape's bounds check, style propagation, and
// invalidator wiring entirely.
func (c *Canvas) Shapes() []framework.Drawable {
	return c.root.Children()
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
			// A focusable node's children are reached via Enter(),
			// not flattened here -- so if it's ALSO a container, stop.
			continue
		}
		// Not focusable itself, but might still hold focusable
		// descendants deeper down -- ask if it can hand us its
		// children without caring what concrete type it is.
		if cl, ok := child.(framework.ChildrenLister); ok {
			collectFocusable(cl.Children(), out)
		}
	}
}

func (c *Canvas) SetLogChannel(ch chan<- core.AppLog) {
	c.root.SetLogChannel(ch)
}

func (c *Canvas) SetParentStyle(s *framework.Style) {
	c.root.SetParentStyle(s)
}
