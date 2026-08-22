package widgets

import (
	"fmt"
	"sync"

	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

// Text always draws with a transparent background, so it blends with
// whatever is behind it rather than punching an opaque box.
type Text struct {
	base.BaseNode

	mu    sync.RWMutex
	value []rune
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

	bn, err := base.NewBaseNode(bounds, cfg.Anchor, style, cfg.Layer, "Text")
	if err != nil {
		return nil, err
	}

	return &Text{BaseNode: bn, value: runes}, nil
}

func (t *Text) Draw(buf *core.Buffer, vec geom.Vector) {
	t.mu.RLock()
	value := t.value
	t.mu.RUnlock()

	s := t.Style()
	pos := t.ComputedPos()
	x, y := pos.X, pos.Y

	for i := 0; i < len(value); i++ {
		buf.Set(vec.X+x+i, vec.Y+y, t.value[i], s.Bg, s.Fg)
	}
}

func (t *Text) SetValue(v string) {
	t.mu.Lock()
	t.value = []rune(v)
	t.mu.Unlock()

	t.Logger().Debug(fmt.Sprintf("value set to %q", v))
}

func (t *Text) Value() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return string(t.value)
}
