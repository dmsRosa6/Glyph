package widgets

import (
	"fmt"

	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Button struct {
	base.FocusableBaseNode
	label string
}

func NewButton(bounds *geom.Bounds, label string, cfg canvas.ContainerConfig) (*Button, error) {
	bn, err := base.NewBaseNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer, "Button")
	if err != nil {
		return nil, err
	}

	b := &Button{
		FocusableBaseNode: base.NewFocusableBaseNode(bn),
		label:             label,
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
