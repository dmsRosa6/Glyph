package shape

import "github.com/dmsRosa6/glyph/core"

type Rect struct {
    X, Y   int
    W, H   int
    Filled bool
    Ch     rune
}

func NewRect(x, y, w, h int, ch rune, filled bool) *Rect {
    return &Rect{
        X:      x,
        Y:      y,
        W:      w,
        H:      h,
        Ch:     ch,
        Filled: filled,
    }
}

func (r *Rect) Draw(buf *core.Buffer) {
    if r.Filled {
        for y := r.Y; y < r.Y+r.H; y++ {
            for x := r.X; x < r.X+r.W; x++ {
                buf.Set(x, y, r.Ch)
            }
        }
    } else {
        for x := r.X; x < r.X+r.W; x++ {
            buf.Set(x, r.Y, r.Ch)
            buf.Set(x, r.Y+r.H-1, r.Ch)
        }
        
		for y := r.Y; y < r.Y+r.H; y++ {
            buf.Set(r.X, y, r.Ch)
            buf.Set(r.X+r.W-1, y, r.Ch)
        }
    }
}
