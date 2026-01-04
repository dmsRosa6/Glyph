package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Border struct{
	Bounds geom.Bounds
	Ch rune
	ThicknessX int
    ThicknessY int
	Fg, Bg core.Color
}

func NewBorder(x, y, w, h, thicknessX, thicknessY int, ch rune, fg, bg core.Color) *Border{
	return &Border{
		Bounds: geom.Bounds{
			Pos: geom.Point{X: x, Y: y},
			W:   w,
			H:   h,
		},
		ThicknessX: thicknessX,
		ThicknessY: thicknessY,
		Ch:     ch,
		Fg:     fg,
		Bg:     bg,
	}
}

func (r *Border) Draw(buf *core.Buffer) {
    x0, y0 := r.Bounds.Pos.X, r.Bounds.Pos.Y
    w, h := r.Bounds.W, r.Bounds.H

    // top border
    for y := y0; y < y0+r.ThicknessY; y++ {
        for x := x0; x < x0+w; x++ {
            buf.Set(x, y, r.Ch, r.Bg, r.Fg)
        }
    }

    // bottom border
    for y := y0+h-r.ThicknessY; y < y0+h; y++ {
        for x := x0; x < x0+w; x++ {
            buf.Set(x, y, r.Ch, r.Bg, r.Fg)
        }
    }

    // left border
    for x := x0; x < x0+r.ThicknessX; x++ {
        for y := y0+r.ThicknessY; y < y0+h-r.ThicknessY; y++ {
            buf.Set(x, y, r.Ch, r.Bg, r.Fg)
        }
    }

    // right border
    for x := x0+w-r.ThicknessX; x < x0+w; x++ {
        for y := y0+r.ThicknessY; y < y0+h-r.ThicknessY; y++ {
            buf.Set(x, y, r.Ch, r.Bg, r.Fg)
        }
    }
}

func (r *Border) MoveTo(p geom.Point) {
    r.Bounds.Pos = p
}

func (r *Border) Translate(v geom.Vector) {
    r.Bounds.Pos = r.Bounds.Pos.Add(v)
}


