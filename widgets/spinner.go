package widgets

import (
	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type Spinner struct {
	base.BaseNode
	framework.SpinnerContext
	value string
}

type SpinnerConfig struct {
	SpinnerType framework.SpinnerContext
	Pos         geom.Point
	Fg          core.Color
	Anchor      framework.Anchor
	Layer       int
}

func NewSpinner(cfg SpinnerConfig) (*Spinner, error) {
	bounds := geom.NewBounds(cfg.Pos.X, cfg.Pos.Y, 1, cfg.SpinnerType.SpinnerLength())
	style := framework.Style{Bg: core.Transparent, Fg: cfg.Fg}

	bn, err := base.NewBaseNode(bounds, cfg.Anchor, style, cfg.Layer, "Spinner")
	if err != nil {
		return nil, err
	}

	return &Spinner{BaseNode: bn, SpinnerContext: cfg.SpinnerType, value: cfg.SpinnerType.Cycle()}, nil
}

func (t *Spinner) Draw(buf *core.Buffer, vec geom.Vector) {

	s := t.Style()
	pos := t.ComputedPos()
	x, y := pos.X, pos.Y

	for i := 0; i < t.SpinnerContext.SpinnerLength(); i++ {
		buf.Set(vec.X+x+i, vec.Y+y, rune(t.value[i]), s.Bg, s.Fg)
	}

	t.value = t.SpinnerContext.Cycle()
}
