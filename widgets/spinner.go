package widgets

import (
	"time"

	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type Spinner struct {
	base.BaseNode
	framework.SpinnerContext
	value          []rune
	TicksPerSecond int
}

type SpinnerConfig struct {
	SpinnerType    framework.SpinnerContext
	Pos            geom.Point
	Style          framework.Style
	Anchor         framework.Anchor
	Layer          int
	TicksPerSecond int
}

func NewSpinner(cfg SpinnerConfig) (*Spinner, error) {
	bounds := geom.NewBounds(cfg.Pos.X, cfg.Pos.Y, 1, cfg.SpinnerType.SpinnerLength())

	bn, err := base.NewBaseNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer, "Spinner")
	if err != nil {
		return nil, err
	}

	t := cfg.TicksPerSecond

	if t <= 0 {
		t = 1
	}

	spinner := &Spinner{BaseNode: bn, SpinnerContext: cfg.SpinnerType, value: []rune(cfg.SpinnerType.Cycle()), TicksPerSecond: t}

	// need to import from the app the cancel context
	go spinner.startCycle()

	return spinner, nil
}

func (t *Spinner) startCycle() {
	var ticker *time.Ticker
	ticker = time.NewTicker(time.Second / time.Duration(t.TicksPerSecond))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.value = []rune(t.SpinnerContext.Cycle())
		}
	}

}

func (t *Spinner) Draw(buf *core.Buffer, vec geom.Vector) {

	s := t.Style()
	pos := t.ComputedPos()
	x, y := pos.X, pos.Y

	for i := 0; i < t.SpinnerContext.SpinnerLength(); i++ {
		buf.Set(vec.X+x+i, vec.Y+y, t.value[i], s.Bg, s.Fg)
	}
}
