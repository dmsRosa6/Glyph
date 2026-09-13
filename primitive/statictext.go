package primitive

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
)

// StaticText is immutable, single-line styled text -- set once at
// construction, never touched again. No mutex, no mixin.TextBuffer:
// the value is written exactly once, before this leaf is ever shared
// across goroutines (construction happens on whichever goroutine
// builds the UI; nothing reads it until it's attached to a Canvas and
// drawn from the render goroutine), which is the same "safe to publish
// a plain value with no lock" guarantee Go's memory model already
// gives an unexported field written once before being handed off. So
// there's no lock to pay for on every Draw the way Text (see its own
// doc comment) has to.
//
// There is deliberately no SetValue here. If you need to change the
// value later -- even once -- use Text instead. StaticText existing at
// all IS the promise that this particular label never will.
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
