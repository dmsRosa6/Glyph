package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Canvas struct {
	Buf    *core.Buffer
	Shapes []Drawable
	Bg     core.Color
	Fg     core.Color

	RequestedWidth int
	RequestedHeight int

    IsDirty bool
}

//TODO if the w,h logic changes dont forget the bound check
func NewCanvas(w, h int, fg, bg core.Color) *Canvas {

    c := &Canvas{
        Buf:             core.NewBuffer(w, h, fg, bg),
        Shapes:          []Drawable{},
        Bg:              bg,
        Fg:              fg,
        RequestedWidth:  w,
        RequestedHeight: h,
    }

    return c
}

func (c *Canvas) ApplySize(termW, termH int) {
    w := c.RequestedWidth
    h := c.RequestedHeight

    if w <= 0 {
        w = termW
    }
    if h <= 0 {
        h = termH
    }

    actualW := min(termW, w)
    actualH := min(termH, h)

    c.Buf = core.NewBuffer(actualW, actualH, c.Fg, c.Bg)
    c.IsDirty = true

    c.Compose()
}

func (c *Canvas) Restore() {
	c.Buf.Clear()
}

func (c *Canvas) AddShape(s Drawable) {
    if !s.IsInBounds(
        geom.Bounds{
            Pos: geom.Point{X:0, Y:0},
            W: c.RequestedWidth,
            H: c.RequestedHeight}){
		panic("Shape out of composite bounds")
	}	

	c.Shapes = append(c.Shapes, s)
}

func (c *Canvas) Compose() {
    
    c.Restore()

	for _, s := range c.Shapes {
		s.Draw(c.Buf)
    }
}
