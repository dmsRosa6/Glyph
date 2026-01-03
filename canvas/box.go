package canvas

import (
	"github.com/dmsRosa6/glyph/core"
)

type Box struct{
	border *Border
	rect *Rect
}


func NewBox(x, y, w, h, thicknessX, thicknessY int, bg, fg, borderBg, borderFg core.Color) *Box{

	border := NewBorder(x, y, w, h,thicknessX, thicknessY, ' ', borderFg, borderBg)

	rect := NewRect(x+thicknessX, y+thicknessY, w-2*thicknessX, h-2*thicknessY, ' ', fg, bg)

	return &Box{
		border: border,
		rect: rect,
	}
}

func (b *Box) Draw(buf *core.Buffer){
	b.border.Draw(buf)
	b.rect.Draw(buf)
}