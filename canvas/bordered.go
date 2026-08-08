package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

// Bordered wraps a single child Drawable with a frame. This replaces the
// old Box, which hardcoded "a Border plus a Composite" as its only two
// parts. Bordered doesn't care what its content is -- a Container, a
// bare Text, a Rect, even another Bordered -- so a border is now
// something you can wrap around anything, not a feature only boxes get.
//
// If the wrapped content implements Composable, Bordered exposes
// AddChild/RemoveChild itself by delegating to it, so `NewBox(...)`
// (a Bordered wrapping a Container) keeps the familiar
// `box.AddChild(child)` call site.
type Bordered struct {
	BaseNode
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
	base, err := newBaseNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer)
	if err != nil {
		return nil, err
	}

	borderBounds := geom.NewBounds(0, 0, bounds.W, bounds.H)
	border, err := NewBorder(borderBounds, cfg.BorderConfig)
	if err != nil {
		return nil, err
	}

	b := &Bordered{
		BaseNode: base,
		border:   border,
		content:  content,
	}

	border.SetParentStyle(b.ResolvedStyle())
	content.SetParentStyle(b.ResolvedStyle())

	return b, nil
}

func (b *Bordered) Draw(buf *core.Buffer, vec geom.Vector) {
	v := geom.Vector{X: vec.X + b.computedPos.X, Y: vec.Y + b.computedPos.Y}

	if l, ok := b.content.(framework.Layoutable); ok {
		l.Layout(b.LocalFrame())
	}
	b.content.Draw(buf, v)
	b.border.Draw(buf, v)
}

// AddChild delegates to the wrapped content if it's Composable (e.g. a
// Container). Panics if the content doesn't accept children -- e.g.
// wrapping a bare Text in a border and then calling AddChild on it isn't
// meaningful.
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