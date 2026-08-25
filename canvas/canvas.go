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
