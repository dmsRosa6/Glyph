package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Button struct {
    FocusableBaseNode
    label string
}

func NewButton(bounds *geom.Bounds, label string, cfg ContainerConfig) (*Button, error) {
    base, err := newBaseNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer)
    if err != nil {
        return nil, err
    }

    b := &Button{
        FocusableBaseNode: newFocusableBaseNode(base),
        label:             label,
    }

    return b, nil
}

func (b *Button) Draw(buf *core.Buffer, vec geom.Vector) {
	panic("Not implemented")
}