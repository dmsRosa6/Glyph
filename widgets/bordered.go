package widgets

import (
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

// Bordered wraps a decorative Border around an inner content Container,
// inset by Padding.
//
// v2: Bordered embeds *canvas.Container directly instead of hand-rolling
// its own base.Propagator. border and content are added to that embedded
// Container via the ordinary AddChild path -- so Draw, SetParentStyle,
// SetInvalidator, SetLogChannel, and SetLayer are ALL inherited for
// free, already correct, because Container's own propagation (itself
// Propagator-backed) already does exactly this job. The only methods
// Bordered needs to write are the three that must redirect from "the
// outer wrapper" to "the inner content slot": AddChild, RemoveChild,
// Children.
//
// RECONSTRUCTED -- see the note atop this package's other composite
// widgets; not your real file.
type Bordered struct {
	*canvas.Container
	content *canvas.Container
}

type BoxConfig struct {
	Padding      int
	Style        framework.Style
	BorderConfig BorderConfig
	Layer        int
	Anchor       framework.Anchor
}

func NewBox(bounds *geom.Bounds, cfg BoxConfig) (*Bordered, error) {
	outer, err := canvas.NewContainer(bounds, canvas.ContainerConfig{
		// Fully transparent: this wrapper has no visual style of its
		// own, it's purely structural. border and content carry the
		// real colors.
		Style:  framework.Style{Bg: core.Transparent, Fg: core.Transparent},
		Layer:  cfg.Layer,
		Anchor: cfg.Anchor,
	})
	if err != nil {
		return nil, err
	}

	border, err := NewBorder(geom.NewBounds(0, 0, bounds.W, bounds.H), cfg.BorderConfig)
	if err != nil {
		return nil, err
	}

	pad := cfg.Padding
	content, err := canvas.NewContainer(geom.NewBounds(pad, pad, bounds.W-2*pad, bounds.H-2*pad), canvas.ContainerConfig{
		Style: cfg.Style,
	})
	if err != nil {
		return nil, err
	}

	outer.AddChild(border)
	outer.AddChild(content)

	return &Bordered{Container: outer, content: content}, nil
}

// AddChild puts user content into content, not the outer wrapper --
// shadows the promoted Container.AddChild, which would otherwise add
// straight alongside border.
func (bx *Bordered) AddChild(child framework.Drawable) {
	bx.content.AddChild(child)
}

func (bx *Bordered) RemoveChild(target framework.Drawable) {
	bx.content.RemoveChild(target)
}

// Children shadows the promoted Container.Children (which would return
// [border, content]) with the actual user-visible children, so focus
// collection walks the right level.
func (bx *Bordered) Children() []framework.Drawable {
	return bx.content.Children()
}
