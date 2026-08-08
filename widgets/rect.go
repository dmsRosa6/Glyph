package widgets

import (
	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type Rect struct {
	base.BaseNode
	ch   rune
	clip geom.Bounds
}

type RectConfig struct {
	Ch    rune
	Style framework.Style

	Anchor framework.Anchor
	Layer  int
}

func NewRect(bounds *geom.Bounds, cfg RectConfig) (*Rect, error) {
	ch := cfg.Ch
	if ch == 0 {
		ch = ' '
	}

	bn, err := base.NewBaseNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer)
	if err != nil {
		return nil, err
	}

	return &Rect{
		BaseNode: bn,
		ch:       ch,
		clip:     *geom.NewBounds(-1, -1, -1, -1),
	}, nil
}

func (r *Rect) Draw(buf *core.Buffer, vec geom.Vector) {
	pos := r.ComputedPos()
	x, y := pos.X, pos.Y
	w, h := r.Size()

	if r.clip.ValidateNoPanic() {
		x += r.clip.Pos.X
		y += r.clip.Pos.Y
		w = r.clip.W
		h = r.clip.H
	}

	s := r.Style()

	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			buf.Set(vec.X+x+dx, vec.Y+y+dy, r.ch, s.Bg, s.Fg)
		}
	}
}

func (r *Rect) SetClip(c geom.Bounds) {
	c.ValidateIfInsideBounds(*r.Bounds())
	r.clip = c
}