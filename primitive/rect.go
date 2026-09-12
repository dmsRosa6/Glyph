package primitive

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
	"github.com/dmsRosa6/glyph/mixin"
)

type Rect struct {
	mixin.Node
}

type RectConfig struct {
	Style  framework.Style
	Layer  int
	Anchor framework.Anchor
}

func NewRect(bounds *geom.Bounds, cfg RectConfig) (*Rect, error) {
	bn, err := mixin.NewNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer, "Rect")
	if err != nil {
		return nil, err
	}
	return &Rect{Node: bn}, nil
}

func (r *Rect) Draw(buf *core.Buffer, vec geom.Vector) {
	s := r.Style()
	pos := r.ComputedPos()
	w, h := r.Size()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			buf.Set(vec.X+pos.X+x, vec.Y+pos.Y+y, ' ', s.Bg, s.Fg)
		}
	}
}
