package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type DrawableWrapper struct {
    Child Drawable
    Pos   geom.Point
    Clip  *geom.Bounds
}

func (w *DrawableWrapper) Draw(buf *core.Buffer) {
    x0, y0 := w.Pos.X, w.Pos.Y
    if w.Clip != nil {
        x0 += w.Clip.Pos.X
        y0 += w.Clip.Pos.Y
    }
    w.Child.Draw(buf, geom.Vector{X: x0, Y: y0})
}
