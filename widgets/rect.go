package widgets

import (
	"github.com/dmsRosa6/glyph/base"
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/framework"
	"github.com/dmsRosa6/glyph/geom"
)

type Rect struct {
	base.BaseNode
}

type RectConfig struct {
	Style  framework.Style
	Layer  int
	Anchor framework.Anchor
}

func NewRect(bounds *geom.Bounds, cfg RectConfig) (*Rect, error) {
	bn, err := base.NewBaseNode(bounds, cfg.Anchor, cfg.Style, cfg.Layer, "Rect")
	if err != nil {
		return nil, err
	}
	return &Rect{BaseNode: bn}, nil
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
