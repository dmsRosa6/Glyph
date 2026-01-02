package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Text struct {
	Pos   geom.Point
	Value string

	Bg, Fg core.Color
}

func NewText(x, y int, value string, bg, fg core.Color) *Text {
	return &Text{
		Pos:   geom.Point{X: x, Y: y},
		Value: value,
		Bg: bg,
		Fg: fg,
	}
}

func (t *Text) Draw(buf *core.Buffer) {
	x := t.Pos.X
	y := t.Pos.Y

	for i := 0; i < len(t.Value); i++{
		
		if t.Fg.IsTransparent {
			continue
		}
	
		r := rune(t.Value[i])
		
		buf.Set(x+i, y, r, t.Bg, t.Fg)
	}
}

func (t *Text) MoveTo(p geom.Point) {
	if t.Pos != p {
		t.Pos = p
	}
}

func (t *Text) Translate(v geom.Vector) {
	t.Pos = t.Pos.Add(v)
}
