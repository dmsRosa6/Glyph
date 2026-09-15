package primitive

import (
	"fmt"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
)

// Text is mutable, single-line styled text -- can be updated after
// construction.
type Text struct {
	mixin.Node
	mixin.TextBuffer
}

type TextConfig struct {
	Value  string
	Fg     core.Color
	Anchor framework.Anchor
	Layer  int
}

func NewText(pos *geom.Point, cfg TextConfig) (*Text, error) {
	runes := []rune(cfg.Value)
	bounds := geom.NewBounds(pos.X, pos.Y, len(runes), 1)
	style := framework.Style{Bg: core.Transparent, Fg: cfg.Fg}

	bn, err := mixin.NewNode(bounds, cfg.Anchor, style, cfg.Layer, "Text")
	if err != nil {
		return nil, err
	}

	t := &Text{Node: bn}
	t.TextBuffer.SetValue(cfg.Value)
	return t, nil
}

func (t *Text) Draw(buf *core.Buffer, vec geom.Vector) {
	value := t.Runes()

	s := t.Style()
	pos := t.ComputedPos()
	x, y := pos.X, pos.Y

	for i, r := range value {
		buf.Set(vec.X+x+i, vec.Y+y, r, s.Bg, s.Fg)
	}
}

func (t *Text) SetValue(v string) {
	t.TextBuffer.SetValue(v)
	t.Logger().Debug(fmt.Sprintf("value set to %q", v))
}
