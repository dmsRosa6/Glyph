package widgets

import (
	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type Border struct {
	base.BaseNode
	borderStyle BorderStyle
	thickness   int
}

type BorderConfig struct {
	Thickness   int
	BorderStyle BorderStyle
	Style       framework.Style
	Layer       int
}

func DefaultBorderConfig() BorderConfig {
	return BorderConfig{
		Thickness:   1,
		BorderStyle: SingleLine,
		Style:       framework.Style{Bg: core.Transparent, Fg: core.White},
	}
}

func NewBorder(bounds *geom.Bounds, cfg BorderConfig) (*Border, error) {
	if cfg.Thickness < 1 {
		panic("border thickness must be >= 1")
	}

	style := cfg.BorderStyle
	if style == (BorderStyle{}) {
		style = EmptyBorder
	}

	bn, err := base.NewBaseNode(bounds, framework.Anchor{}, cfg.Style, cfg.Layer)
	if err != nil {
		return nil, err
	}

	return &Border{
		BaseNode:    bn,
		borderStyle: style,
		thickness:   cfg.Thickness,
	}, nil
}

func (b *Border) Draw(buf *core.Buffer, vec geom.Vector) {
	s := b.Style()
	pos := b.ComputedPos()
	ox, oy := pos.X, pos.Y
	w, h := b.Size()

	for layer := 0; layer < b.thickness; layer++ {
		x0 := ox + layer
		y0 := oy + layer
		x1 := ox + w - 1 - layer
		y1 := oy + h - 1 - layer

		// corners
		buf.Set(vec.X+x0, vec.Y+y0, b.borderStyle.TopLeft, s.Bg, s.Fg)
		buf.Set(vec.X+x1, vec.Y+y0, b.borderStyle.TopRight, s.Bg, s.Fg)
		buf.Set(vec.X+x0, vec.Y+y1, b.borderStyle.BottomLeft, s.Bg, s.Fg)
		buf.Set(vec.X+x1, vec.Y+y1, b.borderStyle.BottomRight, s.Bg, s.Fg)

		// top & bottom edges
		for x := x0 + 1; x < x1; x++ {
			buf.Set(vec.X+x, vec.Y+y0, b.borderStyle.Horizontal, s.Bg, s.Fg)
			buf.Set(vec.X+x, vec.Y+y1, b.borderStyle.Horizontal, s.Bg, s.Fg)
		}

		// left & right edges
		for y := y0 + 1; y < y1; y++ {
			buf.Set(vec.X+x0, vec.Y+y, b.borderStyle.Vertical, s.Bg, s.Fg)
			buf.Set(vec.X+x1, vec.Y+y, b.borderStyle.Vertical, s.Bg, s.Fg)
		}
	}
}

func (b *Border) SetBorderStyle(s BorderStyle) {
	if s == (BorderStyle{}) {
		s = EmptyBorder
	}
	b.borderStyle = s
}
