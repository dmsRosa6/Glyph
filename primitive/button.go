package primitive

import (
	"fmt"

	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
)

type Button struct {
	mixin.FocusableNode
	label string
}

type ButtonConfig struct {
	Style      framework.Style
	Layer      int
	Anchor     framework.Anchor
	OnActivate mixin.FocusableActionFunc
}

func NewButton(bounds *geom.Bounds, label string, cfg ButtonConfig) (*Button, error) {
	bn, err := mixin.NewNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer, "Button")
	if err != nil {
		return nil, err
	}

	b := &Button{
		FocusableNode: mixin.NewFocusableNode(bn),
		label:         label,
	}

	if cfg.OnActivate != nil {
		b.BindAction(framework.KeyRune, func(a mixin.FocusableActionContext) (bool, error) {
			if a.Event().Rune != ' ' {
				return false, nil
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
