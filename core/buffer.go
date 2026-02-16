package core

import "github.com/dmsRosa6/glyph/geom"

type Buffer struct {
    W, H int
    cells [][]*Cell
    rootClip *geom.Bounds

    Bg Color
    Fg Color
}

func NewBuffer(w, h int, fg, bg Color) *Buffer {
    cells := make([][]*Cell, h)

    rootClip := geom.NewBounds(0,0,w-1,h-1)

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
        cells: cells,
        rootClip: rootClip,
    }
}

func (b *Buffer) ClearUsingDefaults() {
    for y := 0; y < b.H; y++ {
        for x := 0; x < b.W; x++ {
            b.cells[y][x] = NewCell(' ', b.Fg, b.Bg)
        }
    }
}

func (b *Buffer) Clear(fg, bg Color) {
    for y := 0; y < b.H; y++ {
        for x := 0; x < b.W; x++ {
            b.cells[y][x] = NewCell(' ', fg, bg)
        }
    }
}

func (b *Buffer) Set(x, y int, ch rune, bg, fg Color){
    if y >= b.H || x >= b.W {
        return
    }

    b.cells[y][x] = NewCell(ch, fg, bg)
}

func (b *Buffer) Get(x, y int) *Cell{
    return b.cells[y][x]
}


func (b *Buffer) GetCells() ([][]*Cell, int, int) {
    return b.cells, b.W,b.H
}
