package primitive

import (
	"fmt"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
)

// Text is mutable, single-line styled text -- the "value can change
// after construction, typically from a background goroutine" half of
// glyph's text-widget split. Use StaticText instead for a label set
// once and never touched again: StaticText pays for no lock at all,
// where Text's embedded mixin.TextBuffer necessarily does, since a
// background goroutine (see the clock example) needs to be able to
// call SetValue safely from outside the render/input goroutines.
//
// Text is deliberately single-line only, with no wrapping or
// justification -- NewText builds its bounds as H: 1, and SetValue
// never touches height. If a value containing '\n' is passed in, each
// rune (including the newline itself) is drawn as one cell wide, in a
// straight horizontal line -- it will NOT visually wrap to a second
// row. Use MultilineText for anything that needs real line breaks or
// word-wrapping.
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

// SetValue shadows the promoted mixin.TextBuffer.SetValue only to add
// the debug log line every other mutator in this codebase gives you --
// the underlying replace-under-lock behavior is entirely TextBuffer's.
func (t *Text) SetValue(v string) {
	t.TextBuffer.SetValue(v)
	t.Logger().Debug(fmt.Sprintf("value set to %q", v))
}
