package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Border struct{
	Pos geom.Point
	Bounds geom.Bounds
	Ch rune
	Fg, Bg core.Color
}

func NewBorder(x, y, w, h int, ch rune, fg, bg core.Color) *Border{
	return &Border{
		Bounds: geom.Bounds{
			Pos: geom.Point{X: x, Y: y},
			W:   w,
			H:   h,
		},
		Ch:     ch,
		Fg:     fg,
		Bg:     bg,
	}
}

func (r *Border) Draw(buf *core.Buffer){
	
	for x := r.Bounds.Pos.X; x < r.Bounds.Pos.X+r.Bounds.W; x++ {
		buf.Set(x, r.Bounds.Pos.Y, r.Ch, r.Bg, r.Fg)
		buf.Set(x, r.Bounds.Pos.Y+r.Bounds.H-1, r.Ch, r.Bg, r.Fg)
	}
	
	for y := r.Bounds.Pos.Y; y < r.Bounds.Pos.Y+r.Bounds.H; y++ {
		buf.Set(r.Bounds.Pos.X, y, r.Ch, r.Bg, r.Fg)
		buf.Set(r.Bounds.Pos.X+r.Bounds.W-1, y, r.Ch, r.Bg, r.Fg)
	}
}

func (r *Border) MoveTo(p geom.Point) {
    r.Bounds.Pos = p
}

func (r *Border) Translate(v geom.Vector) {
    r.Bounds.Pos = r.Bounds.Pos.Add(v)
}
