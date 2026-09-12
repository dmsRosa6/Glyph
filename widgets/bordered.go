package widgets

import (
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/primitive"
)

type Bordered struct {
	*canvas.Container
	border *primitive.Border
	panel  *Panel
	inset  int
}

type BoxConfig struct {
	Padding      int
	Style        framework.Style
	BorderConfig primitive.BorderConfig
	Layer        int
	Anchor       framework.Anchor
}

func NewBox(bounds *geom.Bounds, cfg BoxConfig) (*Bordered, error) {
	outer, err := canvas.NewContainer(bounds, canvas.ContainerConfig{
		Style:  framework.Style{Bg: core.Transparent, Fg: core.Transparent},
		Layer:  cfg.Layer,
		Anchor: cfg.Anchor,
	})
	if err != nil {
		return nil, err
	}

	border, err := primitive.NewBorder(geom.NewBounds(0, 0, bounds.W, bounds.H), cfg.BorderConfig)
	if err != nil {
		return nil, err
	}

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

	return &Bordered{Container: outer, border: border, panel: panel, inset: inset}, nil
}

func (bx *Bordered) Resize(w, h int) {
	bx.Container.Resize(w, h)
	bx.border.Resize(w, h)
	bx.panel.Resize(w-2*bx.inset, h-2*bx.inset)
}

func (bx *Bordered) AddChild(child framework.Drawable) {
	bx.panel.AddChild(child)
}

func (bx *Bordered) RemoveChild(target framework.Drawable) {
	bx.panel.RemoveChild(target)
}

func (bx *Bordered) Children() []framework.Drawable {
	return bx.panel.Children()
}
