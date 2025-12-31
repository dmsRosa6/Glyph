package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Rect struct {
	Bounds geom.Bounds
	Ch     rune
	Filled bool
	Fg, Bg core.Color
}

func NewRect(x, y, w, h int, ch rune, filled bool, fg, bg core.Color) *Rect {
	return &Rect{
		Bounds: geom.Bounds{
			Pos: geom.Point{X: x, Y: y},
			W:   w,
			H:   h,
		},
		Ch:     ch,
		Filled: filled,
		Fg:     fg,
		Bg:     bg,
	}
}

func (r *Rect) Draw(buf *core.Buffer) {

    if r.Filled {
        for y := r.Bounds.Pos.Y; y < r.Bounds.Pos.Y+r.Bounds.H; y++ {
            for x := r.Bounds.Pos.X; x < r.Bounds.Pos.X+r.Bounds.W; x++ {
                buf.Set(x, y, r.Ch, r.Bg, r.Fg)
            }
        }
    } else {
        for x := r.Bounds.Pos.X; x < r.Bounds.Pos.X+r.Bounds.W; x++ {
            buf.Set(x, r.Bounds.Pos.Y, r.Ch, r.Bg, r.Fg)
            buf.Set(x, r.Bounds.Pos.Y+r.Bounds.H-1, r.Ch, r.Bg, r.Fg)
        }
        for y := r.Bounds.Pos.Y; y < r.Bounds.Pos.Y+r.Bounds.H; y++ {
            buf.Set(r.Bounds.Pos.X, y, r.Ch, r.Bg, r.Fg)
            buf.Set(r.Bounds.Pos.X+r.Bounds.W-1, y, r.Ch, r.Bg, r.Fg)
        }
    }
}

func (r *Rect) MoveTo(p geom.Point) {
    r.Bounds.Pos = p
}

func (r *Rect) Translate(v geom.Vector) {
    r.Bounds.Pos = r.Bounds.Pos.Add(v)
}
