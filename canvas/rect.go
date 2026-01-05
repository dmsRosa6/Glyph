package canvas

import (
	"github.com/dmsRosa6/glyph/core"
	"github.com/dmsRosa6/glyph/geom"
)

type Rect struct {
	Bounds geom.Bounds
	Ch     rune
	Fg, Bg core.Color
}

type RectConfig struct {
    Bounds geom.Bounds
    Ch     rune
    Fg, Bg core.Color
}

func NewRect(cfg RectConfig) *Rect {
    return &Rect{
        Bounds: cfg.Bounds,
        Ch:     cfg.Ch,
        Fg:     cfg.Fg,
        Bg:     cfg.Bg,
    }
}

func (r *Rect) Draw(buf *core.Buffer) {
    for y := r.Bounds.Pos.Y; y < r.Bounds.Pos.Y+r.Bounds.H; y++ {
            for x := r.Bounds.Pos.X; x < r.Bounds.Pos.X+r.Bounds.W; x++ {

                buf.Set(x, y, r.Ch, r.Bg, r.Fg)
            }
    }
}

func (r *Rect) IsInBounds(parent geom.Bounds) bool{
	if r.Bounds.Pos.X < parent.Pos.X {
		return false
	}

	if r.Bounds.Pos.Y < parent.Pos.Y {
		return false
	}

	if r.Bounds.Pos.Y + r.Bounds.H > parent.Pos.Y + parent.H {
		return false
	}

	if r.Bounds.Pos.X + r.Bounds.W > parent.Pos.X + parent.W {
		return false
	}

	return true
}


func (r *Rect) MoveTo(p geom.Point) {
    r.Bounds.Pos = p
}

func (r *Rect) Translate(v geom.Vector) {
    r.Bounds.Pos = r.Bounds.Pos.Add(v)
}
