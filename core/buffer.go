package core

type Buffer struct {
    W, H int
    Cells [][]*Cell
    Bg Color
    Fg Color
}

func NewBuffer(w, h int, fg, bg Color) *Buffer {
    cells := make([][]*Cell, h)
    for y := range h {
        cells[y] = make([]*Cell, w)
        for x := range w {
            cells[y][x] = NewCell(' ', fg, bg)
        }
    }

    return &Buffer{
        W:     w,
        H:     h,
        Fg: fg,
        Bg: bg,
        Cells: cells,
    }
}

func (b *Buffer) Clear() {
    for y := 0; y < b.H; y++ {
        for x := 0; x < b.W; x++ {
            b.Cells[y][x] = NewCell(' ', b.Fg, b.Bg)
        }
    }
}

func (b *Buffer) Set(x, y int, ch rune, bg, fg Color){
    if y >= b.H || x >= b.W {
        return
    }

    b.Cells[y][x] = NewCell(ch, fg, bg)
}

func (b *Buffer) Get(x, y int) *Cell{
    return b.Cells[y][x]
}
