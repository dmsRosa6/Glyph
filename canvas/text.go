package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Text struct {
	Pos   geom.Point
	Value string

	dirty bool
}

func NewText(x, y int, value string) *Text {
	return &Text{
		Pos:   geom.Point{X: x, Y: y},
		Value: value,
		dirty: true,
	}
}

func (t *Text) Draw(buf *core.Buffer) {
	x := t.Pos.X
	y := t.Pos.Y

	for _, r := range t.Value {
		buf.Set(x, y, r, core.Color{}, core.Color{})
		x++
	}

	t.dirty = false
}

func (t *Text) MarkAsDirty() {
	t.dirty = true
}

func (t *Text) MoveTo(p geom.Point) {
	if t.Pos != p {
		t.Pos = p
		t.MarkAsDirty()
	}
}

func (t *Text) Translate(v geom.Vector) {
	t.Pos = t.Pos.Add(v)
	t.MarkAsDirty()
}
