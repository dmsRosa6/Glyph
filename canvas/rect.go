package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type Rect struct {
	BaseNode
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

	base, err := newBaseNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer)
	if err != nil {
		return nil, err
	}

	return &Rect{
		BaseNode: base,
		ch:       ch,
		clip:     *geom.NewBounds(-1, -1, -1, -1),
	}, nil
}

func (r *Rect) Draw(buf *core.Buffer, vec geom.Vector) {
	x, y := r.computedPos.X, r.computedPos.Y
	w, h := r.bounds.W, r.bounds.H

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
	c.ValidateIfInsideBounds(*r.bounds)
	r.clip = c
}