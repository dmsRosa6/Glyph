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
	Width, Height int
	Fg, Bg        core.Color
}

func NewCanvas(cfg CanvasConfig) (*Canvas, error) {
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

		RequestedWidth:  cfg.Width,
		RequestedHeight: cfg.Height,
	}, nil
}

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
