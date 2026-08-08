package widgets

import (
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type BoxConfig struct {
	BorderConfig BorderConfig

	Padding int

	Style  framework.Style
	Anchor framework.Anchor
	Layer  int
}

// NewBox is a convenience constructor for the common case: a bordered,
// padded container that holds freely-positioned children. There's no
// bespoke Box type anymore -- this just wires up a Container (the
// content) inside a Bordered (the frame), which is exactly what the old
// hand-written Box did, minus the duplicated layering/style-cascade code.
func NewBox(bounds *geom.Bounds, cfg BoxConfig) (*Bordered, error) {
	if cfg.Padding < 0 {
		panic("padding must be >= 0")
	}

	innerW := bounds.W - 2*cfg.Padding
	innerH := bounds.H - 2*cfg.Padding
	if innerW < 0 || innerH < 0 {
		panic("padding too large for box bounds")
	}

	contentBounds := geom.NewBounds(cfg.Padding, cfg.Padding, innerW, innerH)
	content, err := canvas.NewContainer(contentBounds, canvas.ContainerConfig{Layer: cfg.Layer})
	if err != nil {
		return nil, err
	}

	return NewBordered(bounds, content, BorderedConfig{
		BorderConfig: cfg.BorderConfig,
		Style:        cfg.Style,
		Anchor:       cfg.Anchor,
		Layer:        cfg.Layer,
	})
}
