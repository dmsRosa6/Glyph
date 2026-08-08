package widgets

import (
	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type Bordered struct {
	base.BaseNode
	border  *Border
	content framework.Drawable
}

type BorderedConfig struct {
	BorderConfig BorderConfig
	Style        framework.Style
	Anchor       framework.Anchor
	Layer        int
}

func NewBordered(bounds *geom.Bounds, content framework.Drawable, cfg BorderedConfig) (*Bordered, error) {
	bn, err := base.NewBaseNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer)
	if err != nil {
		return nil, err
	}

	borderBounds := geom.NewBounds(0, 0, bounds.W, bounds.H)
	border, err := NewBorder(borderBounds, cfg.BorderConfig)
	if err != nil {
		return nil, err
	}

	b := &Bordered{
		BaseNode: bn,
		border:   border,
		content:  content,
	}

	border.SetParentStyle(b.ResolvedStyle())
	content.SetParentStyle(b.ResolvedStyle())

	return b, nil
}

func (b *Bordered) Draw(buf *core.Buffer, vec geom.Vector) {
	pos := b.ComputedPos()
	v := geom.Vector{X: vec.X + pos.X, Y: vec.Y + pos.Y}

	if l, ok := b.content.(framework.Layoutable); ok {
		l.Layout(b.LocalFrame())
	}
	b.content.Draw(buf, v)
	b.border.Draw(buf, v)
}

func (b *Bordered) AddChild(child framework.Drawable) {
	c, ok := b.content.(framework.Composable)
	if !ok {
		panic("Bordered content does not accept children")
	}
	c.AddChild(child)
}

func (b *Bordered) RemoveChild(target framework.Drawable) {
	if c, ok := b.content.(framework.Composable); ok {
		c.RemoveChild(target)
	}
}

func (b *Bordered) SetLayer(l int) error {
	if err := b.BaseNode.SetLayer(l); err != nil {
		return err
	}
	if err := b.border.SetLayer(l); err != nil {
		return err
	}
	return b.content.SetLayer(l)
}

func (b *Bordered) SetParentStyle(s *framework.Style) {
	b.BaseNode.SetParentStyle(s)
	b.border.SetParentStyle(b.ResolvedStyle())
	b.content.SetParentStyle(b.ResolvedStyle())
}

func (b *Bordered) SetInvalidator(fn func()) {
	b.BaseNode.SetInvalidator(fn)
	b.border.SetInvalidator(fn)
	b.content.SetInvalidator(fn)
}