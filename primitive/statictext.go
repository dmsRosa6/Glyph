package primitive

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
)

// StaticText is immutable, single-line styled text set once at
// construction.
type StaticText struct {
	mixin.Node
	value []rune
}

type StaticTextConfig struct {
	Value  string
	Fg     core.Color
	Anchor framework.Anchor
	Layer  int
}

func NewStaticText(pos *geom.Point, cfg StaticTextConfig) (*StaticText, error) {
	runes := []rune(cfg.Value)
	bounds := geom.NewBounds(pos.X, pos.Y, len(runes), 1)
	style := framework.Style{Bg: core.Transparent, Fg: cfg.Fg}

	bn, err := mixin.NewNode(bounds, cfg.Anchor, style, cfg.Layer, "StaticText")
	if err != nil {
		return nil, err
	}

	return &StaticText{Node: bn, value: runes}, nil
}

func (t *StaticText) Draw(buf *core.Buffer, vec geom.Vector) {
	s := t.Style()
	pos := t.ComputedPos()
	x, y := pos.X, pos.Y

	for i, r := range t.value {
		buf.Set(vec.X+x+i, vec.Y+y, r, s.Bg, s.Fg)
	}
}

func (t *StaticText) Value() string {
	return string(t.value)
}
