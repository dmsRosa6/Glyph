package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	shape "github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/render"
)

type Canvas struct {
    Rect   shape.Rect
    Buf    *core.Buffer
    Shapes []shape.Drawable
    Render *render.Renderer
}

func NewCanvas(w, h int) *Canvas {
    return &Canvas{
        Render: render.NewRenderer(),
        Rect:   *shape.NewRect(0, 0, w, h, rune(' '),false),
        Buf:    core.NewBuffer(w, h),
        Shapes: []shape.Drawable{},
    }
}

func (c *Canvas) Init() {
    c.Render.Init()
}

func (c *Canvas) Restore() {
    c.Render.Restore()
}

func (c *Canvas) AddShape(s shape.Drawable) {
    c.Shapes = append(c.Shapes, s)
}

func (c *Canvas) Draw() {
    for _, s := range c.Shapes {
        s.Draw(c.Buf)
    }
    c.Render.Render(c.Buf)
}
