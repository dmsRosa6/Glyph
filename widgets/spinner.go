package widgets

import (
	"sync"
	"time"

	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type Spinner struct {
	base.BaseNode
	framework.SpinnerContext

	mu             sync.RWMutex
	value          []rune
	TicksPerSecond int

	startOnce sync.Once
	stopOnce  sync.Once
	// stop ends startCycle independently of ctx.Lifecycle() (the
	// app-wide Done channel) -- see Stop's doc comment for why this
	// exists at all.
	stop chan struct{}
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

	spinner := &Spinner{
		BaseNode:       bn,
		SpinnerContext: cfg.SpinnerType,
		value:          []rune(cfg.SpinnerType.Cycle()),
		TicksPerSecond: t,
		stop:           make(chan struct{}),
	}

	return spinner, nil
}

func (t *Spinner) SetContext(ctx framework.AppContext) {
	t.BaseNode.SetContext(ctx)
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
			t.Invalidate() // was missing entirely before: OnDemand mode never saw the new frame without this
		case <-appDone:
			return
		case <-t.stop:
			return
		}
	}
}

// Stop ends this Spinner's own ticking goroutine independently of the
// whole app. Previously startCycle only ever exited via ctx.Lifecycle()
// -- the app-wide Done channel, closed once by App.Stop() -- so a
// Spinner removed from the tree mid-run (RemoveChild/Untrack) had no
// way to actually stop: its goroutine kept ticking and calling
// Invalidate() forever, a leak scoped to "until the whole app exits,"
// not "until this widget is done."
//
// Stop is idempotent (safe to call more than once) via stopOnce, and
// safe to call even on a Spinner that was never attached to a tree
// (startCycle never started, so this just closes a channel nothing is
// listening on yet -- harmless, and correctly makes a LATER SetContext
// a no-op-for-ticking-purposes too, since startCycle would select on an
// already-closed stop and return immediately).
//
// base.Propagator.Untrack also calls this automatically on any removed
// child implementing framework.Stoppable (Spinner does) -- so plain
// RemoveChild is enough on its own; calling Stop directly is only
// needed for a Spinner never added to a container in the first place,
// or for stopping one deliberately without removing it from the tree.
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
