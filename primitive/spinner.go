package primitive

import (
	"sync"
	"time"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
)

type Spinner struct {
	mixin.Node
	framework.SpinnerContext

	mu             sync.RWMutex
	value          []rune
	TicksPerSecond int

	startOnce sync.Once
	stopOnce  sync.Once
	stop      chan struct{}
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

	bn, err := mixin.NewNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer, "Spinner")
	if err != nil {
		return nil, err
	}

	t := cfg.TicksPerSecond
	if t <= 0 {
		t = 1
	}

	spinner := &Spinner{
		Node:           bn,
		SpinnerContext: cfg.SpinnerType,
		value:          []rune(cfg.SpinnerType.Cycle()),
		TicksPerSecond: t,
		stop:           make(chan struct{}),
	}

	return spinner, nil
}

func (t *Spinner) SetContext(ctx framework.AppContext) {
	t.Node.SetContext(ctx)
	t.startOnce.Do(func() {
		go t.startCycle(ctx.Lifecycle())
	})
}

func (t *Spinner) startCycle(appDone <-chan struct{}) {
	ticker := time.NewTicker(time.Second / time.Duration(t.TicksPerSecond))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.mu.Lock()
			t.value = []rune(t.SpinnerContext.Cycle())
			t.mu.Unlock()
			t.Invalidate()
		case <-appDone:
			return
		case <-t.stop:
			return
		}
	}
}

func (t *Spinner) Stop() {
	t.stopOnce.Do(func() {
		close(t.stop)
	})
}

func (t *Spinner) Draw(buf *core.Buffer, vec geom.Vector) {
	t.mu.RLock()
	value := t.value
	t.mu.RUnlock()

	s := t.Style()
	pos := t.ComputedPos()
	x, y := pos.X, pos.Y

	for i := 0; i < t.SpinnerContext.SpinnerLength(); i++ {
		buf.Set(vec.X+x+i, vec.Y+y, value[i], s.Bg, s.Fg)
	}
}
