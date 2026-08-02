package canvas

import (
	"sync"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

// Text always draws with a transparent background, so it blends with
// whatever is behind it rather than punching an opaque box.
type Text struct {
	BaseNode

	mu    sync.RWMutex
	value string
}

type TextConfig struct {
	Value  string
	Fg     core.Color
	Anchor Anchor
	Layer  int
}

func NewText(pos *geom.Point, cfg TextConfig) (*Text, error) {
	bounds := geom.NewBounds(pos.X, pos.Y, len(cfg.Value), 1)
	style := Style{Bg: core.Transparent, Fg: cfg.Fg}

	base, err := newBaseNode(bounds, cfg.Anchor, style, cfg.Layer)
	if err != nil {
		return nil, err
	}

	return &Text{BaseNode: base, value: cfg.Value}, nil
}

func (t *Text) Draw(buf *core.Buffer, vec geom.Vector) {
	t.mu.RLock()
	value := t.value
	t.mu.RUnlock()

	s := t.Style()
	x, y := t.computedPos.X, t.computedPos.Y

	for i := 0; i < len(value); i++ {
		buf.Set(vec.X+x+i, vec.Y+y, rune(value[i]), s.Bg, s.Fg)
	}
}

// SetValue updates the text's content and asks the renderer for a
// redraw. Safe to call concurrently, including from a background
// goroutine -- this IS the "let a component refresh itself via a
// goroutine" hook: hold a reference to a Text, spin up a goroutine, call
// SetValue on whatever schedule you want.
//
//	clock, _ := canvas.NewText(pos, canvas.TextConfig{Value: "00:00:00"})
//	go func() {
//	    for range time.Tick(time.Second) {
//	        clock.SetValue(time.Now().Format("15:04:05"))
//	    }
//	}()
//
// The node's width is fixed at construction (from the initial value's
// length) since that's what everything else's bounds-checking and
// layout is computed against. A longer replacement is truncated to fit;
// construct with your expected max width if the value can grow.
func (t *Text) SetValue(v string) {
	t.mu.Lock()
	if len(v) > t.bounds.W {
		v = v[:t.bounds.W]
	}
	t.value = v
	t.mu.Unlock()

	t.Invalidate()
}

func (t *Text) Value() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.value
}