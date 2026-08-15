package widgets

import (
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

// Panel is a styled, filled rectangle you can add children to -- the
// "fill + content, correctly ordered, correctly delegated" primitive
// that used to be hand-derived inline inside Bordered's NewBox. Bordered
// is now just this plus a Border sized and inset around it; anything
// else that wants "an area with a real background and children" (a List
// row, say) can reach for this directly instead of a bare
// canvas.Container, which has no fill of its own at all.
type Panel struct {
	*canvas.Container
	content *canvas.Container
}

type PanelConfig struct {
	Style  framework.Style
	Layer  int
	Anchor framework.Anchor
}

func NewPanel(bounds *geom.Bounds, cfg PanelConfig) (*Panel, error) {
	outer, err := canvas.NewContainer(bounds, canvas.ContainerConfig{
		Style:  framework.Style{Bg: core.Transparent, Fg: core.Transparent},
		Layer:  cfg.Layer,
		Anchor: cfg.Anchor,
	})
	if err != nil {
		return nil, err
	}

	content, err := canvas.NewContainer(geom.NewBounds(0, 0, bounds.W, bounds.H), canvas.ContainerConfig{
		Style: cfg.Style,
	})
	if err != nil {
		return nil, err
	}

	// Container.Draw never paints its own area, only its children's --
	// fill is what actually renders cfg.Style's Bg. Sibling of content on
	// the outer container, not nested inside it, so it never shows up in
	// Panel.Children(), which must stay pure user content.
	fill, err := NewRect(geom.NewBounds(0, 0, bounds.W, bounds.H), RectConfig{
		Style: cfg.Style,
	})
	if err != nil {
		return nil, err
	}

	outer.AddChild(fill)
	outer.AddChild(content)

	return &Panel{Container: outer, content: content}, nil
}

func (p *Panel) AddChild(child framework.Drawable) {
	p.content.AddChild(child)
}

func (p *Panel) RemoveChild(target framework.Drawable) {
	p.content.RemoveChild(target)
}

func (p *Panel) Children() []framework.Drawable {
	return p.content.Children()
}
