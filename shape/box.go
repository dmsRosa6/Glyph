package shape

import "github.com/dmsRosa6/glyph/core"

type Box struct {
    r Rect
}


func (b *Box) Draw(buf *core.Buffer) {
    r := b.r

    for x := r.X; x < r.X+r.W; x++ {
        buf.Set(x, r.Y, '#')
        buf.Set(x, r.Y+r.H-1, '#')
    }
    for y := r.Y; y < r.Y+r.H; y++ {
        buf.Set(r.X, y, '#')
        buf.Set(r.X+r.W-1, y, '#')
    }
}
