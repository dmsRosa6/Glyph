package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Box struct{
	border *Border
	composite *Composite
}

//TODO Add colors to the composite
func NewBox(x, y, w, h, thickness int, bg, fg, borderBg, borderFg core.Color) *Box{

	border := NewBorder(x, y, w, h,thickness , ' ', borderFg, borderBg)

	composite := NewComposite(x+thickness, y+thickness, w-2*thickness, h-2*thickness)

	return &Box{
		border: border,
		composite: composite,
	}
}

func (b *Box) Draw(buf *core.Buffer){
	b.border.Draw(buf)
	b.composite.Draw(buf)
}

func (r *Box) IsInBounds(parent geom.Bounds) bool{
	return r.composite.IsInBounds(parent) && r.border.IsInBounds(parent)	
}