package widgets

import (
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
	panic("Not implemented") // TODO: implement Button.Draw
}
