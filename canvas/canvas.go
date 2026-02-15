package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/utils"
)

type Canvas struct {
    bounds *geom.Bounds
	Buf    *core.Buffer
	Shapes []Drawable
    style *Style
	RequestedWidth int
	RequestedHeight int
}

func NewCanvas(w, h int, fg, bg core.Color) *Canvas {

    if bg == (core.Color{}){
       bg = core.White 
    }

    if fg == (core.Color{}){
       fg = core.Black 
    }

    s := &Style{
        Bg: bg,
        Fg: fg,
    }

    c := &Canvas{
        bounds: geom.NewBounds(0,0,w,h),
        Shapes:          []Drawable{},
        Buf:             core.NewBuffer(w, h, fg, bg),
        style: s,
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

    c.Buf = core.NewBuffer(actualW, actualH, c.style.Fg, c.style.Bg)
    c.Buf.Activate()
    c.Compose()
}

func (c *Canvas) Restore() {
	c.Buf.Clear(c.style.Fg, c.style.Bg)
}

func (c *Canvas) AddShape(s Drawable) {
    if !s.IsInBounds(
        geom.Bounds{
            Pos: &geom.Point{X:0, Y:0},
            W: c.RequestedWidth,
            H: c.RequestedHeight}){
		panic("Shape out of composite bounds")
	}	

    s.SetParentStyle(c.style)

	c.Shapes = utils.InsertSortLayered(c.Shapes, s)
}

func (c *Canvas) Compose() {
    
    c.Restore()
	for _, s := range c.Shapes {
        l, ok := s.(Layoutable)

        if ok {
            l.Layout(*c.bounds)   
        }

		s.Draw(c.Buf, *geom.VectorFromPoint(*c.bounds.Pos))
    }
}
