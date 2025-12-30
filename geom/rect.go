package geom

import "github.com/dmsRosa6/glyph/core"

type Bounds struct {
    Pos Point
    W, H int }

type Rect struct {
    Pos Point
    W, H int
    Ch rune
    Filled bool
    Prev Bounds
}

func (r *Rect) MarkAsDirty() {
    r.Prev = Bounds{
        Pos: r.Pos,
        W : r.W,
        H : r.H,
    }
}


func NewRect(x, y, w, h int, ch rune, filled bool) *Rect {
    return &Rect{
        Pos: Point{X: x,Y: y},
        W:      w,
        H:      h,
        Ch:     ch,
        Filled: filled,
    }
}

func (r *Rect) Draw(buf *core.Buffer) {

    clearOldState := r.Prev != Bounds{}

    if clearOldState {

        prevPos := r.Prev.Pos
        
        for y := prevPos.Y; y < prevPos.Y+r.Prev.H; y++ {
            for x := prevPos.X; x < prevPos.X+r.Prev.W; x++ {
                buf.Set(x, y, ' ')
            }
        }
    }

    if r.Filled {
        for y := r.Pos.Y; y < r.Pos.Y+r.H; y++ {
            for x := r.Pos.X; x < r.Pos.X+r.W; x++ {
                buf.Set(x, y, r.Ch)
            }
        }
    } else {
        for x := r.Pos.X; x < r.Pos.X+r.W; x++ {
            buf.Set(x, r.Pos.Y, r.Ch)
            buf.Set(x, r.Pos.Y+r.H-1, r.Ch)
        }
        for y := r.Pos.Y; y < r.Pos.Y+r.H; y++ {
            buf.Set(r.Pos.X, y, r.Ch)
            buf.Set(r.Pos.X+r.W-1, y, r.Ch)
        }
    }

    if clearOldState {
        r.Prev = Bounds{}
    }
}

func (r *Rect) MoveTo(p Point) {
    r.MarkAsDirty()
    r.Pos = p
}

func (r *Rect) Translate(v Vector) {
    r.MarkAsDirty()
    r.Pos = r.Pos.Add(v)
}
