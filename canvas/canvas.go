package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/term"
)

type Canvas struct {
	Buf    *core.Buffer
	Shapes []Drawable
	Render *render.Renderer
	Bg     core.Color
	Fg     core.Color
	RequestedWidth int
	RequestedHeight int
}

func NewCanvas(w, h int, fg, bg core.Color) *Canvas {
    t, err := term.TermSize()
    if err != nil {
        panic("Could not get the terminal size")
    }

    if w <= 0 {
        w = t.Cols
    }
    if h <= 0 {
        h = t.Rows
    }

    actualWidth := min(t.Cols, w)
    actualHeight := min(t.Rows, h)

    c := &Canvas{
        Render:          render.NewRenderer(),
        Buf:             core.NewBuffer(actualWidth, actualHeight, fg, bg),
        Shapes:          []Drawable{},
        Bg:              bg,
        Fg:              fg,
        RequestedWidth:  w,
        RequestedHeight: h,
    }

    // Watch terminal resize
    term.WatchResize(func() {
        c.resize()
    })

    return c
}

func (c *Canvas) resize() {
    t, err := term.TermSize()
    if err != nil {
        panic("Could not get the terminal size")
    }

    actualWidth := min(t.Cols, c.RequestedWidth)
    actualHeight := min(t.Rows, c.RequestedHeight)

    c.Buf = core.NewBuffer(actualWidth, actualHeight, c.Fg, c.Bg)
    c.Draw()
}

func (c *Canvas) Resize(w, h int) {
    if w <= 0 || h <= 0 {
        panic("Canvas.Resize: width and height must be positive")
    }

    c.RequestedWidth = w
    c.RequestedHeight = h

    c.resize()
}


func (c *Canvas) Init() {
	c.Render.Init()
	c.Draw()
}

func (c *Canvas) Restore() {
	c.Render.Restore()
}

func (c *Canvas) AddShape(s Drawable) {
	c.Shapes = append(c.Shapes, s)
}

func (c *Canvas) Draw() {
    
    c.Buf.Clear()

	for _, s := range c.Shapes {
		s.Draw(c.Buf)
	}
	c.Render.Render(c.Buf)
}
