package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Window struct{
	box *Box
	text *Text
}

func NewWindow(x, y, w, h int, bg, fg, borderBg, borderFg core.Color, bs BorderStyle, text string, textXOffset int, textLocation bool, textFg core.Color) *Window{

	if textXOffset > w || textXOffset+len(text) > w{
		panic("Text out of window")
	}

	b := NewBoxWithBorderStyle(x,y,w,h,1,bg,fg,borderBg,borderFg,bs)

	var textY int 

	if textLocation {
		textY = y
	}else{
		textY = y+h
	}

	t := NewText(x+textXOffset, textY, text, core.Transparent, textFg)

	return &Window{
		box: b,
		text: t,
	}
}

func (w *Window) Draw(buf *core.Buffer){
	w.box.Draw(buf)
	w.text.Draw(buf)
}

func (w *Window) IsInBounds(parent geom.Bounds) bool{
	return w.box.IsInBounds(parent)
}