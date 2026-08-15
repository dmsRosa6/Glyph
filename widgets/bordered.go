package widgets

import (
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

// Bordered is a Border wrapped around a Panel, the Panel inset by
// however far the border pushes the interior in. Panel already owns
// "fill + content, correctly ordered, correctly delegated" -- the only
// thing that's actually Bordered's own concern is that inset math, and
// positioning the two pieces relative to each other.
//
// RECONSTRUCTED -- see the note atop this package's other composite
// widgets; not your real file.
type Bordered struct {
	*canvas.Container
	panel *Panel
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
		// own, it's purely structural. border and panel carry the real
		// colors.
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

	// inset accounts for BOTH the border's own Thickness (so the panel
	// never sits on top of border cells) AND the requested Padding (extra
	// breathing room beyond the border), applied identically to X, Y,
	// width, and height.
	inset := cfg.BorderConfig.Thickness + cfg.Padding
	innerW := bounds.W - 2*inset
	innerH := bounds.H - 2*inset

	panel, err := NewPanel(geom.NewBounds(inset, inset, innerW, innerH), PanelConfig{
		Style: cfg.Style,
	})
	if err != nil {
		return nil, err
	}

	outer.AddChild(border)
	outer.AddChild(panel)

	return &Bordered{Container: outer, panel: panel}, nil
}

// AddChild puts user content into the panel, not the outer wrapper --
// shadows the promoted Container.AddChild, which would otherwise add
// straight alongside border.
func (bx *Bordered) AddChild(child framework.Drawable) {
	bx.panel.AddChild(child)
}

func (bx *Bordered) RemoveChild(target framework.Drawable) {
	bx.panel.RemoveChild(target)
}

// Children shadows the promoted Container.Children (which would return
// [border, panel]) with the actual user-visible children.
func (bx *Bordered) Children() []framework.Drawable {
	return bx.panel.Children()
}
