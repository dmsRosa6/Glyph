package widgets

import (
	"fmt"

	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type Button struct {
	base.FocusableBaseNode
	label string
}

type ButtonConfig struct {
	Style  framework.Style
	Layer  int
	Anchor framework.Anchor
	// OnActivate, if set, is bound to Space (KeyRune, rune ' '), not
	// Enter -- deliberately. Enter is a structural key (see
	// framework.IsStructuralKey), dispatched global-first by
	// App.Run: if an app merges app.NavActions() (which binds Enter to
	// drill into FocusContainers), a Button's own Enter binding would
	// silently never fire -- exactly the shadow-key conflict
	// base.Propagator's own warnShadowedKeys exists to warn about the
	// moment this Button is attached to a tree. Space is an ordinary
	// key, dispatched widget-first, so it always reaches this handler
	// regardless of what's bound globally. Bind Enter yourself via the
	// promoted BindAction if you want it too, with that shadow-key
	// tradeoff made explicitly rather than baked in silently here.
	//
	// nil (the default) leaves activation entirely up to the caller,
	// same as before this field existed -- constructing a Button gives
	// you rendering and focus, but no bound action, matching
	// FocusBehavior's mixin philosophy: nothing here is auto-included
	// that a caller didn't ask for.
	OnActivate base.FocusableActionFunc
}

func NewButton(bounds *geom.Bounds, label string, cfg ButtonConfig) (*Button, error) {
	bn, err := base.NewBaseNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer, "Button")
	if err != nil {
		return nil, err
	}

	b := &Button{
		FocusableBaseNode: base.NewFocusableBaseNode(bn),
		label:             label,
	}

	if cfg.OnActivate != nil {
		b.BindAction(framework.KeyRune, func(a base.FocusableActionContext) (bool, error) {
			if a.Event().Rune != ' ' {
				return false, nil // some other printable character -- not activation
			}
			return cfg.OnActivate(a)
		})
	}

	return b, nil
}

func (b *Button) Draw(buf *core.Buffer, vec geom.Vector) {
	s := b.Style()
	pos := b.ComputedPos()
	w, h := b.Size()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			buf.Set(vec.X+pos.X+x, vec.Y+pos.Y+y, ' ', s.Bg, s.Fg)
		}
	}

	label := []rune(b.label)
	if len(label) > w {
		label = label[:w]
	}
	labelY := pos.Y + h/2
	labelX := pos.X + (w-len(label))/2
	for i, r := range label {
		buf.Set(vec.X+labelX+i, vec.Y+labelY, r, s.Bg, s.Fg)
	}
}

func (b *Button) SetLabel(v string) {
	b.label = v
	b.Logger().Debug(fmt.Sprintf("label set to %q", v))
}

func (b *Button) Label() string {
	return b.label
}
