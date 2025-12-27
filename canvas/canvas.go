package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/render"
	"github.com/dmsRosa6/glyph/shape"
)

type Canvas struct {
    Rect   shape.Rect
    Buf    *core.Buffer
    Shapes []shape.Shape
    Render *render.Renderer
}

func NewCanvas(w, h int) *Canvas {
    return &Canvas{
        Render: render.NewRenderer(),
        Rect:   *shape.NewRect(0, 0, w, h, rune(' '),false),
        Buf:    core.NewBuffer(w, h),
        Shapes: []shape.Shape{},
    }
}

func (c *Canvas) Init() {
    c.Render.Init()
}

func (c *Canvas) Restore() {
    c.Render.Restore()
}

func (c *Canvas) AddShape(s shape.Shape) {
    c.Shapes = append(c.Shapes, s)
}

func (c *Canvas) Draw() {
    for _, s := range c.Shapes {
        s.Draw(c.Buf)
    }
    c.Render.Draw(c.Buf)
}
