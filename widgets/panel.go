package widgets

import (
	"github.com/dmsRosa6/glyph/canvas"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/primitive"
)

type Panel struct {
	*canvas.Container
	fill    *primitive.Rect
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

	fill, err := primitive.NewRect(geom.NewBounds(0, 0, bounds.W, bounds.H), primitive.RectConfig{
		Style: cfg.Style,
	})
	if err != nil {
		return nil, err
	}

	outer.AddChild(fill)
	outer.AddChild(content)

	return &Panel{Container: outer, fill: fill, content: content}, nil
}

func (p *Panel) Resize(w, h int) {
	p.Container.Resize(w, h)
	p.fill.Resize(w, h)
	p.content.Resize(w, h)
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
