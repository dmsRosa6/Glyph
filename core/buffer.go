package core

type Buffer struct {
    W, H int
    Cells [][]*Cell
}

func NewBuffer(w, h int) *Buffer {
    cells := make([][]*Cell, h)
    for y := 0; y < h; y++ {
        cells[y] = make([]*Cell, w)
        for x := 0; x < w; x++ {
            cells[y][x] = NewCell(' ', *NewColor(0, 0, 0), *NewColor(127, 127, 127))
        }
    }

    return &Buffer{
        W:     w,
        H:     h,
        Cells: cells,
    }
}

func (b *Buffer) Clear() {
    for y := 0; y < b.H; y++ {
        for x := 0; x < b.W; x++ {
            b.Cells[y][x] = NewCell(' ', *NewColor(0, 0, 0), *NewColor(127, 127, 127))
        }
    }
}

//TODO eventually use actual colors
func (b *Buffer) Set(x, y int, ch rune){
    b.Cells[y][x] = NewCell(ch, *NewColor(0,0,0),*NewColor(127,127,127))
}

func (b *Buffer) Get(x, y int) *Cell{
    return b.Cells[y][x]
}
