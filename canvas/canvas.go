package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/render"
)

type Canvas struct {
	Rect   Rect
	Buf    *core.Buffer
	Shapes []Drawable
	Render *render.Renderer
	Bg     core.Color
	Fg     core.Color
}

func NewCanvas(w, h int, fg, bg core.Color) *Canvas {
	return &Canvas{
		Render: render.NewRenderer(),
		Rect:   *NewRect(0, 0, w, h, rune(' '), false, fg, bg),
		Buf:    core.NewBuffer(w, h, fg, bg),
		Shapes: []Drawable{},
	}
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
