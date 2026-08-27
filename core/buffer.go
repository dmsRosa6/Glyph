package core

type Buffer struct {
	W, H  int
	cells [][]*Cell

	Bg Color
	Fg Color

	clipStack []bufferClip
}

type bufferClip struct {
	x0, y0, x1, y1 int
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
		Fg:    fg,
		Bg:    bg,
		cells: cells,
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

func (b *Buffer) PushClip(x, y, w, h int) {
	nx0, ny0, nx1, ny1 := x, y, x+w, y+h

	if len(b.clipStack) > 0 {
		cur := b.clipStack[len(b.clipStack)-1]
		if cur.x0 > nx0 {
			nx0 = cur.x0
		}
		if cur.y0 > ny0 {
			ny0 = cur.y0
		}
		if cur.x1 < nx1 {
			nx1 = cur.x1
		}
		if cur.y1 < ny1 {
			ny1 = cur.y1
		}
	}

	b.clipStack = append(b.clipStack, bufferClip{x0: nx0, y0: ny0, x1: nx1, y1: ny1})
}

func (b *Buffer) PopClip() {
	if len(b.clipStack) == 0 {
		return
	}
	b.clipStack = b.clipStack[:len(b.clipStack)-1]
}

func (b *Buffer) Set(x, y int, ch rune, bg, fg Color) {
	if y >= b.H || x >= b.W || y < 0 || x < 0 {
		return
	}

	if len(b.clipStack) > 0 {
		c := b.clipStack[len(b.clipStack)-1]
		if x < c.x0 || x >= c.x1 || y < c.y0 || y >= c.y1 {
			return
		}
	}

	b.cells[y][x] = NewCell(ch, fg, bg)
}

func (b *Buffer) Get(x, y int) *Cell {
	return b.cells[y][x]
}

func (b *Buffer) GetCells() ([][]*Cell, int, int) {
	return b.cells, b.W, b.H
}
